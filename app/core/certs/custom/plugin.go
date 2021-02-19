package custom

import (
	"crypto/x509"
	"encoding/pem"
	"github.com/ihaiker/aginx/v2/api"
	"github.com/ihaiker/aginx/v2/core/util"
	"github.com/ihaiker/aginx/v2/core/util/errors"
	"github.com/ihaiker/aginx/v2/plugins/certificate"
	"net/url"
	"path/filepath"
	"time"
)

type (
	customPlugin struct {
		path  string
		aginx api.Aginx
	}
	customFile struct {
		expireTime  time.Time
		certificate string
		privateKey  string
	}
)

func LoadCertificate() *customPlugin {
	return &customPlugin{}
}

func (c *customPlugin) Scheme() string {
	return "custom"
}

func (c *customPlugin) Name() string {
	return "自定义证书"
}

func (c *customPlugin) Version() string {
	return "v2.0.0"
}

func (c *customPlugin) Help() string {
	return "用户自定义证书，只是存放证书，不支持续租，用户需要自己通过file自己管理上传。并且文件存储在<storage>上，\n" +
		"可以设置: custom://<storage path> 设置存储路径，默认为：certs/custom。\n" +
		"文件的存放方式为：<storage path>/<domain>/server.key, <storage path>/<domain>/server.crt 例如：certs/custom/api.aginx.io/server.key"
}

func (c *customPlugin) Initialize(config url.URL, aginx api.Aginx) error {
	c.path = filepath.Join(config.Host, config.Path)
	c.aginx = aginx
	_, err := c.List()
	return err
}

func (c *customPlugin) New(domain string) (certificate.Files, error) {
	return c.Get(domain)
}

func (c *customPlugin) Get(domain string) (certificate.Files, error) {
	if files, err := c.List(); err != nil {
		return nil, err
	} else if cert, has := files[domain]; has {
		return cert, nil
	} else {
		return nil, errors.New("not fount %s", domain)
	}
}

func (c *customPlugin) List() (map[string]certificate.Files, error) {
	if files, err := c.aginx.Files().Search(c.path+"/*/server.crt", c.path+"/*/server.key"); err != nil {
		return nil, err
	} else {
		certs := map[string]certificate.Files{}
		for _, file := range files {
			domain, fileName := util.Split2(file.Name[len(c.path)+1:], "/")
			cert, has := certs[domain]
			if !has {
				cert = new(customFile)
			}
			if fileName == "server.key" {
				cert.(*customFile).privateKey = file.Name
			} else {
				cert.(*customFile).certificate = file.Name
				{
					block, _ := pem.Decode([]byte(file.Content))
					certInfo, _ := x509.ParseCertificate(block.Bytes)
					cert.(*customFile).expireTime = certInfo.NotAfter
				}
			}
			certs[domain] = cert
		}
		return certs, nil
	}
}

func (c customFile) GetExpiredTime() time.Time {
	return c.expireTime
}

func (c customFile) Certificate() string {
	return c.certificate
}

func (c customFile) PrivateKey() string {
	return c.privateKey
}
