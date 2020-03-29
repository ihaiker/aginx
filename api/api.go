package api

import (
	"crypto/tls"
	ngx "github.com/ihaiker/aginx/nginx/config"
	"net/http"
	"net/url"
	"time"
)

type aginx struct {
	*client
}

func New(address string, maker ...func(client *http.Client)) *aginx {
	apiUrl, _ := url.Parse(address)
	tp := &BaseAuthTransport{
		Transport: &http.Transport{},
	}
	if apiUrl.Scheme == "https" {
		tp.Transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	httpClient := &http.Client{Transport: tp}
	httpClient.Timeout = time.Second * 3
	for _, f := range maker {
		f(httpClient)
	}
	return &aginx{
		client: &client{httpClient: httpClient, address: address},
	}
}

func (self *aginx) Auth(name, password string) {
	if tp, match := self.httpClient.Transport.(*BaseAuthTransport); match {
		tp.Name = name
		tp.Password = password
	}
}

func (self *aginx) Configuration() (*ngx.Configuration, error) {
	if directives, err := self.Directive().Select(); err != nil {
		return nil, err
	} else {
		return (*ngx.Configuration)(directives[0]), nil
	}
}

func (self *aginx) Reload() error {
	return self.request(http.MethodGet, "/reload", nil, nil)
}

func (self *aginx) File() AginxFile {
	return &aginxFile{client: self.client}
}

func (self *aginx) Directive() AginxDirective {
	return &aginxDirective{client: self.client}
}

func (self *aginx) SSL() AginxSSL {
	return &aginxSSL{client: self.client}
}

func (self *aginx) Simple() AginxSimple {
	return &aginxSimple{client: self.client}
}
