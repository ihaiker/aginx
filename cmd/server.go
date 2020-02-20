package cmd

import (
	"fmt"
	"github.com/ihaiker/aginx/lego"
	"github.com/ihaiker/aginx/logs"
	"github.com/ihaiker/aginx/nginx/client"
	"github.com/ihaiker/aginx/nginx/configuration"
	nginxDaemon "github.com/ihaiker/aginx/nginx/daemon"
	"github.com/ihaiker/aginx/server"
	ig "github.com/ihaiker/aginx/server/ignore"
	"github.com/ihaiker/aginx/server/watcher"
	"github.com/ihaiker/aginx/storage"
	"github.com/ihaiker/aginx/storage/consul"
	"github.com/ihaiker/aginx/storage/etcd"
	fileStorage "github.com/ihaiker/aginx/storage/file"
	"github.com/ihaiker/aginx/storage/zookeeper"
	. "github.com/ihaiker/aginx/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"net/url"
	"strings"
)

var logger = logs.New("server-cmd")

func clusterConfiguration(cluster string, ignore ig.Ignore) (engine storage.Engine) {
	var err error
	if cluster == "" {
		engine, err = fileStorage.System()
		PanicIfError(err)
	} else {
		config, err := url.Parse(cluster)
		PanicIfError(err)
		switch config.Scheme {
		case "consul":
			engine, err = consul.New(config, ignore)
			PanicIfError(err)
		case "etcd":
			engine, err = etcd.New(config, ignore)
			PanicIfError(err)
		case "zk":
			engine, err = zookeeper.New(config, ignore)
			PanicIfError(err)
		}
	}
	return
}

func selectDirective(api *client.Client, domain string) (queries []string, directive *configuration.Directive) {
	serverQuery := fmt.Sprintf("server.[server_name('%s') & listen('80')]", domain)
	queries = client.Queries("http", "include", "*", serverQuery)
	if directives, err := api.Select(queries...); err == nil {
		directive = directives[0]
		return
	}
	queries = client.Queries("http", serverQuery)
	if directives, err := api.Select(queries...); err == nil {
		directive = directives[0]
		return
	}
	return
}

func apiServer(domain, address string) *configuration.Directive {
	directive := configuration.NewDirective("server")
	directive.AddBody("listen", "80")
	directive.AddBody("server_name", domain)

	if strings.HasPrefix(address, ":") {
		address = "127.0.0.1" + address
	}
	location := directive.AddBody("location", "/")
	location.AddBody("proxy_pass", fmt.Sprintf("http://%s", address))
	location.AddBody("proxy_set_header", "Host", domain)
	location.AddBody("proxy_set_header", "X-Real-IP", "$remote_addr")
	location.AddBody("proxy_set_header", "X-Forwarded-For", "$proxy_add_x_forwarded_for")
	return directive
}

func exposeApi(cmd *cobra.Command, address string, engine storage.Engine) {
	domain := GetString(cmd, "expose", "")
	if domain == "" {
		return
	}
	logger.Info("expose api for : ", domain)
	api, err := client.NewClient(engine)
	PanicIfError(err)

	_, directive := selectDirective(api, domain)
	if directive == nil {
		apiServer := apiServer(domain, address)

		err = api.Add(client.Queries("http"), apiServer)
		PanicIfError(err)

		err = engine.StoreConfiguration(api.Configuration())
		PanicIfError(err)
	}
}

var ServerCmd = &cobra.Command{
	Use: "server", Short: "the aginx server", Long: "the api server", Example: "aginx server",
	RunE: func(cmd *cobra.Command, args []string) error {
		defer Catch(func(err error) {
			fmt.Println(err)
		})

		email := GetString(cmd, "email", "aginx@renzhen.la")
		address := GetString(cmd, "api", ":8011")
		auth := GetString(cmd, "security", "")
		storage := GetString(cmd, "storage", "")
		withWatcher := (storage != "") && GetBool(cmd, "watcher")

		daemon := NewDaemon()

		var ignore ig.Ignore = ig.Empty()
		if withWatcher {
			ignore = ig.Cluster()
		}
		storageEngine := clusterConfiguration(storage, ignore)
		if service, matched := storageEngine.(Service); matched {
			daemon.Add(service)
		}
		exposeApi(cmd, address, storageEngine)

		sslManager, err := lego.NewManager(storageEngine)
		PanicIfError(err)

		svr := new(nginxDaemon.Supervister)
		routers := server.Routers(email, svr, storageEngine, sslManager, auth)
		http := server.NewHttp(address, routers)
		daemon.Add(http, svr, sslManager)

		if bridge, err := findBridge(cmd); err != nil {
			return nil
		} else if bridge != nil {
			daemon.Add(bridge)
		}

		if withWatcher {
			daemon.Add(watcher.NewFileWatcher(storageEngine, ignore))
		}
		return daemon.Start()
	},
}

func AddServerFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP("email", "u", "aginx@renzhen.la", "Register the current account to the ACME server.")

	cmd.PersistentFlags().StringP("storage", "S", "", `Use centralized storage NGINX configuration, for example. 
	consul://127.0.0.1:8500/aginx[?token=authtoken]   config from consul.  
	zk://127.0.0.1:2182/aginx[?scheme=&auth=]         config from zookeeper.
	etcd://127.0.0.1:2379/aginx[?user=&password]      config from etcd.
`)
	cmd.PersistentFlags().StringP("expose", "e", "", "Exposing API services to NGINX。example: api.aginx.io")
	cmd.PersistentFlags().BoolP("watcher", "w", false, `Listen to local configuration file changes and automatically sync to storage.
If you use 'storage' to store the NGINX configuration file, it will be synchronized to the local configuration at startup.
`)
	AddRegistryFlag(cmd)
}

func init() {
	AddServerFlags(ServerCmd)
	_ = viper.BindPFlags(ServerCmd.PersistentFlags())
}
