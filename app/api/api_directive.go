package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/ihaiker/aginx/v2/core/nginx/config"
	"net/http"
)

type httpAginxDirective struct {
	*client
}

func (self *httpAginxDirective) Batch(batch []*DirectiveBatch) error {
	body := bytes.NewBufferString("")
	if err := json.NewEncoder(body).Encode(batch); err != nil {
		return err
	}
	return self.request(http.MethodPost, self.get("/api/directive/batch", []string{}), body, nil)
}

func (self *httpAginxDirective) Select(queries ...string) (directives []*config.Directive, err error) {
	directives = make([]*config.Directive, 0)
	err = self.request(http.MethodGet, self.get("/api/directive", queries), nil, &directives)
	return
}

func (self *httpAginxDirective) Add(queries []string, addDirectives ...*config.Directive) error {
	if len(addDirectives) == 0 {
		return errors.New("addDirectives is empty")
	}
	body := bytes.NewBufferString("")
	for _, directive := range addDirectives {
		body.WriteString(directive.Pretty(0))
		body.WriteString("\n")
	}
	return self.request(http.MethodPut, self.get("/api/directive", queries), body, nil)
}

func (self *httpAginxDirective) Delete(queries ...string) error {
	return self.request(http.MethodDelete, self.get("/api/directive", queries), nil, nil)
}

func (self *httpAginxDirective) Modify(queries []string, directive *config.Directive) error {
	body := bytes.NewBufferString("")
	body.WriteString(directive.Pretty(0))
	return self.request(http.MethodPost, self.get("/api/directive", queries), body, nil)
}
