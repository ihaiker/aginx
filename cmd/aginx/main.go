package aginx

import (
	"github.com/ihaiker/aginx/lego"
	"github.com/ihaiker/aginx/nginx/client"
	"github.com/ihaiker/aginx/nginx/configuration"
	"github.com/ihaiker/aginx/server"
	"github.com/ihaiker/aginx/storage"
	fileStorage "github.com/ihaiker/aginx/storage/file"
	"github.com/ihaiker/aginx/util"
	"github.com/spf13/cobra"
	"path/filepath"
)

//如果集群配置落地
func clusterDown(engine storage.Engine) error {
	if _, match := engine.(*fileStorage.FileStorage); match {
		return nil
	}

	if _, conf, err := fileStorage.GetInfo(); err != nil {
		return err
	} else if api, err := client.NewClient(engine); err != nil {
		return err
	} else {
		root := filepath.Dir(conf)
		return configuration.Down(root, api.Configuration())
	}
}

func Start(cmd *cobra.Command, args []string) error {
	address, err := cmd.Root().PersistentFlags().GetString("api")
	if err != nil {
		return err
	}

	auth, err := cmd.Root().PersistentFlags().GetString("security")
	if err != nil {
		return err
	}

	daemon := util.NewDaemon()
	engine, err := fileStorage.System()
	if err != nil {
		return err
	} else if err := clusterDown(engine); err != nil {
		return err
	}

	manager, err := lego.NewManager(engine)
	if err != nil {
		return err
	}

	svr := new(server.Supervister)

	routers := server.Routers(svr, engine, manager, auth)
	http := server.NewHttp(address, routers)

	return daemon.Add(http, svr, manager).Start()
}
