package http

import (
	"github.com/ihaiker/aginx/v2/api"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/hero"
)

func Routers(aginx api.Aginx) func(*iris.Application) {
	h := hero.New()

	fileCtl := &fileController{aginx: aginx}
	dirCtl := &directiveController{aginx: aginx}
	sslCtl := &sslController{aginx: aginx}
	ausCtl := &serverAndUpstreamController{aginx: aginx}
	backupCtl := &backupController{aginx: aginx}

	return func(app *iris.Application) {
		api := app.Party("/api")
		{
			api.Get("/info", h.Handler(func() map[string]map[string]string {
				info, _ := aginx.Info()
				return info
			}))

			file := api.Party("/file")
			{
				file.Post("", h.Handler(fileCtl.New))          //上传一个文件
				file.Get("", h.Handler(fileCtl.Get))           //搜索一个文件
				file.Delete("", h.Handler(fileCtl.Remove))     //删除一个文件
				file.Get("/search", h.Handler(fileCtl.Search)) //搜索一个文件
			}
			directive := api.Party("/directive")
			{
				directive.Get("", h.Handler(dirCtl.Select))
				directive.Put("", h.Handler(dirCtl.Add))
				directive.Delete("", h.Handler(dirCtl.Delete))
				directive.Post("", h.Handler(dirCtl.Modify))
				directive.Post("/batch", h.Handler(dirCtl.Batch))
			}
			ssl := api.Party("/cert")
			{
				ssl.Get("", h.Handler(sslCtl.Get))
				ssl.Post("", h.Handler(sslCtl.New))
				ssl.Get("/list", h.Handler(sslCtl.List))
			}

			server := api.Party("/server")
			{
				server.Get("", h.Handler(ausCtl.GetServer))
				server.Post("", h.Handler(ausCtl.SetServer))
			}
			upstream := api.Party("/upstream")
			{
				upstream.Get("", h.Handler(ausCtl.GetUpstream))
				upstream.Post("", h.Handler(ausCtl.SetUpstream))
			}

			backup := api.Party("/backup")
			{
				backup.Get("", h.Handler(backupCtl.list))
				backup.Delete("", h.Handler(backupCtl.delete))
				backup.Post("", h.Handler(backupCtl.backup))
				backup.Put("", h.Handler(backupCtl.rollback))
			}
		}
	}
}
