package api

import (
	"bytes"
	"errors"
	nginx "github.com/ihaiker/aginx/nginx/config"
	"net/http"
)

type aginxDirective struct {
	*client
}

func (self *aginxDirective) Select(queries ...string) (directives []*nginx.Directive, err error) {
	directives = make([]*nginx.Directive, 0)
	err = self.request(http.MethodGet, self.get("/api", queries), nil, &directives)
	return
}

func (self *aginxDirective) Add(queries []string, addDirectives ...*nginx.Directive) error {
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

func (self *aginxDirective) Modify(queries []string, directive *nginx.Directive) error {
	body := bytes.NewBufferString("")
	body.WriteString(directive.Pretty(0))
	return self.request(http.MethodPost, self.get("/api", queries), body, nil)
}
