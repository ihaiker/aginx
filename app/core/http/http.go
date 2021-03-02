package http

import (
	"context"
	"github.com/ihaiker/aginx/v2/core/config"
	"github.com/ihaiker/aginx/v2/core/logs"
	"github.com/ihaiker/aginx/v2/core/util/errors"
	"github.com/kataras/iris/v12"
	"path/filepath"
	"time"
)

var logger = logs.New("http")

type httpServer struct {
	app     *iris.Application
	address string
	routers []func(app *iris.Application)
}

func New(address string, routers ...func(*iris.Application)) *httpServer {
	return &httpServer{
		app: iris.New(), address: address, routers: routers,
	}
}

func (this *httpServer) Start() error {
	this.app.Use(recoverHandler())

	//错误消息
	this.app.OnErrorCode(iris.StatusNotFound, func(ctx iris.Context) {
		_, _ = ctx.JSON(map[string]string{
			"error": "notfound", "message": "the page not found!",
			"url": ctx.Request().RequestURI,
		})
	})
	this.app.Get("/health", func(ctx iris.Context) {
		_, _ = ctx.JSON(map[string]string{
			"status": "UP", "version": "v2.0",
		})
	})

	if len(config.Config.AllowIp) > 0 {
		this.app.UseGlobal(func(ctx iris.Context) {
			addr := ctx.RemoteAddr()
			for _, allow := range config.Config.AllowIp {
				if match, _ := filepath.Match(allow, addr); match {
					ctx.Next()
					return
				}
			}
			ctx.StatusCode(iris.StatusForbidden)
		})
	}

	for _, router := range this.routers {
		router(this.app)
	}
	if err := errors.Async(time.Second, func() error {
		return this.app.Run(
			iris.Addr(this.address), iris.WithoutBanner,
			iris.WithRemoteAddrHeader("X-Real-Ip"),
			iris.WithRemoteAddrHeader("X-Forwarded-For"),
			iris.WithRemoteAddrHeader("CF-Connecting-IP"),
			iris.WithoutServerError(iris.ErrServerClosed),
		)
	}); err != nil && err != errors.ErrTimeout {
		return err
	}
	logger.Info("start at: http://", this.address)
	return nil
}

func (this *httpServer) Stop() error {
	logger.Info("http server stop.")
	return this.app.Shutdown(context.TODO())
}
