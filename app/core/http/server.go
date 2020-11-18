package http

import (
	"github.com/ihaiker/aginx/v2/api"
	"github.com/ihaiker/aginx/v2/core/util/errors"
	"github.com/kataras/iris/v12"
)

type serverAndUpstreamController struct {
	aginx api.Aginx
}

func (sau *serverAndUpstreamController) GetServer(ctx iris.Context) []*api.Server {
	filter := new(api.Filter)
	filter.Name = ctx.URLParam("name")
	filter.Commit = ctx.URLParam("commit")
	filter.Protocol = api.Protocol(ctx.URLParam("protocol"))
	filter.ExactMatch = ctx.URLParam("exactMatch") == "true"
	server, err := sau.aginx.GetServers(filter)
	errors.Panic(err)
	return server
}

func (sua *serverAndUpstreamController) SetServer(ctx iris.Context) []string {
	server := new(api.Server)
	errors.PanicMessage(ctx.ReadJSON(server), "表单信息错误")
	queries, err := sua.aginx.SetServer(server)
	errors.Panic(err)
	return queries
}

func (sau *serverAndUpstreamController) GetUpstream(ctx iris.Context) []*api.Upstream {
	filter := new(api.Filter)
	filter.Name = ctx.URLParam("name")
	filter.Commit = ctx.URLParam("commit")
	filter.Protocol = api.Protocol(ctx.URLParam("protocol"))
	filter.ExactMatch = ctx.URLParam("exactMatch") == "true"
	upstreams, err := sau.aginx.GetUpstream(filter)
	errors.Panic(err)
	return upstreams
}

func (sua *serverAndUpstreamController) SetUpstream(ctx iris.Context) []string {
	upstream := new(api.Upstream)
	errors.PanicMessage(ctx.ReadJSON(upstream), "表单信息错误")
	queries, err := sua.aginx.SetUpstream(upstream)
	errors.Panic(err)
	return queries
}
