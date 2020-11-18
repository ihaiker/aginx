package registry

import (
	"github.com/ihaiker/aginx/v2/api"
	"net/url"
	"text/template"
)

const (
	//加载自定义插件的方法
	PLUGIN_FUNC_MAP = "LoadRegistryFuncMap"

	//插件加载方法名
	PLUGIN_REGISTRY = "LoadRegistry"
)

type (
	//加载registry插件中间监听器
	LoadRegistry func() Plugin

	//模板方法插件支持
	LoadRegistryFuncMap func() template.FuncMap
)

type (
	Domain struct {
		ID       string   //注册中心服务的ID号
		Domain   string   //注册的域名
		Address  []string //绑定地址
		Weight   int      //权重
		AutoSSL  bool     //是否开启ssl
		Provider string   //证书获取提供商
		Alive    bool     //是否存活，如果不是存活就是删除配置
		Template string   //明确使用模板，如果模板不存在也将使用默认的
	}

	//标签注册方式的事件
	LabelsEvent []Domain

	Plugin interface {
		Scheme() string  //插件前缀
		Name() string    //注册中心名称, consul, docker,
		Version() string //当前版本号
		Help() string    //配置方式帮助

		Watch(config url.URL, aginx api.Aginx) error //监听一个注册器

		Label() <-chan LabelsEvent //levels

		//注册监听体提供的模板方法
		TemplateFuncMap() template.FuncMap
	}
)
