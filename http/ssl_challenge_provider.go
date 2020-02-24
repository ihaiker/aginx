package http

import (
	"fmt"
	"github.com/go-acme/lego/v3/challenge/http01"
	"github.com/ihaiker/aginx/nginx"
	"time"
)

type sslProvider struct {
	queries   []string
	directive *nginx.Directive
	api       *nginx.Client
	process   *nginx.Process
}

func NewAginxProvider(api *nginx.Client, process *nginx.Process) *sslProvider {
	return &sslProvider{api: api, process: process}
}

func (self *sslProvider) selectDirective(domain string) (queries []string, directive *nginx.Directive) {
	serverQuery := fmt.Sprintf("server.[server_name('%s') & listen('80')]", domain)
	queries = nginx.Queries("http", "include", "*", serverQuery)
	if directives, err := self.api.Select(queries...); err == nil {
		directive = directives[0]
		return
	}
	queries = nginx.Queries("http", serverQuery)
	if directives, err := self.api.Select(queries...); err == nil {
		directive = directives[0]
		return
	}
	return
}

func (self *sslProvider) location(token, keyAuth string) *nginx.Directive {
	location := nginx.NewDirective("location", http01.ChallengePath(token))
	location.AddBody("add_header", "Content-Type", `"text/plain"`)
	location.AddBody("return", "200", fmt.Sprintf("'%s'", keyAuth))
	return location
}

func (self *sslProvider) server(domain, token, keyAuth string) *nginx.Directive {
	server := nginx.NewDirective("server")
	server.AddBody("listen", "80")
	server.AddBody("server_name", domain)
	server.AddBodyDirective(self.location(token, keyAuth))
	return server
}

func (self *sslProvider) storeAndReload() error {
	if err := self.api.Store(); err != nil {
		return err
	}
	return self.process.Reload()
}

func (self *sslProvider) Present(domain, token, keyAuth string) error {
	self.queries, self.directive = self.selectDirective(domain)

	if self.directive != nil {
		location := self.location(token, keyAuth)
		if err := self.api.Add(self.queries, location); err != nil {
			return err
		}
	} else {
		server := self.server(domain, token, keyAuth)
		if err := self.api.Add(nginx.Queries("http"), server); err != nil {
			return err
		}
	}

	return self.storeAndReload()
}

func (self *sslProvider) CleanUp(domain, token, keyAuth string) error {
	if self.directive != nil {
		locationQuery := fmt.Sprintf("location.('%s')", http01.ChallengePath(token))
		_ = self.api.Delete(append(self.queries, locationQuery)...)
	} else {
		_ = self.api.Delete(self.queries...)
	}
	return self.storeAndReload()
}

func (self *sslProvider) Timeout() (timeout, interval time.Duration) {
	return time.Minute, time.Second * 3
}
