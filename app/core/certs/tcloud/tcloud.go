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
	"net/url"
	"path/filepath"
	"reflect"
	"strings"
	"time"
)

var logger = logs.New("cert", "module", "tcloud")

type tencentPlugin struct {
	client     *ssl.Client
	aginx      api.Aginx
	storageDir string
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
		dir = fmt.Sprintf("certs/%s", t.Name())
	}
	if strings.HasPrefix(dir, "/") {
		dir = dir[1:]
	}
	return dir
}

func (t *tencentPlugin) Initialize(config url.URL, aginx api.Aginx) (err error) {
	t.aginx = aginx
	t.storageDir = t.dir(config)

	region := config.Query().Get("region")
	if region == "" {
		region = regions.Beijing
	}
	credential := common.NewCredential(
		config.Query().Get("secretId"), config.Query().Get("secretKey"))
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = config.Host

	t.client, err = ssl.NewClient(credential, region, cpf)
	_, err = t.refresh()
	return
}

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

func (t *tencentPlugin) describe(certId string) (cert *tencentCertificate, err error) {
	req := ssl.NewDescribeCertificateDetailRequest()
	req.CertificateId = common.StringPtr(certId)
	var resp *ssl.DescribeCertificateDetailResponse
	if resp, err = t.client.DescribeCertificateDetail(req); err != nil {
		return
	}
	cert = new(tencentCertificate)
	cert.certId = certId
	cert.domain = *resp.Response.Domain
	cert.insertTime, _ = time.Parse("2006-01-02 15:04:05", *resp.Response.InsertTime)

	if *resp.Response.Status == 0 /*待验证*/ || *resp.Response.Status == 11 /*重发待验证*/ {
		cert.status = NoVerified
		cert.loc = fmt.Sprintf("%s%s", *resp.Response.DvAuthDetail.DvAuthPath,
			*resp.Response.DvAuthDetail.DvAuthKey)
		cert.txt = *resp.Response.DvAuthDetail.DvAuthValue
		logger.Debugf("认证信息：%s, %s,%s", cert.domain, cert.loc, cert.txt)
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

func (t *tencentPlugin) complete(domain, certId string) error {
	logger.Debugf("配置完成，申请验证：%s,%s", domain, certId)
	req := ssl.NewCompleteCertificateRequest()
	req.CertificateId = common.StringPtr(certId)
	_, err := t.client.CompleteCertificate(req)
	return err
}

func (t *tencentPlugin) New(domain string) (certificate.Files, error) {
	certs, err := t.refresh()
	if err != nil {
		return nil, errors.Wrap(err, "刷新列表")
	}

	var cert *tencentCertificate
	for _, c := range certs {
		//获取第一个未验证的，如果有就是用他去验证
		if c.domain == domain && c.status == NoVerified {
			cert = c
			break
		}
	}

	nvp := newVerifiedProvider(t.aginx)
	//如果没有申请一个
	if cert == nil {
		certId, err := t.apply(domain)
		if err != nil {
			return nil, errors.Wrap(err, "申请证书")
		}
		if cert, err = t.describe(certId); err != nil {
			return nil, errors.Wrap(err, "获取验证信息")
		}
	}

	//添加验证信息
	if err = nvp.present(domain, cert.loc, cert.txt); err != nil {
		return nil, errors.Wrap(err, "设置验证信息")
	}

	//推送验证
	if err = t.complete(domain, cert.certId); err != nil {
		return nil, err
	}
	//可以忽略清楚直接保存信息
	_ = nvp.cleanUp()

	//获取完整验证信息
	if cert, err = t.describe(cert.certId); err != nil {
		return nil, errors.Wrap(err, "获取验证信息")
	}

	cert.certificatePath = filepath.Join(t.storageDir, domain, "server.crt")
	cert.privateKeyPath = filepath.Join(t.storageDir, domain, "server.key")

	if err = t.aginx.Files().New(cert.certificatePath, cert.certificate); err != nil {
		return nil, errors.Wrap(err, "存储证书")
	}
	if err = t.aginx.Files().New(cert.privateKeyPath, cert.privateKey); err != nil {
		return nil, errors.Wrap(err, "存储证书私钥")
	}
	return cert, nil
}

func (t *tencentPlugin) Get(domain string) (certificate.Files, error) {
	certs, err := t.List()
	if err != nil {
		return nil, err
	}
	cert, has := certs[domain]
	if !has {
		return nil, fmt.Errorf("not found: %s", domain)
	}
	return cert, nil
}

func (t *tencentPlugin) refresh() ([]*tencentCertificate, error) {
	//获取全部列表
	request := ssl.NewDescribeCertificatesRequest()
	request.Limit = common.Uint64Ptr(500)
	resp, err := t.client.DescribeCertificates(request)
	if err != nil {
		return nil, err
	}
	certs := make([]*tencentCertificate, 0)
	for _, ct := range resp.Response.Certificates {
		cert, _ := t.describe(*ct.CertificateId)
		if cert == nil || reflect.ValueOf(cert).IsZero() {
			continue
		}
		cert.certificatePath = filepath.Join(t.storageDir, cert.domain, "server.crt")
		cert.privateKeyPath = filepath.Join(t.storageDir, cert.domain, "server.key")

		certs = append(certs, cert)
	}
	return certs, nil
}

func (t *tencentPlugin) List() (map[string]certificate.Files, error) {
	rets := map[string]certificate.Files{}
	certs, err := t.refresh()
	if err != nil {
		return nil, err
	}
	for _, cert := range certs {
		if cert.status == Normal {
			/*默认倒叙排列，最后的一定是最新的*/
			if _, has := rets[cert.domain]; !has {
				rets[cert.domain] = cert
			}
		}
	}
	return rets, nil
}
