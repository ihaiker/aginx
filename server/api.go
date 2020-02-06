package server

import (
	"bytes"
	"errors"
	"github.com/ihaiker/aginx/nginx"
	"github.com/ihaiker/aginx/util"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"github.com/kataras/iris/v12/hero"
	"github.com/kataras/iris/v12/middleware/basicauth"
	"strings"
	"time"
)

func Routers(vister *Supervister, auth string) func(*iris.Application) {
	handlers := make([]context.Handler, 0)
	if auth != "" {
		authConfig := strings.SplitN(auth, ":", 2)
		handlers = append(handlers, basicauth.New(basicauth.Config{
			Users: map[string]string{authConfig[0]: authConfig[1]},
			Realm: "Authorization Required", Expires: time.Duration(30) * time.Minute,
		}))
	}

	h := hero.New()
	h.Register(func(ctx iris.Context) []string {
		return ctx.Request().URL.Query()["q"]
	})
	h.Register(func(ctx iris.Context) nginx.Client {
		doc, err := nginx.AnalysisNginx()
		util.PanicIfError(err)
		return nginx.NewClient(doc)
	})
	h.Register(func(ctx iris.Context) []*nginx.Directive {
		body, err := ctx.GetBody()
		util.PanicIfError(err)
		conf, err := nginx.Analysis("", nginx.NamedReader(bytes.NewBuffer(body), ""))
		util.PanicIfError(err)
		return conf.Directive().Body
	})
	service := &apiService{vister: vister}
	return func(app *iris.Application) {
		api := app.Party("/api", handlers...)
		{
			api.Get("", h.Handler(service.queryDirective))
			api.Put("", h.Handler(service.addDirective))
			api.Delete("", h.Handler(service.deleteDirective))
			api.Post("", h.Handler(service.modifyDirective))
		}
	}
}

type apiService struct {
	vister *Supervister
}

func (as *apiService) queryDirective(client nginx.Client, queries []string) []*nginx.Directive {
	directives, err := client.Select(queries...)
	util.PanicIfError(err)
	return directives
}

func (as *apiService) addDirective(client nginx.Client, queries []string, directives []*nginx.Directive) int {
	util.PanicIfError(client.Add(queries, directives...))
	util.PanicIfError(as.vister.Test(client.Configuration()))
	return iris.StatusNoContent
}

func (as *apiService) deleteDirective(client nginx.Client, queries []string) int {
	util.PanicIfError(client.Delete(queries...))
	util.PanicIfError(as.vister.Test(client.Configuration()))
	return iris.StatusNoContent
}

func (as *apiService) modifyDirective(client nginx.Client, queries []string, directives []*nginx.Directive) int {
	if len(directives) == 0 {
		panic(errors.New("new directive is empty"))
	}
	util.PanicIfError(client.Modify(queries, directives[0]))
	util.PanicIfError(as.vister.Test(client.Configuration()))
	return iris.StatusNoContent
}
