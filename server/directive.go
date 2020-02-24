package server

import (
	"errors"
	"github.com/ihaiker/aginx/lego"
	"github.com/ihaiker/aginx/nginx"
	"github.com/ihaiker/aginx/nginx/client"
	"github.com/ihaiker/aginx/nginx/daemon"
	"github.com/ihaiker/aginx/storage"
	"github.com/ihaiker/aginx/util"
	"github.com/kataras/iris/v12"
	"strings"
)

type directiveController struct {
	vister  *daemon.Supervister
	engine  storage.Engine
	manager *lego.Manager
}

func (as *directiveController) queryDirective(client *client.Client, queries []string) []*nginx.Directive {
	directives, err := client.Select(queries...)
	util.PanicIfError(err)
	return directives
}

func (as *directiveController) addDirective(client *client.Client, queries []string, directives []*nginx.Directive) int {
	util.PanicIfError(client.Add(queries, directives...))
	util.PanicIfError(as.vister.Test(client.Configuration()))
	util.PanicIfError(client.Store())

	if !as.engine.IsCluster() {
		_ = as.vister.Reload()
	}
	return iris.StatusNoContent
}

func (as *directiveController) deleteDirective(client *client.Client, queries []string) int {
	util.PanicIfError(client.Delete(queries...))
	util.PanicIfError(as.vister.Test(client.Configuration()))
	util.PanicIfError(client.Store())
	if !as.engine.IsCluster() {
		_ = as.vister.Reload()
	}
	return iris.StatusNoContent
}

func (as *directiveController) modifyDirective(client *client.Client, queries []string, directives []*nginx.Directive) int {
	if len(directives) == 0 {
		panic(errors.New("new directive is empty"))
	}
	util.PanicIfError(client.Modify(queries, directives[0]))
	util.PanicIfError(as.vister.Test(client.Configuration()))
	util.PanicIfError(client.Store())
	if !as.engine.IsCluster() {
		_ = as.vister.Reload()
	}
	return iris.StatusNoContent
}

func (as *directiveController) reload() int {
	util.PanicIfError(as.vister.Reload())
	return iris.StatusNoContent
}

func (as *directiveController) selectDirective(queries ...string) func(*client.Client) []*nginx.Directive {
	return func(client *client.Client) []*nginx.Directive {
		directives := make([]*nginx.Directive, 0)
		for _, query := range queries {
			if ds, err := client.Select(strings.Split(query, ",")...); err == nil {
				directives = append(directives, ds...)
			}
		}
		return directives
	}
}
