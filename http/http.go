package http

import (
	"context"
	"fmt"
	"github.com/ihaiker/aginx/util"
	"github.com/kataras/iris/v12"
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
	this.app.Use(func(ctx iris.Context) {
		defer func() {
			if err := recover(); err != nil {
				if ctx.IsStopped() {
					return
				}
				ctx.StatusCode(iris.StatusInternalServerError)
				_, _ = ctx.JSON(map[string]string{
					"error":   "InternalServerError",
					"message": fmt.Sprintf("%v", err),
				})
				if _, match := err.(*util.WrapError); !match {
					logger.Error("handler error: ", err)
				}
				ctx.StopExecution()
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
	logger.Info("start restful api at: ", this.address)
	return nil
}

func (this *Http) Stop() error {
	logger.Info("http server stop.")
	return this.app.Shutdown(context.TODO())
}
