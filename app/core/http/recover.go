package http

import (
	"fmt"
	"github.com/ihaiker/aginx/v2/core/util/errors"
	"github.com/kataras/iris/v12"
)

func recoverHandler() iris.Handler {
	return func(ctx iris.Context) {
		defer func() {
			if err := recover(); err != nil {
				if ctx.IsStopped() {
					return
				}
				ctx.StatusCode(iris.StatusInternalServerError)
				_, match := err.(*errors.WrapError)
				if match {
					_, _ = ctx.JSON(map[string]string{
						"error":   "InternalServerError",
						"message": fmt.Sprintf("%v", err),
					})
				} else {
					_, _ = ctx.JSON(map[string]string{
						"error":   "InternalServerError",
						"message": "系统异常请稍后重试！",
					})
					logger.Error("handler error: ", err, ", \nstack:", errors.Stack())
				}
				ctx.StopExecution()
			}
		}()
		ctx.Next()
	}
}
