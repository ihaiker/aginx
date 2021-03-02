package tcloud

import (
	"fmt"
	"github.com/ihaiker/aginx/v2/api"
	"github.com/ihaiker/aginx/v2/core/logs"
	"github.com/ihaiker/aginx/v2/core/util/errors"
	"github.com/ihaiker/aginx/v2/plugins/certificate"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/regions"
	ssl "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ssl/v20191205"
	"hash/crc32"
	"net/url"
	"path/filepath"
	"strings"
	"time"
)

var logger = logs.New("cert", "module", "tcloud")

type tencentPlugin struct {
	client     *ssl.Client
	aginx      api.Aginx
	storageDir string

	certs         []*tencentCertificate
	nextAsyncTime time.Time //下次同步时间
}

func LoadCertificate() *tencentPlugin {
	return &tencentPlugin{}
}

func (l *tencentPlugin) Scheme() string {
	return "tcloud"
}

func (t *tencentPlugin) Name() string {
	return "腾讯云SSL证书"
}

func (t *tencentPlugin) Version() string {
	return "v2.0.0"
}

func (t *tencentPlugin) Help() string {
	return `腾讯云申请免费证书
配置方式：tcloud://ssl.tencentcloudapi.com/<storage path>?secretId=&secretKey=;
参数：             说明
storage path    证书存储路径 默认： certs/tcloud
secretId        腾讯云 secretId.
secretKey       腾讯云 secretKey
region          腾讯云 region
`
}

func (t *tencentPlugin) dir(config url.URL) string {
	dir := config.Path
	if dir == "" || dir == "/" {
		dir = fmt.Sprintf("certs/%s", t.Scheme())
	}
	if strings.HasPrefix(dir, "/") {
		dir = dir[1:]
	}
	return dir
}

func (t *tencentPlugin) Initialize(config url.URL, aginx api.Aginx) (err error) {
	t.aginx = aginx
	t.storageDir = t.dir(config)
	t.certs = make([]*tencentCertificate, 0)

	region := config.Query().Get("region")
	if region == "" {
		region = regions.Beijing
	}
	credential := common.NewCredential(
		config.Query().Get("secretId"), config.Query().Get("secretKey"))
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = config.Host

	if t.client, err = ssl.NewClient(credential, region, cpf); err != nil {
		return
	}

	err = t.syncCerts()
	return
}

//apply 申请证书
func (t *tencentPlugin) apply(domain string) (string, error) {
	req := ssl.NewApplyCertificateRequest()
	req.DomainName = common.StringPtr(domain)
	req.DvAuthMethod = common.StringPtr("FILE")
	resp, err := t.client.ApplyCertificate(req)
	if err != nil {
		return "", err
	}
	logger.Debugf("申请腾讯证书：%s,%s", domain, *resp.Response.CertificateId)
	return *resp.Response.CertificateId, nil
}

//describe 获取证书详细信息
func (t *tencentPlugin) describe(certId string, cert *tencentCertificate) (err error) {
	req := ssl.NewDescribeCertificateDetailRequest()
	req.CertificateId = common.StringPtr(certId)
	var resp *ssl.DescribeCertificateDetailResponse
	if resp, err = t.client.DescribeCertificateDetail(req); err != nil {
		return
	}

	cert.certId = certId
	cert.domain = *resp.Response.Domain
	cert.insertTime, _ = time.Parse("2006-01-02 15:04:05", *resp.Response.InsertTime)

	if *resp.Response.Status == 0 /*待验证*/ || *resp.Response.Status == 11 /*重发待验证*/ {
		cert.status = NoVerified
		cert.loc = fmt.Sprintf("%s%s", *resp.Response.DvAuthDetail.DvAuthPath,
			*resp.Response.DvAuthDetail.DvAuthKey)
		cert.txt = *resp.Response.DvAuthDetail.DvAuthValue
	} else if *resp.Response.Status == 1 /*已经颁发*/ {
		cert.status = Normal
		cert.expiredTime, _ = time.Parse("2006-01-02 15:04:05", *resp.Response.CertEndTime)
		cert.certificate = *resp.Response.CertificatePublicKey
		cert.privateKey = *resp.Response.CertificatePrivateKey
	} else {
		cert.status = Expired //其他的全部归为过期处理
	}
	return
}

//complete 触发主动认证
func (t *tencentPlugin) complete(domain, certId string) error {
	logger.Debugf("配置完成，申请验证：%s,%s", domain, certId)
	req := ssl.NewCompleteCertificateRequest()
	req.CertificateId = common.StringPtr(certId)
	_, err := t.client.CompleteCertificate(req)
	return err
}

