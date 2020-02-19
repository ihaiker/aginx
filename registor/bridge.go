package registor

import (
	"github.com/ihaiker/aginx/api"
	aginx "github.com/ihaiker/aginx/nginx/client"
	"github.com/ihaiker/aginx/nginx/configuration"
)

type RegistorBridge struct {
	Registrator Registrator
	TemplateDir string
	Api         api.Aginx
}

func (rb *RegistorBridge) listenChange() error {
	go func() {
		for {
			select {
			case event := <-rb.Registrator.Listener():
				if event.EventType == Online {

				}
			}
		}
	}()
	return nil
}

func (rb *RegistorBridge) createRegDInclude() error {
	if _, err := rb.Api.Select("http", "include('reg.d/*.ngx.conf')"); err != nil {
		if err != aginx.ErrNotFound {
			return err
		}
		if err = rb.Api.Add(api.Queries("http"),
			configuration.NewDirective("include", "reg.d/*.ngx.conf")); err != nil {
			return err
		}
	}
	return nil
}

func (rb *RegistorBridge) domainIncludeFile(domain string) *configuration.Configuration {
	return nil
}

func (rb *RegistorBridge) Start() error {
	if err := rb.createRegDInclude(); err != nil {
		return err
	}

	//services := rb.Registrator.Sync().Group()
	//for domain, servers := range services {
	//	//templateFile, err := rb.Api.Select(filepath.Join(rb.TemplateDir, domain+".ngx.conf"))
	//	//nginxFile := rb.domainIncludeFile(domain)
	//}

	if err := rb.Registrator.Start(); err != nil {
		return err
	}
	return rb.listenChange()
}

func (rb *RegistorBridge) Stop() error {
	if rb.Registrator != nil {
		return rb.Stop()
	}
	return nil
}
