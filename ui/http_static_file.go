// +build !bindata

package ui

import (
	"github.com/kataras/iris/v12"
)

func Static(app *iris.Application) {
	app.Favicon("ui/dist/favicon.ico")
	app.HandleDir("/static", "ui/dist/static")
	app.RegisterView(iris.HTML("ui/dist", ".html"))
	app.Get("/ui", func(ctx iris.Context) {
		_ = ctx.View("index.html")
	})
}
