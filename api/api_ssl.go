package api

import (
	"fmt"
	"github.com/ihaiker/aginx/lego"
	"net/http"
	"net/url"
)

type aginxSSL struct {
	*client
}

func (self *aginxSSL) New(accountEmail, domain string) (sf *lego.StoreFile, err error) {
	sf = new(lego.StoreFile)
	err = self.request(http.MethodPut, fmt.Sprintf("/ssl/%s?email=%s", domain, url.QueryEscape(accountEmail)), nil, sf)
	return
}

func (a aginxSSL) ReNew(domain string) (sf *lego.StoreFile, err error) {
	sf = new(lego.StoreFile)
	err = a.request(http.MethodPut, fmt.Sprintf("/ssl/%s"+domain), nil, sf)
	return
}
