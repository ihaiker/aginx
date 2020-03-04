package http

import (
	"fmt"
	"github.com/ihaiker/aginx/lego"
	"github.com/ihaiker/aginx/logs"
	"github.com/ihaiker/aginx/nginx"
	"github.com/ihaiker/aginx/plugins"
	"github.com/ihaiker/aginx/util"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"github.com/kataras/iris/v12/hero"
	"github.com/kataras/iris/v12/middleware/basicauth"
	"strings"
	"time"
)

var logger = logs.New("http")

func Routers(email, auth string, process *nginx.Process, engine plugins.StorageEngine, manager *lego.Manager) func(*iris.Application) {
	handlers := make([]context.Handler, 0)
	if auth != "" {
		authConfig := strings.SplitN(auth, ":", 2)
		handlers = append(handlers, basicauth.New(basicauth.Config{
			Users: map[string]string{authConfig[0]: authConfig[1]},
			Realm: "Authorization Required", Expires: time.Duration(30) * time.Minute,
		}))
	}

	h := hero.New()
	h.Register(
		func(ctx iris.Context) []string {
			return ctx.Request().URL.Query()["q"]
		},
		func(ctx iris.Context) *nginx.Client {
			return nginx.MustClient(email, engine, manager, process)
		},
		func(ctx iris.Context) []*nginx.Directive {
			body, err := ctx.GetBody()
			util.PanicIfError(err)
			conf, err := nginx.ReaderReadable(engine, plugins.NewFile("", body))
			util.PanicIfError(err)
			return conf.Body
		},
	)

	fileCtrl := &fileController{engine: engine, process: process}
	directive := &directiveController{process: process}
	ssl := &sslController{email: email}
	simpleCtl := &simpleController{}

	manager.Expire(func(domain string) {
		ssl.Renew(nginx.MustClient(email, engine, manager, process), domain)
	})

	return func(app *iris.Application) {
		api := app.Party("/api", handlers...)
		{
			api.Get("", h.Handler(directive.queryDirective))
			api.Put("", h.Handler(directive.addDirective))
			api.Delete("", h.Handler(directive.deleteDirective))
			api.Post("", h.Handler(directive.modifyDirective))
		}

		simple := app.Party("/simple", handlers...)
		{
			for _, directiveNameTop := range []string{"http", "stream"} {
				for _, directiveNameSub := range []string{"server", "upstream"} {
					simple.Get(fmt.Sprintf("/%s/%s", directiveNameTop, directiveNameSub),
						h.Handler(simpleCtl.selectDirective(
							[]string{directiveNameTop, directiveNameSub},
							[]string{directiveNameTop, "include", "*", directiveNameSub},
						)),
					)
				}
			}
			simple.Put("/server", h.Handler(simpleCtl.newSimpleServer))
		}

		limit := iris.LimitRequestBodySize(1024 * 1024 * 10)
		app.Post("/file", limit, h.Handler(fileCtrl.New))
		app.Delete("/file", h.Handler(fileCtrl.Remove))
		app.Get("/file", h.Handler(fileCtrl.Search))

		sslRouter := app.Party("/ssl", handlers...)
		{
			sslRouter.Put("/{domain:string}", h.Handler(ssl.New))
			sslRouter.Post("/{domain:string}", h.Handler(ssl.Renew))
		}

		app.Any("/reload", h.Handler(directive.reload))
	}
}
