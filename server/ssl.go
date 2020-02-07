package server

import (
	"github.com/ihaiker/aginx/lego"
	"github.com/ihaiker/aginx/storage"
)

type sslController struct {
	vister  *Supervister
	engine  storage.Engine
	manager *lego.Manager
}
