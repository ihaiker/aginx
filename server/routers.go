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
	"time"
)

func Routers(sv *Supervister, st storage.Engine, manager *lego.Manager, auth string) func(*iris.Application) {
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
		api, err := client.NewClient(st)
		util.PanicIfError(err)
		return api
	}, func(ctx iris.Context) []*configuration.Directive {
		body, err := ctx.GetBody()
		util.PanicIfError(err)
		conf, err := client.ReaderReadable(st, util.NamedReader(bytes.NewBuffer(body), ""))
		util.PanicIfError(err)
		return conf.Directive().Body
	})

	ctrl := &apiController{vister: sv, manager: manager, engine: st}
	return func(app *iris.Application) {
		api := app.Party("/api", handlers...)
		{
			api.Get("", h.Handler(ctrl.queryDirective))
			api.Put("", h.Handler(ctrl.addDirective))
			api.Delete("", h.Handler(ctrl.deleteDirective))
			api.Post("", h.Handler(ctrl.modifyDirective))
		}
		for _, f := range []string{"http", "stream"} {
			extendApi := app.Party("/"+f, handlers...)
			{
				for _, s := range []string{"server", "upstream"} {
					extendApi.Get("/"+s, h.Handler(ctrl.selectDirective(
						f+","+s,
						f+",include,*,"+s,
					)))
				}
			}
		}
		limit := iris.LimitRequestBodySize(1024 * 1024 * 10)
		app.Post("/upload", limit, h.Handler(ctrl.upload))
		app.Any("/reload", h.Handler(ctrl.reload))
	}
}
