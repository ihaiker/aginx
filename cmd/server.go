package cmd

import (
	"fmt"
	"github.com/ihaiker/aginx/http"
	"github.com/ihaiker/aginx/lego"
	"github.com/ihaiker/aginx/logs"
	"github.com/ihaiker/aginx/nginx"
	"github.com/ihaiker/aginx/plugins"
	"github.com/ihaiker/aginx/registry"
	"github.com/ihaiker/aginx/storage"
	. "github.com/ihaiker/aginx/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"net"
)

var logger = logs.New("cmd")

func AddServerFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP("email", "u", "aginx@renzhen.la", "Register the current account to the ACME server.")

	cmd.PersistentFlags().StringP("storage", "S", "", `Use centralized storage NGINX configuration, for example. 
	consul://127.0.0.1:8500/aginx[?token=authtoken]   config from consul.  
	zk://127.0.0.1:2182/aginx[?scheme=&auth=]         config from zookeeper.
	etcd://127.0.0.1:2379/aginx[?user=&password]      config from etcd.
`)
	cmd.PersistentFlags().StringP("expose", "e", "", "Exposing API services to NGINX。example: api.aginx.io")
	cmd.PersistentFlags().BoolP("disable-watcher", "", false, `Listen to local configuration file changes and automatically sync to storage.
If you use '--storage' to store the NGINX configuration file, it will be synchronized to the local configuration at startup.`)

	AddRegistryFlag(cmd)
}

func init() {
	AddServerFlags(ServerCmd)
	_ = viper.BindPFlags(ServerCmd.PersistentFlags())
}

func exposeApi(cmd *cobra.Command, address string, engine plugins.StorageEngine) {
	domain := GetString(cmd, "expose", "")
	if domain == "" {
		return
	}
	host, port, err := net.SplitHostPort(address)
	PanicIfError(err)
	if host == "" {
		host = "127.0.0.1"
	}
	apiAddress := fmt.Sprintf("%s:%s", host, port)
	logger.Infof("expose api %s to %s ", domain, apiAddress)

	api := nginx.MustClient(engine)
	err = api.SimpleServer(domain, apiAddress)
	PanicIfError(err)
	PanicIfError(api.Store())
}

var ServerCmd = &cobra.Command{
	Use: "server", Short: "the AGINX server", Long: "the api server", Example: "AGINX server",
	RunE: func(cmd *cobra.Command, args []string) error {
		defer Catch(func(err error) {
			fmt.Println(err)
		})

		email := GetString(cmd, "email", "aginx@renzhen.la")
		address := GetString(cmd, "api", ":8011")
		auth := GetString(cmd, "security", "")

		daemon := NewDaemon()

		storageEngine := storage.NewBridge(GetString(cmd, "storage", ""),
			!GetBool(cmd, "disable-watcher"), nginx.MustConf())

		exposeApi(cmd, address, storageEngine)

		sslManager, err := lego.NewManager(storageEngine)
		PanicIfError(err)

		process := new(nginx.Process)
		http := http.NewHttp(address, http.Routers(email, auth, process, storageEngine, sslManager))
		daemon.Add(storageEngine, http, process, sslManager)

		if registry := registry.FindRegistry(cmd); registry != nil {
			daemon.Add(registry)
		}

		return daemon.Start()
	},
}
