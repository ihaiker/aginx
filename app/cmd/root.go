package cmd

import (
	"fmt"
	"github.com/ihaiker/aginx/v2/api"
	"github.com/ihaiker/aginx/v2/core/admin"
	"github.com/ihaiker/aginx/v2/core/certs"
	"github.com/ihaiker/aginx/v2/core/config"
	"github.com/ihaiker/aginx/v2/core/http"
	"github.com/ihaiker/aginx/v2/core/logs"
	"github.com/ihaiker/aginx/v2/core/nginx"
	"github.com/ihaiker/aginx/v2/core/nginx/client"
	"github.com/ihaiker/aginx/v2/core/registry"
	"github.com/ihaiker/aginx/v2/core/registry/functions"
	"github.com/ihaiker/aginx/v2/core/storage"
	"github.com/ihaiker/aginx/v2/core/util"
	"github.com/ihaiker/aginx/v2/core/util/errors"
	"github.com/ihaiker/aginx/v2/core/util/files"
	"github.com/ihaiker/aginx/v2/core/util/services"
	"github.com/ihaiker/aginx/v2/plugins/certificate"
	registryPlugin "github.com/ihaiker/aginx/v2/plugins/registry"
	storagePlugin "github.com/ihaiker/aginx/v2/plugins/storage"
	"github.com/ihaiker/cobrax"
	"github.com/kataras/iris/v12"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"runtime"
	"strings"
)

func getNginx() (bin, prefix, conf string, err error) {
	if config.Config.Nginx == "" {
		if config.Config.Nginx, err = nginx.Lookup(); err != nil {
			return
		} else {
			bin = config.Config.Nginx
		}
	}
	if !files.Exists(config.Config.Nginx) {
		err = fmt.Errorf("not found %s", config.Config.Nginx)
		return
	}
	prefix, conf, err = nginx.HelpInfo(config.Config.Nginx)
	return
}

func getNginxDaemon() (nginx.Daemon, error) {
	if bin, prefix, conf, err := getNginx(); err != nil {
		return nil, err
	} else {
		return nginx.NewDaemon(bin, prefix, conf)
	}
}

func getStorage() (storagePlugin.Plugin, error) {
	if config.Config.Storage == "" { //如果未提供存储器系统nginx通常默认的
		if files.Exists("/etc/nginx/nginx.conf") {
			config.Config.Storage = "file://etc/nginx/nginx.conf"
		} else if _, _, conf, err := getNginx(); err != nil { //获取NGINX系统的
			return nil, err
		} else {
			config.Config.Storage = "file:/" + conf
		}
	}
	return storage.Get(config.Config.Storage)
}

func exposeApi(aginx api.Aginx) error {
	logs.Infof("公开API访问域名：%s", config.Config.Expose)

	var server *api.Server
	servers, err := aginx.GetServers(&api.Filter{
		Name:       config.Config.Expose,
		Protocol:   api.ProtocolHTTP,
		ExactMatch: true,
	})
	if err != nil {
		return err
	}

	if len(servers) == 0 {
		server = new(api.Server)
		server.Protocol = api.ProtocolHTTP
		server.Domains = []string{config.Config.Expose}
		server.Commit = "aginx api"
		server.Listens = []api.ServerListen{
			{
				HostAndPort: api.HostAndPort{Port: 80},
			},
		}
	} else {
		server = servers[0]
	}

	server.Locations = []api.ServerLocation{
		{
			Path: "/", Type: api.ProxyHTTP,
			HTTP: &api.ServerLocationHTTP{
				To: fmt.Sprintf("http://%s", config.Config.Bind),
			},
			BasicHeader: true, WebSocket: true,
		},
	}
	if len(config.Config.AllowIp) > 0 && config.Config.AllowIp[0] != "*" {
		server.Locations[0].Allows = config.Config.AllowIp
		server.Locations[0].Denys = []string{"all"}
	}
	_, err = aginx.SetServer(server)
	return err
}

func loadPlugins() error {
	plugins, err := util.FindPlugins(config.Config.Plugins)
	if err != nil {
		return err
	}
	for name, p := range plugins {
		//storage
		if fn, e := p.Lookup(storagePlugin.PLUGIN_STORAGE); e == nil {
			if loadStorage, match := fn.(storagePlugin.LoadStorage); match {
				p := loadStorage()
				logs.Infof("load storage plugin %s:%s in %s", p.Scheme(), p.Version(), name)
				storage.Plugins[p.Scheme()] = p
			}
		}
		//registry
		if fn, e := p.Lookup(registryPlugin.PLUGIN_REGISTRY); e == nil {
			if regFn, match := fn.(registryPlugin.LoadRegistry); match {
				reg := regFn()
				logs.Infof("load registry plugin %s:%s in %s", reg.Scheme(), reg.Version(), name)
				registry.Plugins[reg.Scheme()] = reg
			}
		}
		//certificates
		if fn, e := p.Lookup(certificate.PLUGIN_CERTIFICATES); e == nil {
			if certFn, match := fn.(certificate.LoadCertificates); match {
				cert := certFn()
				logs.Infof("load certificate plugin %s:%S in ", cert.Scheme(), cert.Version(), name)
				certs.Plugins[cert.Scheme()] = cert
			}
		}
		//functions
		if fn, e := p.Lookup(registryPlugin.PLUGIN_FUNC_MAP); e == nil {
			if loadFn, match := fn.(registryPlugin.LoadRegistryFuncMap); match {
				logs.Infof("load functions plugin %s", name)
				registry.Functions = functions.Merge(registry.Functions, loadFn())
			}
		}
	}
	return nil
}

