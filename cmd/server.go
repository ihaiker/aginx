package cmd

import (
	"fmt"
	"github.com/ihaiker/aginx/conf"
	"github.com/ihaiker/aginx/http"
	"github.com/ihaiker/aginx/lego"
	"github.com/ihaiker/aginx/logs"
	"github.com/ihaiker/aginx/nginx"
	"github.com/ihaiker/aginx/registry"
	"github.com/ihaiker/aginx/storage"
	. "github.com/ihaiker/aginx/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"net"
	"strings"
)

var logger = logs.New("cmd")

func AddServerFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP("email", "u", "aginx@renzhen.la", "Register the current account to the ACME server.")

	cmd.PersistentFlags().StringP("storage", "S", "", `Use centralized storage NGINX configuration, for example. 
	consul://127.0.0.1:8500/aginx[?token=authtoken]   config from consul.  
	zk://127.0.0.1:2182/aginx[?scheme=&auth=]         config from zookeeper.
	etcd://127.0.0.1:2379/aginx[?user=&password]      config from etcd.
`)
	cmd.PersistentFlags().StringP("expose", "e", "", "Exposing API services to NGINX。example: api.aginx.io or api.aginx.io,ssl")
	cmd.PersistentFlags().BoolP("disable-watcher", "", false, `Listen to local configuration file changes and automatically sync to storage.
If you use '--storage' to store the NGINX configuration file, it will be synchronized to the local configuration at startup.`)

	cmd.PersistentFlags().StringArrayP("server", "", []string{}, "Adding a simple service proxy.\n"+
		"example: --server 'a1.aginx.io=172.0.0.1:8080' --server 'a2.aginx.io=ssl,172.0.0.1:8083,127.0.0.1:8084'")

	AddRegistryFlag(cmd)
}

func init() {
	AddServerFlags(ServerCmd)
	_ = viper.BindPFlags(ServerCmd.PersistentFlags())
}

func exposeApi(address string, api *nginx.Client) bool {
	domain := viper.GetString("expose")
	if domain == "" {
		return false
	}
	host, port, err := net.SplitHostPort(address)
	PanicIfError(err)
	//host 如果不是指定了，就要获取地址
	if host == "" || host == "0.0.0.0" {
		host = GetRecommendIp()[0]
	}
	apiAddress := fmt.Sprintf("%s:%s", host, port)
	logger.Infof("expose api %s to %s ", domain, apiAddress)

	domainAndSsl := strings.Split(domain, ",")
	ssl := len(domainAndSsl) == 2 && domainAndSsl[1] == "ssl"
	if ssl {
		domain = domainAndSsl[0]
	}
	err = api.SimpleServer(domain, ssl, apiAddress)
	PanicIfError(err)
	return true
}

func simpleServer(cmd *cobra.Command, api *nginx.Client) bool {
	services := GetStringArray(cmd, "server")
	for _, server := range services {
		kva := strings.SplitN(server, "=", 2)
		domain := kva[0]
		proxies := strings.Split(kva[1], ",")
		ssl := proxies[0] == "ssl"
		if ssl {
			proxies = proxies[1:]
		}
		PanicIfError(api.SimpleServer(domain, ssl, proxies...))
	}
	return len(services) > 0
}

var ServerCmd = &cobra.Command{
	Use: "server", Short: "the AGINX server", Long: "the api server", Example: "AGINX server",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if configFile := viper.GetString("conf"); configFile != "" {
			return conf.ReadConfig(configFile, cmd)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		defer Catch(func(err error) {
			cmd.PrintErrln(err)
			fmt.Println(Stack())
		})

		email := viper.GetString("email")
		address := viper.GetString("api")
		auth := viper.GetString("security")

		daemon := NewDaemon()
		storageEngine := storage.NewBridge(viper.GetString("storage"),
			!viper.GetBool("disable-watcher"), nginx.MustConf())

		manager, err := lego.NewManager(storageEngine)
		PanicIfError(err)

		process := new(nginx.Process)
		http := http.NewHttp(address, http.Routers(email, auth, process, storageEngine, manager))

		daemon.Add(storageEngine, http, process, manager)
		daemon.AddStart(func() error {
			api := nginx.MustClient(email, storageEngine, manager, process)
			writeApi := exposeApi(address, api)
			writeSimpleServer := simpleServer(cmd, api)
			if writeApi || writeSimpleServer {
				return api.Store()
			}
			return nil
		})

		if registry := registry.FindRegistry(cmd); registry != nil {
			daemon.Add(registry)
		}
		return daemon.Start()
	},
}
