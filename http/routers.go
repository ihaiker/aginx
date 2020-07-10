package http

import (
	"encoding/base64"
	"fmt"
	"github.com/ihaiker/aginx/lego"
	"github.com/ihaiker/aginx/logs"
	"github.com/ihaiker/aginx/nginx"
	ngx "github.com/ihaiker/aginx/nginx/config"
	"github.com/ihaiker/aginx/plugins"
	"github.com/ihaiker/aginx/ui"
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
	authConfig := strings.SplitN(auth, ":", 2)
	handlers = append(handlers, basicauth.New(basicauth.Config{
		Users: map[string]string{authConfig[0]: authConfig[1]},
		Realm: "Authorization Required", Expires: time.Duration(30) * time.Minute,
	}))

	h := hero.New()
	h.Register(
		func(ctx iris.Context) []string {
			return ctx.Request().URL.Query()["q"]
		},
		func(ctx iris.Context) *nginx.Client {
			return nginx.MustClient(email, engine, manager, process)
		},
		func(ctx iris.Context) *nginx.Service {
			return nginx.NewService(nginx.MustClient(email, engine, manager, process))
		},
		func(ctx iris.Context) []*ngx.Directive {
			body, err := ctx.GetBody()
			util.PanicIfError(err)
			conf, err := nginx.ReaderReadable(engine, plugins.NewFile("", body))
			util.PanicIfError(err)
			return conf.Body
		},
	)

	fileCtrl := &fileController{engine: engine, process: process}
	directive := &directiveController{process: process}
	serverCtrl := &serverController{process: process}
	ssl := &sslController{email: email}

	manager.Expire(func(domain string) {
		ssl.Renew(nginx.MustClient(email, engine, manager, process), domain)
	})

	return func(app *iris.Application) {
		app.Post("/login", h.Handler(func(ctx context.Context) map[string]string {
			data := map[string]string{}
			util.PanicIfError(ctx.ReadJSON(&data))
			userAuth := fmt.Sprintf("%s:%s", data["user"], data["passwd"])
			util.AssertTrue(userAuth == auth, "wrong user name or password!")
			token := "Basic " + base64.StdEncoding.EncodeToString([]byte(userAuth))
			return map[string]string{"token": token}
		}))

		ui.Static(app)

		api := app.Party("/api", handlers...)
		{
			api.Get("", h.Handler(directive.queryDirective))
			api.Put("", h.Handler(directive.addDirective))
			api.Delete("", h.Handler(directive.deleteDirective))
			api.Post("", h.Handler(directive.modifyDirective))
		}

		server := app.Party("/server", handlers...)
		{
			server.Get("", h.Handler(serverCtrl.listServer))
			server.Delete("", h.Handler(serverCtrl.deleteServer))
			server.Post("", h.Handler(serverCtrl.postServer))
		}

		limit := iris.LimitRequestBodySize(1024 * 1024 * 10)
		file := app.Party("/file", handlers...)
		{
			file.Post("", limit, h.Handler(fileCtrl.New))
			file.Post("/ctx", limit, h.Handler(fileCtrl.NewFileContent))
			file.Delete("", h.Handler(fileCtrl.Remove))
			file.Get("", h.Handler(fileCtrl.Search))
		}

		sslRouter := app.Party("/ssl", handlers...)
		{
			sslRouter.Put("/{domain:string}", h.Handler(ssl.New))
			sslRouter.Post("/{domain:string}", h.Handler(ssl.Renew))
		}

		app.Any("/reload", h.Handler(directive.reload))
	}
}
