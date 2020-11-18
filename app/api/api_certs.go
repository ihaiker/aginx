package api

import (
	"net/http"
	"net/url"
	"time"
)

type httpApiCerts struct {
	*client
}

func (self *httpApiCerts) New(provider, domain string) (sf *CertFile, err error) {
	uri := "/api/cert?domain=" + url.QueryEscape(domain) + "&provider=" + url.QueryEscape(provider)
	sf = new(CertFile)
	err = self.request(http.MethodPost, uri, nil, sf, self.timeout(time.Second*15))
	return
}

func (self *httpApiCerts) Get(domain string) (sf *CertFile, err error) {
	uri := "/api/cert?domain=" + url.QueryEscape(domain)
	sf = new(CertFile)
	err = self.request(http.MethodGet, uri, nil, sf, self.timeout(time.Second*15))
	return
}

func (self *httpApiCerts) List() (sfs []*CertFile, err error) {
	uri := "/api/cert/list"
	sfs = make([]*CertFile, 0)
	err = self.request(http.MethodGet, uri, nil, &sfs, self.timeout(time.Second*15))
	return
}
