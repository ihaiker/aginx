package client

import (
	"fmt"
	"github.com/ihaiker/aginx/v2/api"
	"github.com/ihaiker/aginx/v2/core/nginx"
	"github.com/ihaiker/aginx/v2/core/util/errors"
	"github.com/ihaiker/aginx/v2/plugins/certificate"
	"github.com/ihaiker/aginx/v2/plugins/storage"
)

type clientCert struct {
	engine  storage.Plugin
	daemon  nginx.Daemon
	certs   map[string]certificate.Plugin
	certDef string
}

func (c *clientCert) toFile(plugin certificate.Plugin, domain string, file certificate.Files) *api.CertFile {
	cert := new(api.CertFile)
	cert.Provider = plugin.Scheme()
	cert.Domain = domain
	cert.ExpireTime = file.GetExpiredTime()
	cert.Certificate = file.Certificate()
	cert.PrivateKey = file.PrivateKey()
	return cert
}

func (c *clientCert) New(provider, domain string) (*api.CertFile, error) {
	cert, err := c.Get(domain)
	if provider == "" {
		provider = c.certDef
	}
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	plugin, has := c.certs[provider]
	if !has {
		return nil, errors.New("not found provider %s", provider)
	}
	file, err := plugin.New(domain)
	if err != nil {
		return nil, err
	}
	cert = c.toFile(plugin, domain, file)
	return cert, err
}

func (c *clientCert) Get(domain string) (*api.CertFile, error) {
	for _, plugin := range c.certs {
		file, err := plugin.Get(domain)
		if errors.IsNotFound(err) {
			continue
		} else if err != nil {
			logger.WithError(err).Warnf("从 %s 获取域名 %s", plugin.Scheme(), domain)
			continue
		}
		cert := c.toFile(plugin, domain, file)
		return cert, nil
	}
	return nil, fmt.Errorf("not found: %s", domain)
}

func (c *clientCert) List() ([]*api.CertFile, error) {
	certFiles := make([]*api.CertFile, 0)
	for _, plugin := range c.certs {
		items, err := plugin.List()
		if err != nil {
			logger.WithError(err).Warnf("获取证书列表：%s", plugin.Scheme())
			continue
		}
		for domain, item := range items {
			file := c.toFile(plugin, domain, item)
			certFiles = append(certFiles, file)
		}
	}
	return certFiles, nil
}