var root = &cobra.Command{
	Use: "aginx",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if level, err := logrus.ParseLevel(config.Config.LogLevel); err != nil {
			return err
		} else {
			logs.SetLevel(level)
		}
		if config.Config.LogFile != "" && config.Config.LogFile != "stdout" {
			logs.Debug("log file：", config.Config.LogFile)
			if f, err := os.OpenFile(config.Config.LogFile,
				os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0655); err != nil {
				return err
			} else {
				logs.SetOutput(f)
			}
		}
		if err := loadPlugins(); err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		var engine storagePlugin.Plugin
		var daemon nginx.Daemon

		manager := services.Manager()

		var aginx api.Aginx

		if not(config.Config.OnlyAdmin() || config.Config.OnlyRegistry()) {
			if engine, err = getStorage(); err != nil {
				return
			}
			manager.Add(engine)
			if daemon, err = getNginxDaemon(); err != nil {
				return
			}
			//不是daemon节点不需要文件变化重启
			//并且如果不是daemon节点不需要同步数据存储
			if config.Config.HasDaemon() {
				fch := storage.Changed(engine, func(event storagePlugin.FileEvent) {
					if e := daemon.Reload(); e != nil {
						logs.Warn("文件改变重启异常：" + e.Error())
					}
				})
				manager.Add(fch)
				//同步插件应该在之前启动，这样才可以先下载数据然后启动daemon
				manager.Add(daemon)
			}
			if aginx, err = client.New(engine, daemon, config.Config.Cert, config.Config.CertDef); err != nil {
				return
			}
		}

		if aginx == nil && !config.Config.OnlyAdmin() { //并不是只有web节点
			if config.Config.Api == "" {
				err = errors.New("未发现API节点，请指定 --api")
				return
			}
			if !strings.HasPrefix(config.Config.Api, "http://") ||
				!strings.HasPrefix(config.Config.Api, "https://") {
				config.Config.Api = "http://" + config.Config.Api
			}
			user, password := "", ""
			for authUserName, authPassword := range config.Config.Auth {
				user, password = authUserName, authPassword
				break
			}
			aginx = api.New(config.Config.Api, user, password)
			if _, err = aginx.Info(); err != nil {
				return
			}
		}

		routers := make([]func(*iris.Application), 0)

		//启用web控制台
		if config.Config.HasAdmin() {
			logs.Info("启用admin管理台")
			configFiles, _ := cmd.PersistentFlags().GetStringSlice("conf")
			if len(configFiles) == 0 {
				configFiles = []string{"/etc/aginx/aginx.conf"}
			}
			routers = append(routers, admin.Routers(configFiles[0]))
		}
		//禁用api
		if config.Config.HasApi() {
			logs.Info("启用restful api")
			routers = append(routers, http.Routers(aginx))
		}
		//禁用api和web就没有必要开启http了
		if len(routers) != 0 {
			httpServer := http.New(config.Config.Bind, routers...)
			manager.Add(httpServer)
		}

		if config.Config.HasRegistry() {
			regHandler := registry.Handler(aginx)
			for _, r := range config.Config.Registry {
				if _, err = regHandler.Add(r); err != nil {
					return err
				}
			}
			manager.Add(regHandler)
		}

		if config.Config.HasApi() { //只有API节点可以重新申请证书
			renewal := certs.Renewal(aginx, func(cert *api.CertFile) error {
				_, err := aginx.Certs().New(cert.Provider, cert.Domain)
				return err
			})
			manager.Add(renewal)

			//使用域名方式暴露API
			if config.Config.Expose != "" {
				if err = exposeApi(aginx); err != nil {
					return
				}
			}
		}
		return manager.Start()
	},
}

func not(b bool) bool {
	return !b
}

func init() {
	root.SilenceUsage = true
	root.SilenceErrors = true
}

func Execute(version, buildTime, gitTag string) error {
	root.Long = fmt.Sprintf(`aginx: restful api for nginx.
Build: %s, Go: %s, Commit: %s`, buildTime, runtime.Version(), gitTag)
	root.Version = version

	if err := cobrax.Flags(root, config.Config, "", "AGINX", config.Help); err != nil {
		return err
	}
	if err := cobrax.ConfigFrom(root, config.Config, "AGINX_CONF", config.Unmarshal); err != nil {
		return err
	}

	root.AddCommand(syncCmd, pluginCmd, completionCmd)
	return root.Execute()
}
