package api

import (
	"bytes"
	"errors"
	"github.com/ihaiker/aginx/nginx/configuration"
	"net/http"
)

type aginxDirective struct {
	*client
}

func (self *aginxDirective) HttpUpstream(names ...string) (directives []*configuration.Directive, err error) {
	directives = make([]*configuration.Directive, 0)
	err = self.request(http.MethodGet, self.get("/http/upstream", names), nil, &directives)
	return
}

func (self *aginxDirective) HttpServer(names ...string) (directives []*configuration.Directive, err error) {
	directives = make([]*configuration.Directive, 0)
	err = self.request(http.MethodGet, self.get("/http/server", names), nil, &directives)
	return
}

func (self *aginxDirective) StreamUpstream(names ...string) (directives []*configuration.Directive, err error) {
	directives = make([]*configuration.Directive, 0)
	err = self.request(http.MethodGet, self.get("/stream/upstream", names), nil, &directives)
	return
}

func (self *aginxDirective) StreamServer(listens ...string) (directives []*configuration.Directive, err error) {
	directives = make([]*configuration.Directive, 0)
	err = self.request(http.MethodGet, self.get("/stream/server", listens), nil, &directives)
	return
}

func (self *aginxDirective) Select(queries ...string) (directives []*configuration.Directive, err error) {
	directives = make([]*configuration.Directive, 0)
	err = self.request(http.MethodGet, self.get("/api", queries), nil, &directives)
	return
}

func (self *aginxDirective) Add(queries []string, addDirectives ...*configuration.Directive) error {
	if len(addDirectives) == 0 {
		return errors.New("addDirectives is empty")
	}
	body := bytes.NewBufferString("")
	for _, directive := range addDirectives {
		body.WriteString(directive.Pretty(0))
		body.WriteString("\n")
	}
	return self.request(http.MethodPut, self.get("/api", queries), body, nil)
}

func (self *aginxDirective) Delete(queries ...string) error {
	return self.request(http.MethodDelete, self.get("/api", queries), nil, nil)
}

func (self *aginxDirective) Modify(queries []string, directive *configuration.Directive) error {
	body := bytes.NewBufferString("")
	body.WriteString(directive.Pretty(0))
	return self.request(http.MethodPost, self.get("/api", queries), body, nil)
}
