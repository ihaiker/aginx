package client

import (
	"fmt"
	"github.com/ihaiker/aginx/v2/api"
	"github.com/ihaiker/aginx/v2/core/certs"
	cfg "github.com/ihaiker/aginx/v2/core/config"
	"github.com/ihaiker/aginx/v2/core/logs"
	"github.com/ihaiker/aginx/v2/core/nginx"
	"github.com/ihaiker/aginx/v2/core/nginx/config"
	"github.com/ihaiker/aginx/v2/core/registry"
	storagePlugin "github.com/ihaiker/aginx/v2/core/storage"
	"github.com/ihaiker/aginx/v2/core/util/errors"
	"github.com/ihaiker/aginx/v2/plugins/certificate"
	"github.com/ihaiker/aginx/v2/plugins/storage"
	"net/url"
	"strings"
)

var logger = logs.New("client")

type client struct {
	engine  storage.Plugin
	daemon  nginx.Daemon
	certs   map[string]certificate.Plugin
	certDef string
}

func (c *client) addProvider(certConfigs []string) error {
	for _, cert := range certConfigs {
		if certConfig, err := url.Parse(cert); err != nil {
			return errors.Wrap(err, "错误配置："+cert)
		} else if certPlugin, has := certs.Plugins[certConfig.Scheme]; !has {
			return fmt.Errorf("未发现cert插件: %s", cert)
		} else if err = certPlugin.Initialize(*certConfig, c); err != nil {
			return fmt.Errorf("初始化 %s", certConfig.Scheme)
		} else {
			c.certs[certConfig.Scheme] = certPlugin
		}
	}
	return nil
}

func New(engine storage.Plugin, daemon nginx.Daemon, certConfigs []string, certDef string) (*client, error) {
	c := &client{
		engine: engine, daemon: daemon,
		certs: map[string]certificate.Plugin{}, certDef: certDef,
	}
	if err := c.addProvider(certConfigs); err != nil {
		return nil, err
	}
	if certDef != "" {
		if _, has := c.certs[certDef]; !has {
			return nil, fmt.Errorf("未发现cert默认配置: %s in (%s)", certDef, strings.Join(certConfigs, ","))
		}
	}
	return c, nil
}

func (c *client) Configuration() (*config.Configuration, error) {
	return nginx.Configuration(c.engine)
}

func (c *client) Files() api.File {
	return &clientFile{engine: c.engine, daemon: c.daemon}
}

func (c *client) Directive() api.Directive {
	return &clientDirective{engine: c.engine, daemon: c.daemon}
}

func (c *client) Certs() api.Certs {
	return &clientCert{
		engine: c.engine, daemon: c.daemon,
		certs: c.certs, certDef: c.certDef,
	}
}

func (c *client) Backup() api.Backup {
	return &clientBackup{
		engine: c.engine, daemon: c.daemon,
		dir:      cfg.Config.Backup.Dir,
		limit:    cfg.Config.Backup.Limit,
		dayLimit: cfg.Config.Backup.DayLimit,
	}
}

func (c *client) Info() (map[string]map[string]string, error) {
	pluginInfo := map[string]map[string]string{}

	pluginInfo["certificate"] = map[string]string{}
	for _, plugin := range certs.Plugins {
		pluginInfo["certificate"][plugin.Scheme()] = plugin.Name()
	}

	pluginInfo["storage"] = map[string]string{}
	for _, plugin := range storagePlugin.Plugins {
		pluginInfo["storage"][plugin.Scheme()] = plugin.Name()
	}

	pluginInfo["registry"] = map[string]string{}
	for _, plugin := range registry.Plugins {
		pluginInfo["registry"][plugin.Scheme()] = plugin.Name()
	}
	return pluginInfo, nil
}
