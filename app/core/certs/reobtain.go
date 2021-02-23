package certs

import (
	"github.com/ihaiker/aginx/v2/api"
	"github.com/ihaiker/aginx/v2/core/certs/custom"
	"github.com/ihaiker/aginx/v2/core/certs/lego"
	"github.com/ihaiker/aginx/v2/core/certs/tcloud"
	"github.com/ihaiker/aginx/v2/core/logs"
	"github.com/ihaiker/aginx/v2/core/util/errors"
	"github.com/ihaiker/aginx/v2/plugins/certificate"
	"time"
)

var logger = logs.New("renewal")

var Plugins = map[string]certificate.Plugin{
	"lego":   lego.LoadCertificates(),
	"tcloud": tcloud.LoadCertificate(),
	"custom": custom.LoadCertificate(),
}

type certReObtain struct {
	aginx   api.Aginx
	renewal func(file *api.CertFile) error
	timer   *time.Timer
}

func ReObtain(aginx api.Aginx, renewal func(file *api.CertFile) error) *certReObtain {
	return &certReObtain{
		aginx: aginx, renewal: renewal,
	}
}

func (c *certReObtain) Check() {
	logger.Debug("检查证书过期")
	if certs, err := c.aginx.Certs().List(); err != nil {
		logger.WithError(err).Warn("检查证书出错")
	} else {
		for name, cert := range certs {
			if time.Now().Add(time.Hour * 24 * 15).After(cert.ExpireTime) {
				if cert.Provider != "custom" {
					logger.Infof("自定义域名证书过期：%s，请及时处理！！", cert.Domain)
					continue
				}
				logger.Infof("域名证书过期续租：%s", cert.Domain)
				errors.Try(func() {
					if e := c.renewal(cert); e != nil {
						logger.WithError(e).Warnf("域名续租失败: %s", name)
					}
				})
			}
		}
	}
}
