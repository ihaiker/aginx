package cmd

import (
	"fmt"
	"github.com/ihaiker/aginx/lego"
	"github.com/ihaiker/aginx/nginx/client"
	"github.com/ihaiker/aginx/nginx/configuration"
	"github.com/ihaiker/aginx/server"
	ig "github.com/ihaiker/aginx/server/ignore"
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

func getString(cmd *cobra.Command, key, def string) string {
	value, err := cmd.PersistentFlags().GetString(key)
	PanicIfError(err)
	if value == "" {
		value = viper.GetString(key)
	}
	if value == "" {
		return def
	}
	return value
}

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
	location.AddBody("client_max_body_size", "10m")
	location.AddBody("client_body_buffer_size", "128k")
	location.AddBody("proxy_connect_timeout", "90")
	location.AddBody("proxy_send_timeout", "90")
	location.AddBody("proxy_read_timeout", "90")
	location.AddBody("proxy_buffers", "32", "4k")
	return directive
}

func exposeApi(cmd *cobra.Command, address string, engine storage.Engine) {
	domain := getString(cmd, "expose", "")
	if domain == "" {
		return
	}
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

		address := getString(cmd, "api", ":8011")
		auth := getString(cmd, "security", "")

		daemon := NewDaemon()
		cluster := getString(cmd, "cluster", "")
		ignore := ig.Empty()

		engine := clusterConfiguration(cluster, ignore)
		if service, matched := engine.(Service); matched {
			daemon.Add(service)
		}

		exposeApi(cmd, address, engine)

		manager, err := lego.NewManager(engine)
		PanicIfError(err)

		svr := new(server.Supervister)
		routers := server.Routers(svr, engine, manager, auth)
		http := server.NewHttp(address, routers)

		return daemon.Add(http, svr, manager).Start()
	},
}

func AddServerFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP("api", "a", "", "restful api port. (default :8081)")
	cmd.PersistentFlags().StringP("security", "s", "", "base auth for restful api, example: user:passwd")

	cmd.PersistentFlags().StringP("cluster", "c", "", `cluster config
for example. 
	consul://127.0.0.1:8500/aginx[?token=authtoken]   config from consul.  
	zk://127.0.0.1:2182/aginx[?scheme=&auth=]         config from zookeeper.
	etcd://127.0.0.1:2379/aginx[?user=&password]      config from etcd.
`)
	cmd.PersistentFlags().StringP("expose", "e", "", "expose api use domain")
	cmd.PersistentFlags().BoolP("watcher", "w", false, "watcher local file changes and sync to cluster configuration")
}

func init() {
	AddServerFlags(ServerCmd)
	_ = viper.BindPFlags(ServerCmd.PersistentFlags())
}
