package api

import (
	"crypto/tls"
	"github.com/ihaiker/aginx/v2/core/nginx/config"
	"net/http"
	"net/url"
	"time"
)

type httpAginx struct {
	address string
	client  *http.Client
}
type baseAuthTransport struct {
	name     string
	password string
	*http.Transport
}

func (bat *baseAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if bat.name != "" {
		req.SetBasicAuth(bat.name, bat.password)
	}
	return bat.Transport.RoundTrip(req)
}

func New(address, name, password string) *httpAginx {
	apiUrl, _ := url.Parse(address)
	tp := &baseAuthTransport{
		name: name, password: password,
		Transport: &http.Transport{},
	}
	if apiUrl.Scheme == "https" {
		tp.Transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	client := &http.Client{
		Transport: tp,
		Timeout:   time.Second * 3,
	}
	return NewWithClient(address, client)
}

func NewWithClient(address string, client *http.Client) *httpAginx {
	return &httpAginx{address: address, client: client}
}

func (a *httpAginx) http() *client {
	return &client{
		address: a.address, client: a.client,
	}
}

func (a *httpAginx) Configuration() (*config.Configuration, error) {
	if directives, err := a.Directive().Select(); err != nil {
		return nil, err
	} else {
		conf := config.New("nginx.conf")
		conf.AddBodyDirective(directives...)
		return conf, nil
	}
}

func (a *httpAginx) Files() File {
	return &httpAginxFile{client: a.http()}
}

func (a *httpAginx) Directive() Directive {
	return &httpAginxDirective{client: a.http()}
}

func (a *httpAginx) Certs() Certs {
	return &httpApiCerts{client: a.http()}
}

func (a *httpAginx) Info() (map[string]map[string]string, error) {
	plugins := map[string]map[string]string{}
	err := a.http().request(http.MethodGet, "/api/info", nil, plugins)
	return plugins, err
}
