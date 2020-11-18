package http

import (
	"encoding/json"
	"github.com/ihaiker/aginx/v2/api"
	"github.com/ihaiker/aginx/v2/core/nginx/config"
	"github.com/ihaiker/aginx/v2/core/util/errors"
	"github.com/kataras/iris/v12"
)

type directiveController struct {
	aginx api.Aginx
}

func (d *directiveController) Select(ctx iris.Context) []*config.Directive {
	queries := ctx.Request().URL.Query()["q"]
	selDirective, err := d.aginx.Directive().Select(queries...)
	errors.Panic(err)
	return selDirective
}

func (d *directiveController) Add(ctx iris.Context) int {
	queries := ctx.Request().URL.Query()["q"]
	if queries == nil {
		queries = []string{}
	}
	pretty := ctx.URLParam("pretty")

	body, err := ctx.GetBody()
	errors.PanicMessage(err, "添加内容获取错误")

	var addDirective []*config.Directive
	if pretty == "json" {
		err := json.Unmarshal(body, &addDirective)
		errors.PanicMessage(err, "解析添加内容错误")
	} else {
		cfg, err := config.ParseWith("body.conf", body, nil)
		errors.PanicMessage(err, "解析添加内容错误")
		addDirective = cfg.Body
	}
	err = d.aginx.Directive().Add(queries, addDirective...)
	errors.Panic(err)
	return iris.StatusNoContent
}

func (d *directiveController) Delete(ctx iris.Context) int {
	queries := ctx.Request().URL.Query()["q"]
	err := d.aginx.Directive().Delete(queries...)
	errors.Panic(err)
	return iris.StatusNoContent
}

func (d *directiveController) Modify(ctx iris.Context) int {
	queries := ctx.Request().URL.Query()["q"]
	pretty := ctx.URLParam("pretty")

	body, err := ctx.GetBody()
	errors.PanicMessage(err, "添加内容获取错误")

	var directive *config.Directive
	if pretty == "json" {
		err := json.Unmarshal(body, directive)
		errors.PanicMessage(err, "解析添加内容错误")
	} else {
		cfg, err := config.ParseWith("body.conf", body, nil)
		errors.PanicMessage(err, "解析添加内容错误")
		errors.Assert(len(cfg.Body) == 1, "更新内容只能为一项")
		directive = cfg.Body[0]
	}
	err = d.aginx.Directive().Modify(queries, directive)
	errors.Panic(err)
	return iris.StatusNoContent
}

func (d *directiveController) Batch(ctx iris.Context) int {
	body, err := ctx.GetBody()
	errors.PanicMessage(err, "添加内容获取错误")

	batchs := make([]*api.DirectiveBatch, 0)
	errors.PanicMessage(json.Unmarshal(body, &batchs), "解析内容错误")

	err = d.aginx.Directive().Batch(batchs)
	errors.Panic(err)

	return iris.StatusNoContent
}
