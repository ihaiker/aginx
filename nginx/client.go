package nginx

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound            = errors.New("not found")
	ErrRootCannotBeDeleted = errors.New("root cannot be deleted")
)

type Client interface {

	//查询指令
	Select(queries ...string) ([]*Directive, error)

	//添加指令
	Add(queries []string, directives ...*Directive) error

	//删除指令
	Delete(queries ...string) error

	//更新指令
	Modify(queries []string, directive *Directive) error

	Configuration() *Configuration
}

func Queries(query ...string) []string {
	return query
}

type client struct {
	doc *Configuration
}

func NewClient(doc *Configuration) Client {
	return &client{doc: doc}
}

func (client *client) find(directives []*Directive, query string) ([]*Directive, error) {
	expr, err := Parser(query)
	if err != nil {
		return nil, fmt.Errorf("Search condition error：[%s]", query)
	}
	matched := make([]*Directive, 0)
	for _, directive := range directives {
		for _, body := range directive.Body {
			if expr.Match(body) {
				matched = append(matched, body)
			}
		}
	}
	return matched, nil
}

func (client client) Configuration() *Configuration {
	return client.doc
}

func (client *client) Select(queries ...string) ([]*Directive, error) {
	current := []*Directive{client.doc.Directive()}
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

func (client *client) Add(queries []string, addDirectives ...*Directive) error {
	if directives, err := client.Select(queries...); err == ErrNotFound {
		return err
	} else {
		for _, directive := range directives {
			directive.Body = append(directive.Body, addDirectives...)
		}
		return nil
	}
}

func (client *client) Delete(queries ...string) error {
	if len(queries) == 0 {
		return ErrRootCannotBeDeleted
	}
	finder := queries[0 : len(queries)-1]
	directives, err := client.Select(finder...)
	if err != nil {
		return err
	}

	deleteQuery := queries[len(queries)-1]
	expr, err := Parser(deleteQuery)
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
			err = nil
		}

		for i := len(deleteDirectiveIdx) - 1; i >= 0; i-- {
			idx := deleteDirectiveIdx[i]
			directive.Body = append(directive.Body[:idx], directive.Body[idx+1:]...)
		}
	}
	return err
}

func (client *client) Modify(queries []string, directive *Directive) error {
	selectDirectives, err := client.Select(queries...)
	if err != nil {
		return err
	}
	for _, selectDirective := range selectDirectives {
		selectDirective.Name = directive.Name
		selectDirective.Args = directive.Args
		selectDirective.Body = directive.Body
	}
	return nil
}
