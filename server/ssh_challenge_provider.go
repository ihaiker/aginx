package server

import (
	"fmt"
	"github.com/go-acme/lego/v3/challenge/http01"
	"github.com/ihaiker/aginx/nginx/client"
	"github.com/ihaiker/aginx/nginx/configuration"
	"github.com/ihaiker/aginx/storage/file"
	"path/filepath"
	"time"
)

type aginxProvider struct {
	queries   []string
	directive *configuration.Directive
	api       *client.Client
	vistor    *Supervister
}

func NewAginxProvider(api *client.Client, vister *Supervister) *aginxProvider {
	return &aginxProvider{api: api, vistor: vister}
}

func (self *aginxProvider) selectDirective(domain string) (queries []string, directive *configuration.Directive) {
	serverQuery := fmt.Sprintf("server.[server_name('%s') & listen('80')]", domain)
	queries = client.Queries("http", "include", "*", serverQuery)
	if directives, err := self.api.Select(queries...); err == nil {
		directive = directives[0]
		return
	}
	queries = client.Queries("http", serverQuery)
	if directives, err := self.api.Select(queries...); err == nil {
		directive = directives[0]
		return
	}
	return
}

func (self *aginxProvider) location(token, keyAuth string) *configuration.Directive {
	location := configuration.NewDirective("location", http01.ChallengePath(token))
	location.AddBody("add_header", "Content-Type", `"text/plain"`)
	location.AddBody("return", "200", fmt.Sprintf("'%s'", keyAuth))
	return location
}

func (self *aginxProvider) server(domain, token, keyAuth string) *configuration.Directive {
	server := configuration.NewDirective("server")
	server.AddBody("listen", "80")
	server.AddBody("server_name", domain)
	server.AddBodyDirective(self.location(token, keyAuth))
	return server
}

func (self *aginxProvider) reload() error {
	_, conf, _ := file.GetInfo()
	if err := configuration.Down(filepath.Dir(conf), self.api.Configuration()); err != nil {
		return err
	}
	return self.vistor.Reload()
}

func (self *aginxProvider) Present(domain, token, keyAuth string) error {
	self.queries, self.directive = self.selectDirective(domain)

	if self.directive != nil {
		location := self.location(token, keyAuth)
		if err := self.api.Add(self.queries, location); err != nil {
			return err
		}
	} else {
		server := self.server(domain, token, keyAuth)
		if err := self.api.Add(client.Queries("http"), server); err != nil {
			return err
		}
	}

	return self.reload()
}

func (self *aginxProvider) CleanUp(domain, token, keyAuth string) error {

	if self.directive != nil {
		location := configuration.NewDirective("location", http01.ChallengePath(token))
		_ = self.api.Delete(append(self.queries, location.Query())...)
	} else {
		_ = self.api.Delete(self.queries...)
	}

	return self.reload()
}

func (self *aginxProvider) Timeout() (timeout, interval time.Duration) {
	return time.Minute, time.Second * 3
}
