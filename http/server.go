package http

import (
	"github.com/ihaiker/aginx/api"
	"github.com/ihaiker/aginx/nginx"
	"github.com/ihaiker/aginx/util"
	"github.com/kataras/iris/v12/context"
	"net/http"
)

type serverController struct {
	process *nginx.Process
}

func (sc *serverController) listServer(service *nginx.Service) []*api.Server {
	return service.ListService()
}
func (sc *serverController) deleteServer(names []string, service *nginx.Service) int {
	util.PanicMessage(service.DeleteService(names[0]), "delete")
	util.PanicIfError(sc.process.Test(service.Configuration()))
	util.PanicIfError(service.Store())
	return http.StatusNoContent
}

func (sc *serverController) postServer(ctx context.Context, names []string, service *nginx.Service) int {
	server := &api.Server{}
	util.PanicMessage(ctx.ReadJSON(server), "read json")
	service.ModifyService(names, server)
	util.PanicIfError(sc.process.Test(service.Configuration()))
	util.PanicIfError(service.Store())
	return http.StatusNoContent
}
