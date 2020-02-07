package server

import (
	"context"
	"fmt"
	"github.com/ihaiker/aginx/util"
	"github.com/kataras/iris/v12"
	"github.com/sirupsen/logrus"
	"time"
)

type Http struct {
	app     *iris.Application
	address string
	routers func(app *iris.Application)
}

func NewHttp(address string, routers func(*iris.Application)) *Http {
	return &Http{
		app: iris.New(), address: address, routers: routers,
	}
}

func (this *Http) Start() error {
	this.app.UseGlobal(iris.Gzip)
	this.app.UseGlobal(func(ctx iris.Context) {
		defer func() {
			if err := recover(); err != nil {
				if ctx.IsStopped() {
					return
				}
				ctx.StopExecution()

				ctx.StatusCode(iris.StatusInternalServerError)
				_, _ = ctx.JSON(map[string]string{
					"error":   "InternalServerError",
					"message": fmt.Sprintf("%v", err),
				})
			}
		}()
		ctx.Next()
	})

	this.app.OnErrorCode(iris.StatusNotFound, func(ctx iris.Context) {
		_, _ = ctx.JSON(map[string]string{
			"error":   "notfound",
			"message": "the page not found!",
			"url":     ctx.Request().RequestURI,
		})
	})

	this.app.Get("/health", func(ctx iris.Context) {
		_, _ = ctx.JSON(map[string]string{"status": "UP"})
	})

	this.routers(this.app)

	if err := util.Async(time.Second, func() error {
		return this.app.Run(
			iris.Addr(this.address),
			iris.WithoutBanner,
			iris.WithoutServerError(iris.ErrServerClosed),
		)
	}); err != nil && err != util.ErrTimeout {
		return err
	}
	logrus.Info("http server start at: ", this.address)
	return nil
}

func (this *Http) Stop() error {
	logrus.Info("http server stop.")
	return this.app.Shutdown(context.TODO())
}