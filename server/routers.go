package server

import (
	"bytes"
	"github.com/ihaiker/aginx/lego"
	"github.com/ihaiker/aginx/nginx/client"
	"github.com/ihaiker/aginx/nginx/configuration"
	"github.com/ihaiker/aginx/storage"
	"github.com/ihaiker/aginx/util"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"github.com/kataras/iris/v12/hero"
	"github.com/kataras/iris/v12/middleware/basicauth"
	"strings"
	"sync"
	"time"
)

func Routers(vister *Supervister, engine storage.Engine, manager *lego.Manager, auth string) func(*iris.Application) {
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
	}, func(ctx iris.Context) *client.Client {
		api, err := client.NewClient(engine)
		util.PanicIfError(err)
		return api
	}, func(ctx iris.Context) []*configuration.Directive {
		body, err := ctx.GetBody()
		util.PanicIfError(err)
		conf, err := client.ReaderReadable(engine, util.NamedReader(bytes.NewBuffer(body), ""))
		util.PanicIfError(err)
		return conf.Directive().Body
	})

	fileCtrl := &fileController{engine: engine}
	directive := &directiveController{vister: vister, manager: manager, engine: engine}
	ssl := &sslController{vister: vister, manager: manager, engine: engine, lock: new(sync.Mutex)}
	_ = util.EBus.Subscribe(util.SSLExpire, ssl.Expire)

	return func(app *iris.Application) {
		api := app.Party("/api", handlers...)
		{
			api.Get("", h.Handler(directive.queryDirective))
			api.Put("", h.Handler(directive.addDirective))
			api.Delete("", h.Handler(directive.deleteDirective))
			api.Post("", h.Handler(directive.modifyDirective))
		}
		for _, f := range []string{"http", "stream"} {
			extendApi := app.Party("/"+f, handlers...)
			{
				for _, s := range []string{"server", "upstream"} {
					extendApi.Get("/"+s, h.Handler(directive.selectDirective(
						f+","+s,
						f+",include,*,"+s,
					)))
				}
			}
		}

		//
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
