package api

import (
	"bytes"
	"encoding/json"
	ngx "github.com/ihaiker/aginx/nginx/config"
	"net/http"
)

type aginxSimple struct {
	*client
}

func (self *aginxSimple) HttpUpstream(names ...string) (directives []*ngx.Directive, err error) {
	directives = make([]*ngx.Directive, 0)
	err = self.request(http.MethodGet, self.get("/simple/http/upstream", names), nil, &directives)
	return
}

func (self *aginxSimple) HttpServer(names ...string) (directives []*ngx.Directive, err error) {
	directives = make([]*ngx.Directive, 0)
	err = self.request(http.MethodGet, self.get("/simple/http/server", names), nil, &directives)
	return
}

func (self *aginxSimple) StreamUpstream(names ...string) (directives []*ngx.Directive, err error) {
	directives = make([]*ngx.Directive, 0)
	err = self.request(http.MethodGet, self.get("/simple/stream/upstream", names), nil, &directives)
	return
}

func (self *aginxSimple) StreamServer(listens ...string) (directives []*ngx.Directive, err error) {
	directives = make([]*ngx.Directive, 0)
	err = self.request(http.MethodGet, self.get("/simple/stream/server", listens), nil, &directives)
	return
}

func (self *aginxSimple) SimpleServer(domain string, ssl bool, addresses []string) error {
	data := map[string]interface{}{
		"domain": domain, "ssl": ssl, "addresses": addresses,
	}
	bs, _ := json.Marshal(data)
	return self.request(http.MethodPut, "/simple/server", bytes.NewBuffer(bs), nil)
}
