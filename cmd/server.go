package cmd

import (
	"fmt"
	"github.com/ihaiker/aginx/lego"
	"github.com/ihaiker/aginx/server"
	"github.com/ihaiker/aginx/storage"
	"github.com/ihaiker/aginx/storage/consul"
	fileStorage "github.com/ihaiker/aginx/storage/file"
	. "github.com/ihaiker/aginx/util"
	"github.com/spf13/cobra"
	"net/url"
	"os"
	"strings"
)

func getString(cmd *cobra.Command, key string) string {
	envKey := strings.ToUpper(fmt.Sprintf("aginx_%s", key))
	if value := os.Getenv(envKey); value != "" {
		return value
	}
	value, err := cmd.PersistentFlags().GetString(key)
	PanicIfError(err)
	return value
}

func clusterConfiguration(cmd *cobra.Command) (engine storage.Engine) {
	var err error
	cluster := getString(cmd, "cluster")
	if cluster == "" {
		engine, err = fileStorage.System()
		PanicIfError(err)
	} else {
		config, err := url.Parse(cluster)
		PanicIfError(err)

		switch config.Scheme {
		case "consul":
			token := config.Query().Get("token")
			folder := config.EscapedPath()[1:]
			engine, err = consul.New(config.Host, folder, token)
			PanicIfError(err)
		}
	}
	return
}

var ServerCmd = &cobra.Command{
	Use: "server", Long: "the api server",
	RunE: func(cmd *cobra.Command, args []string) error {
		defer Catch(func(err error) {
			fmt.Println(err)
		})

		address := getString(cmd, "api")
		auth := getString(cmd, "security")

		daemon := NewDaemon()
		engine := clusterConfiguration(cmd)
		if service, matched := engine.(Service); matched {
			daemon.Add(service)
		}

		manager, err := lego.NewManager(engine)
		PanicIfError(err)

		svr := new(server.Supervister)
		routers := server.Routers(svr, engine, manager, auth)
		http := server.NewHttp(address, routers)

		return daemon.Add(http, svr, manager).Start()
	},
}

func AddServerFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP("api", "a", ":8011", "restful api port")
	cmd.PersistentFlags().StringP("security", "s", "", "base auth for restful api, example: user:passwd")
	cmd.PersistentFlags().StringP("cluster", "c", "", `cluster config. 
for example. 
	consul://127.0.0.1:8500/aginx?token=authtoken   config from consul.  
	zk://127.0.0.1:2182/aginx                       config from zookeeper.
	etcd://127.0.0.1:1234/aginx                     config from etcd.
`)
}

func init() {
	AddServerFlags(ServerCmd)
}
