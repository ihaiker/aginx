package http

import (
	"github.com/ihaiker/aginx/nginx"
	"github.com/ihaiker/aginx/nginx/config"
	"github.com/ihaiker/aginx/util"
	"github.com/kataras/iris/v12"
)

type simpleController struct {
}

func (simple *simpleController) selectDirective(queries ...[]string) interface{} {
	return func(client *nginx.Client) []*config.Directive {
		directives := make([]*config.Directive, 0)
		for _, query := range queries {
			if ds, err := client.Select(query...); err == nil {
				directives = append(directives, ds...)
			}
		}
		return directives
	}
}

type simpleService struct {
	Domain    string   `json:"domain"`
	SSL       bool     `json:"ssl"`
	Addresses []string `json:"addresses"`
}

func (simple *simpleController) newSimpleServer(ctx iris.Context, client *nginx.Client) int {
	ss := new(simpleService)
	util.PanicIfError(ctx.ReadJSON(ss))
	util.AssertTrue(ss.Domain != "", "the domain is empty")
	util.AssertTrue(ss.Addresses != nil && len(ss.Addresses) > 0, "the proxy address is empty")
	util.PanicIfError(client.SimpleServer(ss.Domain, ss.SSL, ss.Addresses...))
	util.PanicIfError(client.Store())
	return iris.StatusNoContent
}
