// +build !bindata

package admin

import (
	"github.com/kataras/iris/v12"
)

func Static(app *iris.Application) {
	app.Favicon("web/dist/favicon.ico")
	app.HandleDir("/static", "web/dist/static")
	app.RegisterView(iris.HTML("web/dist", ".html"))
	app.Get("/console", func(ctx iris.Context) {
		_ = ctx.View("index.html")
	})
}
