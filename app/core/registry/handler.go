package registry

import (
	"bytes"
	"fmt"
	"github.com/ihaiker/aginx/v2/api"
	"github.com/ihaiker/aginx/v2/core/logs"
	"github.com/ihaiker/aginx/v2/core/nginx/config"
	"github.com/ihaiker/aginx/v2/core/registry/consul"
	"github.com/ihaiker/aginx/v2/core/registry/docker"
	"github.com/ihaiker/aginx/v2/core/registry/functions"
	"github.com/ihaiker/aginx/v2/core/util/errors"
	"github.com/ihaiker/aginx/v2/core/util/services"
	"github.com/ihaiker/aginx/v2/plugins/registry"
	"net/url"
	"text/template"
)

var logger = logs.New("registry")

var Plugins = map[string]registry.Plugin{
	"consul": consul.LoadRegistry(),
	"docker": docker.LoadRegistry(),
}
var Functions = map[string]interface{}{}

type EventHandler struct {
	aginx   api.Aginx
	plugins []registry.Plugin
}

type templateData struct {
	UpstreamName string
	Domain       string
	AutoSSL      bool
	Cert         *api.CertFile
	Servers      []api.UpstreamServer
}

func Handler(aginx api.Aginx) (b *EventHandler) {
	b = &EventHandler{
		aginx: aginx, plugins: make([]registry.Plugin, 0),
	}
	return
}

func (b *EventHandler) Add(path string) (reg registry.Plugin, err error) {
	var c *url.URL
	if c, err = url.Parse(path); err != nil {
		return
	}
	for name, p := range Plugins {
		if name == c.Scheme {
			reg = p
			logger.Infof("启用registry %s", name)
			err = reg.Watch(*c, b.aginx)
			return
		}
	}
	err = fmt.Errorf("未知 registry ：%s", path)
	return
}

//安全获取upstream
func (b *EventHandler) getUpstream(domain string) (up *api.Upstream) {
	upstreamName := functions.UpstreamName(domain)
	upstreams, err := b.aginx.GetUpstream(&api.Filter{
		Name: upstreamName, Protocol: api.ProtocolHTTP, ExactMatch: true,
	})
	if err != nil {
		logger.WithError(err).Warn("获取upstream错误")
	}
	if upstreams == nil || len(upstreams) == 0 {
		up = new(api.Upstream)
		up.Servers = make([]api.UpstreamServer, 0)
	} else {
		up = upstreams[0]
	}
	return
}

func (b *EventHandler) findTemplate(event registry.Domain) string {
	domain := event.Domain
	paths := []string{"templates/" + domain + ".tpl", "templates/default.tpl"}
	//指定模板
	if event.Template != "" {
		paths = append([]string{"templates/" + event.Template + ".tpl"}, paths...)
	}
	for _, path := range paths {
		if f, err := b.aginx.Files().Get(path); err == nil {
			return string(f.Content)
		} else {
			logger.Debugf("%s 模板文件 %s 未找到", domain, path)
		}
	}
	return default_template
}

func (b *EventHandler) genConfigByTemplate(plugin registry.Plugin, event registry.Domain, servers map[string]api.UpstreamServer) ([]byte, error) {
	domain := event.Domain
	autoSSL := event.AutoSSL
	configTemplate := b.findTemplate(event)
	funcs := functions.Merge(Functions, plugin.TemplateFuncMap())
	out := bytes.NewBufferString("")
	t, err := template.New("").Funcs(funcs).Parse(configTemplate)
	if err != nil {
		return nil, err
	}
	data := templateData{
		UpstreamName: functions.UpstreamName(domain),
		Domain:       domain, AutoSSL: autoSSL,
	}
	for _, server := range servers {
		data.Servers = append(data.Servers, server)
	}
	if autoSSL {
		if data.Cert, err = b.aginx.Certs().Get(domain); errors.IsNotFound(err) {
			provider := event.Provider
			if data.Cert, err = b.aginx.Certs().New(provider, domain); err != nil {
				return nil, err
			}
		}
	}
	if err := t.Execute(out, data); err != nil {
		return nil, err
	}
	//判断输出的文件是否正确，并且会把文件格式话一下
	if conf, err := config.ParseWith(domain, out.Bytes(), nil); err != nil {
		return nil, err
	} else {
		return conf.BodyBytes(), nil
	}
}

func (b *EventHandler) handleEvent(plugin registry.Plugin, event registry.Domain) {
	upstream := b.getUpstream(event.Domain)

	addresses := map[string]api.UpstreamServer{}
	for _, server := range upstream.Servers {
		addresses[server.String()] = server
	}
	if event.Alive {
		logger.Infof("%s %s 上线", event.Domain, event.Address)
		for _, address := range event.Address {
			if _, has := addresses[address]; !has {
				addresses[address] = api.UpstreamServer{
					HostAndPort: api.ParseHostAndPort(address),
					Weight:      event.Weight,
				}
			}
		}
	} else {
		logger.Infof("%s %s 下线", event.Domain, event.Address)
		for _, address := range event.Address {
			delete(addresses, address)
		}
	}

	if _, err := b.aginx.Directive().Select(
		"http", fmt.Sprintf("include('%d.d/*.conf')", plugin.Scheme()),
	); errors.IsNotFound(err) {
		if err = b.aginx.Directive().Add(
			[]string{"http"}, config.New("include", plugin.Scheme()+".d/*.conf")); err != nil {
			logger.WithError(err).Warnf("添加include错误 %s", plugin.Scheme())
		}
	}

	path := fmt.Sprintf("%s.d/%s.conf", plugin.Scheme(), event.Domain)
	if len(addresses) == 0 { //删除
		logger.Debugf("删除注册配置(%s,%s)", plugin.Scheme(), event.Domain)
		if err := b.aginx.Files().Remove(path); err != nil {
			logger.WithError(err).Warnf("删除注册 %s %s", plugin.Scheme(), event.Domain)
		}
	} else { //重新生成
		if ctx, err := b.genConfigByTemplate(plugin, event, addresses); err != nil {
			logger.WithError(err).Warnf("模板生成配置(%s,%s)", plugin.Scheme(), event.Domain)
		} else if err = b.aginx.Files().NewWithContent(path, ctx); err != nil {
			logger.WithError(err).Warn("上传文件错误")
		} else {
			logger.Debugf("重新生成(%s,%s)", plugin.Scheme(), event.Domain)
		}
	}
}

func (b *EventHandler) Start() error {
	for _, pg := range b.plugins {
		go func(p registry.Plugin) {
			eventsChan := p.Label()
			for {
				select {
				case events, ok := <-eventsChan:
					if ok {
						for _, event := range events {
							b.handleEvent(p, event)
						}
					}
				}
			}
		}(pg)

		if err := services.Start(pg); err != nil {
			return err
		}
		logger.Debug("start registry ", pg.Scheme())
	}
	return nil
}

//关闭存储的连接等操作
func (b *EventHandler) Stop() error {
	for _, pg := range b.plugins {
		logger.Debug("close registry ", pg.Scheme())
		_ = services.Stop(pg)
	}
	return nil
}
