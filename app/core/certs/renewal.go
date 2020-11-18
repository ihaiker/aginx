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
	"TCloud": tcloud.LoadCertificate(),
	"custom": custom.LoadCertificate(),
}

type certRenewal struct {
	aginx   api.Aginx
	renewal func(file *api.CertFile) error
	timer   *time.Timer
}

func Renewal(aginx api.Aginx, renewal func(file *api.CertFile) error) *certRenewal {
	return &certRenewal{
		aginx: aginx, renewal: renewal,
	}
}

func (c *certRenewal) Start() error {
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
	c.nextCheck()
	return nil
}

func (c *certRenewal) nextCheck() {
	//凌晨2点检查续租问题
	next := time.Now().Add(time.Hour * 24).Truncate(time.Hour * 24).Add(-time.Hour * 6)
	t := next.Sub(time.Now())
	c.timer = time.AfterFunc(t, func() {
		if err := c.Start(); err != nil {
			logger.WithError(err).Warn("tls续租问题")
		}
	})
}

func (c *certRenewal) Stop() error {
	if c.timer != nil {
		c.timer.Stop()
	}
	return nil
}
