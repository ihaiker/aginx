package http

import (
	"errors"
	"github.com/ihaiker/aginx/nginx"
	ngx "github.com/ihaiker/aginx/nginx/config"
	"github.com/ihaiker/aginx/util"
	"github.com/kataras/iris/v12"
)

type directiveController struct {
	process *nginx.Process
}

func (as *directiveController) queryDirective(client *nginx.Client, queries []string) []*ngx.Directive {
	directives, err := client.Select(queries...)
	util.PanicIfError(err)
	return directives
}

func (as *directiveController) addDirective(client *nginx.Client, queries []string, directives []*ngx.Directive) int {
	util.PanicIfError(client.Add(queries, directives...))
	util.PanicIfError(as.process.Test(client.Configuration()))
	util.PanicIfError(client.Store())
	return as.reload()
}

func (as *directiveController) deleteDirective(client *nginx.Client, queries []string) int {
	util.PanicIfError(client.Delete(queries...))
	util.PanicIfError(as.process.Test(client.Configuration()))
	util.PanicIfError(client.Store())
	return as.reload()
}

func (as *directiveController) modifyDirective(client *nginx.Client, queries []string, directives []*ngx.Directive) int {
	if len(directives) == 0 {
		panic(errors.New("new directive is empty"))
	}
	util.PanicIfError(client.Modify(queries, directives[0]))
	util.PanicIfError(as.process.Test(client.Configuration()))
	util.PanicIfError(client.Store())
	return as.reload()
}

func (as *directiveController) reload() int {
	util.PanicIfError(as.process.Reload())
	return iris.StatusNoContent
}
