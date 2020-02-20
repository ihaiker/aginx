package client

import (
	"errors"
	"fmt"
	"github.com/ihaiker/aginx/nginx/configuration"
	queryLexer "github.com/ihaiker/aginx/nginx/query"
	"github.com/ihaiker/aginx/storage"
	"os"
)

var (
	ErrNotFound            = os.ErrNotExist
	ErrRootCannotBeDeleted = errors.New("root cannot be deleted")
)

func Queries(query ...string) []string {
	return query
}

type Client struct {
	doc    *configuration.Configuration
	Engine storage.Engine
}

func NewClient(store storage.Engine) (*Client, error) {
	doc, err := Readable(store)
	if err != nil {
		return nil, err
	}
	return &Client{doc: doc, Engine: store}, nil
}

func (client *Client) find(directives []*configuration.Directive, query string) ([]*configuration.Directive, error) {
	expr, err := queryLexer.Parser(query)
	if err != nil {
		return nil, fmt.Errorf("Search condition error：[%s]", query)
	}
	matched := make([]*configuration.Directive, 0)
	for _, directive := range directives {
		for _, body := range directive.Body {
			if expr.Match(body) {
				matched = append(matched, body)
			}
		}
	}
	return matched, nil
}

func (client Client) Configuration() *configuration.Configuration {
	return client.doc
}

func (client *Client) Select(queries ...string) ([]*configuration.Directive, error) {
	current := []*configuration.Directive{client.doc.Directive()}
	for _, query := range queries {
		directives, err := client.find(current, query)
		if err != nil {
			return nil, err
		}
		if directives == nil || len(directives) == 0 {
			return nil, ErrNotFound
		}
		current = directives
	}
	return current, nil
}

func (client *Client) Add(queries []string, addDirectives ...*configuration.Directive) error {
	if directives, err := client.Select(queries...); err == ErrNotFound {
		return err
	} else {
		for _, directive := range directives {
			directive.Modify = true
			directive.Body = append(directive.Body, addDirectives...)
		}
		return nil
	}
}

func (client *Client) Delete(queries ...string) error {
	if len(queries) == 0 {
		return ErrRootCannotBeDeleted
	}
	finder := queries[0 : len(queries)-1]
	directives, err := client.Select(finder...)
	if err != nil {
		return err
	}

	deleteQuery := queries[len(queries)-1]
	expr, err := queryLexer.Parser(deleteQuery)
	if err != nil {
		return err
	}

	err = ErrNotFound
	for _, directive := range directives {

		deleteDirectiveIdx := make([]int, 0)
		for i, body := range directive.Body {
			if expr.Match(body) {
				deleteDirectiveIdx = append(deleteDirectiveIdx, i)
			}
		}
		if len(deleteDirectiveIdx) > 0 {
			directive.Modify = true
			err = nil
		}

		for i := len(deleteDirectiveIdx) - 1; i >= 0; i-- {
			idx := deleteDirectiveIdx[i]
			directive.Body = append(directive.Body[:idx], directive.Body[idx+1:]...)
		}
	}
	return err
}

func (client *Client) Modify(queries []string, directive *configuration.Directive) error {
	selectDirectives, err := client.Select(queries...)
	if err != nil {
		return err
	}
	for _, selectDirective := range selectDirectives {
		selectDirective.Modify = true
		selectDirective.Name = directive.Name
		selectDirective.Args = directive.Args
		selectDirective.Body = directive.Body
	}
	return nil
}
