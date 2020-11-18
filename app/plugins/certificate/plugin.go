package certificate

import (
	"github.com/ihaiker/aginx/v2/api"
	"net/url"
	time "time"
)

const (
	//插件加载方法名
	PLUGIN_CERTIFICATES = "LoadCertificates"
)

type (
	LoadCertificates func() Plugin

	//证书文件
	Files interface {
		//过期时间
		GetExpiredTime() time.Time

		//证书公钥
		Certificate() string

		//证书私钥文件
		PrivateKey() string
	}

	Plugin interface {
		Scheme() string  //插件前缀
		Name() string    //注册中心名称
		Version() string //当前版本号
		Help() string    //配置方式帮助

		Initialize(config url.URL, aginx api.Aginx) error

		//申请一个SSL证书，可以调用多次，多次调用每次都会申请
		New(domain string) (Files, error)

		//获取SSL证书
		Get(domain string) (Files, error)

		//获取证书列表
		List() (map[string]Files, error)
	}
)
