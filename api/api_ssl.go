package api

import (
	"fmt"
	"github.com/ihaiker/aginx/lego"
	"net/http"
	"net/url"
	"time"
)

type aginxSSL struct {
	*client
}

func (self *aginxSSL) New(accountEmail, domain string) (sf *lego.StoreFile, err error) {
	uri := fmt.Sprintf("/ssl/%s?email=%s", domain, url.QueryEscape(accountEmail))
	sf = new(lego.StoreFile)
	err = self.request(http.MethodPut, uri, nil, sf, self.timeout(time.Second*7))
	return
}

func (a aginxSSL) ReNew(domain string) (sf *lego.StoreFile, err error) {
	uri := fmt.Sprintf("/ssl/%s", domain)
	sf = new(lego.StoreFile)
	err = a.request(http.MethodPut, uri, nil, sf, a.timeout(time.Second*7))
	return
}
