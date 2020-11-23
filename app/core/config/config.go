package config

type (
	backup struct {
		Dir      string `help:"备份文件位置" def:"./backups"`
		Crontab  string `help:"定时备份时间策略，配置方式可以参阅crontab。"`
		Limit    int    `help:"备份最大个数" def:"30"`
		DayLimit int    `help:"备份最大保存天数" def:"7"`
	}

	config struct {
		LogFile  string `flag:"log-file" help:"日志输出到文件的位置，默认输出到控制台"`
		LogLevel string `flag:"log-level" short:"L" help:"日志级别" def:"info"`

		//管理端
		Bind string `help:"api服务开放地址" def:"127.0.0.1:8011"`
		Api  string `help:"连接API节点地址"`

		//安全控制，允许调用的IP
		AllowIp []string `help:"api服务允许调用的IP地址" def:"*"`

		//管理用户
		Auth map[string]string `help:"管理台认证用户" def:"aginx=aginx"`

		DisableDaemon bool `help:"禁用nginx托管，禁用后将不会托管启动nginx"`

		//禁用API服务
		DisableApi bool `help:"禁用API服务" def:"false"`

		//开放域名给api
		Expose       string `help:"为API服务暴露一个域名。例如: api.aginx.io 或 api.aginx.io,ssl"`
		DisableAdmin bool   `help:"禁用管理控制台" def:"false"`

		Nginx string `help:"nginx 可执行程序的位置，默认将自动搜索.如果搜索不到并且未指定将报错"`

		//配置存储方式
		Storage string `help:"{{storage.help}}" short:"S"`
		//注册管理器
		Registry []string `help:"{{registry.help}}" short:"R"`

		//cert证书插件
		Cert    []string `help:"使用 aginx help certs <provider> 查询更新帮助信息" short:"C" def:"lego://aginx@renzhen.la/certs/lego,custom://certs/custom"`
		CertDef string   `help:"默认cert使用名字" def:"lego"`

		Plugins string `help:"插件文件夹" short:"P" def:"./plugins"`

		Backup backup
	}
)

var (
	Config = &config{Backup: backup{}}
)

func (c *config) HasDaemon() bool {
	return !c.DisableDaemon
}

//是不是含有api节点
func (c *config) HasApi() bool {
	return !c.DisableApi
}

//是不是仅仅只有web节点
func (c *config) OnlyAdmin() bool {
	return c.HasAdmin() && c.DisableApi && c.DisableDaemon && (!c.HasRegistry())
}

//是不是仅仅只有web节点
func (c *config) OnlyRegistry() bool {
	return c.DisableAdmin && c.DisableApi && c.DisableDaemon && c.HasRegistry()
}

func (c *config) HasAdmin() bool {
	return !c.DisableAdmin
}

func (c *config) HasRegistry() bool {
	return c.Registry != nil && len(c.Registry) > 0
}

func Help(_, _, value string) string {
	if value == "{{storage.help}}" {
		return `集中存储配置方式.
	consul://127.0.0.1:8500/aginx[?token=authtoken]   consul k/v配置.
	zk://127.0.0.1:2182/aginx[?scheme=&auth=]         zookeeper 配置.
	etcd://127.0.0.1:2379/aginx[?user=&password]      
	file://etc/nginx/nginx.conf                       本机配置
`
	} else if value == "{{registry.help}}" {
		return `配置注册管理器`
	}
	return ""
}
