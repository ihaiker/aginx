package aginx

import (
	"github.com/ihaiker/aginx/server"
	"github.com/ihaiker/aginx/util"
	"github.com/spf13/cobra"
)

func Start(cmd *cobra.Command, args []string) error {
	address, err := cmd.Root().PersistentFlags().GetString("api")
	if err != nil {
		return err
	}
	auth, err := cmd.Root().PersistentFlags().GetString("security")
	if err != nil {
		return err
	}
	d := util.NewDaemon()
	vister := new(server.Supervister)
	routers := server.Routers(vister, auth)
	http := server.NewHttp(address, routers)
	d.Add(vister, http)
	return d.Start()
}