func (t *tencentPlugin) New(domain string) (certificate.Files, error) {
	var err error
	var cert *tencentCertificate

	//获取第一个未验证的，如果有就是用他去验证
	for _, c := range t.getCerts() {
		if c.domain == domain && c.status == NoVerified {
			cert = c
		}
	}

	//如果没有申请一个
	if cert == nil {
		cert = new(tencentCertificate)
		if cert.certId, err = t.apply(domain); err != nil {
			return nil, errors.Wrap(err, "申请证书")
		}
		if err = t.describe(cert.certId, cert); err != nil {
			return nil, errors.Wrap(err, "获取验证信息")
		}
		//即使下面逻辑没有成功，这里也需要保存一下啊，防止重新申请再次申请一个新的。
		t.certs = append(t.certs, cert)
	}

	logger.Debugf("认证信息：%s, %s,%s", cert.domain, cert.loc, cert.txt)

	nvp := newVerifiedProvider(t.aginx)
	//添加验证信息
	if err = nvp.present(domain, cert.loc, cert.txt); err != nil {
		return nil, errors.Wrap(err, "设置验证信息")
	}
	//可以忽略清楚直接保存信息
	defer func() { _ = nvp.cleanUp() }()

	//推送验证
	if err = t.complete(domain, cert.certId); err != nil {
		logger.Warn("推送验证信息错误：", err.Error())
		return nil, errors.Wrap(err, "推送验证信息")
	}

	logger.Debug("获取证书详细信息：", domain)
	//获取完整验证信息
	if err = t.describe(cert.certId, cert); err != nil {
		logger.Warn("申请后获取证书信息错误：", err)
		return nil, errors.Wrap(err, "获取验证信息")
	}

	err = t.storeCertFiles(cert)
	return cert, err
}

func (t *tencentPlugin) Get(domain string) (certificate.Files, error) {
	var cert *tencentCertificate

	for _, c := range t.getCerts() {
		if c.domain == domain && c.status == Normal {
			if cert == nil {
				cert = c
			} else if c.expiredTime.After(cert.expiredTime) {
				cert = c
			}
		}
	}

	if cert == nil {
		return nil, errors.New("not found: %s", domain)
	}
	return cert, nil
}

//syncCerts 同步证书
func (t *tencentPlugin) syncCerts() error {
	t.nextAsyncTime = time.Now().Add(time.Minute * 10)

	//获取全部列表
	request := ssl.NewDescribeCertificatesRequest()
	request.Limit = common.Uint64Ptr(500)
	resp, err := t.client.DescribeCertificates(request)
	if err != nil {
		return err
	}

LOOP:
	for _, ct := range resp.Response.Certificates {
		cert := new(tencentCertificate)
		cert.certId = *ct.CertificateId
		if err = t.describe(*ct.CertificateId, cert); err != nil {
			logger.Warnf("获取证书信息错误:", ct.Domain, " ", err)
			continue
		}

		if cert.status == Expired {
			continue //过期文件不做保存
		} else {

			for i, c := range t.certs {
				if c.certId == cert.certId {
					t.certs[i] = cert
					continue LOOP //之前已经刷过该证书
				} else if c.domain == cert.domain && cert.expiredTime.Before(c.expiredTime) {
					continue LOOP //相同域名证书,只保存最后时间的。
				}
			}

			t.certs = append(t.certs, cert)
			_ = t.storeCertFiles(cert)
		}
	}

	return nil
}

func (t tencentPlugin) fileDiff(path string, content []byte) bool {
	if f, err := t.aginx.Files().Get(path); err == nil {
		return crc32.ChecksumIEEE(f.Content) != crc32.ChecksumIEEE(content)
	}
	return true
}

//保存证书文件
func (t tencentPlugin) storeCertFiles(cert *tencentCertificate) error {
	cert.certificatePath = filepath.Join(t.storageDir, cert.domain, "server.crt")
	cert.privateKeyPath = filepath.Join(t.storageDir, cert.domain, "server.key")

	if cert.status == Normal {
		if t.fileDiff(cert.certificatePath, []byte(cert.certificate)) {
			logger.Info("存储证书文件：", cert.certificatePath)
			if err := t.aginx.Files().NewWithContent(cert.certificatePath, []byte(cert.certificate)); err != nil {
				return errors.Wrap(err, "存储证书")
			}
		}
		if t.fileDiff(cert.privateKeyPath, []byte(cert.privateKey)) {
			logger.Info("存储证书文件：", cert.privateKeyPath)
			if err := t.aginx.Files().NewWithContent(cert.privateKeyPath, []byte(cert.privateKey)); err != nil {
				return errors.Wrap(err, "存储证书私钥")
			}
		}
	}
	return nil
}

func (t *tencentPlugin) List() (map[string]certificate.Files, error) {
	rets := map[string]certificate.Files{}
	for _, cert := range t.getCerts() {
		if cert.status == Normal {
			rets[cert.domain] = cert
		}
	}
	return rets, nil
}

func (t *tencentPlugin) getCerts() []*tencentCertificate {
	if t.nextAsyncTime.Before(time.Now()) {
		if err := t.syncCerts(); err != nil {
			logger.Warn("同步证书异常：", err)
		}
	}
	return t.certs
}
