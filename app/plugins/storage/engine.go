package storage

import (
	"net/url"
)

//加载存储插件的方法

const (
	//插件加载的名字
	PLUGIN_STORAGE = "LoadStorage"
)

type (
	LoadStorage func() Plugin

	Plugin interface {
		Scheme() string  //插件前缀
		Name() string    //存储插件名称，file,zk,consul
		Version() string //当前版本号
		Help() string    //配置方式帮助
		GetConfig() url.URL

		Initialize(config url.URL) error //一个存储器

		//获取文件变动监听，如果不支持可以不返回事件
		Listener() <-chan FileEvent

		//存储文件内容
		Put(file string, content []byte) error

		//删除文件
		Remove(file string) error

		//搜索文件,如果pattern长度为空，则显示全部文件，例如 vhost.d/*.conf
		Search(pattern ...string) ([]*File, error)

		//获取文件
		Get(file string) (*File, error)
	}
)
