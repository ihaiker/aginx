package nginx

import (
	"fmt"
	"github.com/ihaiker/aginx/nginx/config"
	"github.com/ihaiker/aginx/nginx/query"
	"github.com/ihaiker/aginx/util"
)

func find(directives []*config.Directive, q string) ([]*config.Directive, error) {
	expr, err := query.Lexer(q)
	if err != nil {
		return nil, fmt.Errorf("Search condition error：[%s]", q)
	}
	matched := make([]*config.Directive, 0)
	for _, directive := range directives {
		for _, body := range directive.Body {
			if expr.Match(body) {
				matched = append(matched, body)
			}
		}
	}
	return matched, nil
}

func Select(conf *config.Configuration, queries ...string) ([]*config.Directive, error) {
	current := []*config.Directive{conf}
	for _, q := range queries {
		directives, err := find(current, q)
		if err != nil {
			return nil, err
		}
		if directives == nil || len(directives) == 0 {
			return nil, util.ErrNotFound
		}
		current = directives
	}
	return current, nil
}

func MustSelect(conf *config.Configuration, queries ...string) []*config.Directive {
	directives, err := Select(conf, queries...)
	util.PanicIfError(err)
	return directives
}
