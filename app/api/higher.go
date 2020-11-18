package api

import (
	"fmt"
	"github.com/ihaiker/aginx/v2/core/nginx/config"
	"strconv"
	"strings"
)

const (
	ProtocolHTTP Protocol = "http"
	ProtocolTCP  Protocol = "tcp"
	ProtocolUDP  Protocol = "udp"

	ProxyHTTP     ProxyType = "http"     //反向代理
	ProxyUpstream ProxyType = "upstream" //负载均衡
	ProxyHTML     ProxyType = "html"     //代理静态文件
	ProxyCustom   ProxyType = "custom"   //自定义路径
)

type (
	//代理协议
	Protocol    string
	HostAndPort struct {
		Host string `json:"host,omitempty"`
		Port int    `json:"port"`
	}

	//http://nginx.org/en/docs/http/ngx_http_ssl_module.html
	ServerSSL struct {
		//是否http转接到https
		HTTPRedirect   bool   `json:"httpRedirect"`
		Certificate    string `json:"certificate"`
		CertificateKey string `json:"certificateKey"`
		Protocols      string `json:"protocols"` //TLSv1 TLSv1.1 TLSv1.2 TLSv1.3
	}

	//代理类型
	ProxyType          string
	ServerLocationHTTP struct {
		To string `json:"to"`
	}
	ServerLocationHTML struct {
		Path    string `json:"path"`    //路径
		Model   string `json:"model"`   //root/alias模式
		Indexes string `json:"indexes"` //默认主页
	}

	ServerLocationUpstream struct {
		Name string `json:"name"` //upstream 名字
		Path string `json:"path"` //额外路径
	}

	AuthBasic struct {
		Switch   string `json:"switch"`
		UserFile string `json:"userFile"`
	}

	//代理路径
	ServerLocation struct {
		Commit      string                  `json:"commit,omitempty"` //注释内容，第一行必须是这个
		Path        string                  `json:"path"`
		Type        ProxyType               `json:"type"`
		HTTP        *ServerLocationHTTP     `json:"http,omitempty"`
		Upstream    *ServerLocationUpstream `json:"upstream,omitempty"`
		HTML        *ServerLocationHTML     `json:"html,omitempty"`
		BasicHeader bool                    `json:"basicHeader"` //是否添加基本参数
		WebSocket   bool                    `json:"webSocket"`
		AuthBasic   *AuthBasic              `json:"authBasic,omitempty"` //是否开启auth basic认证
		Parameters  []*config.Directive     `json:"parameters"`
		//允许地址
		Allows []string `json:"allows,omitempty"`
		//禁用地址，允许地址在前，deny 最后
		Denys []string `json:"denys,omitempty"`
	}

	RewriteMobile struct {
		Agents string `json:"agents"`
		Domain string `json:"domain"`
	}

	ServerListen struct {
		HostAndPort

		//是否默认
		Default bool `json:"default"`

		HTTP2 bool `json:"http2"`

		SSL bool `json:"ssl"`
	}
	//服务
	Server struct {
		//用于确定当前server,如果当前server是新添加的此字段可以为空，更新设删除使用这个字段
		Queries []string `json:"queries,omitempty"`
		Commit  string   `json:"commit,omitempty"` //注释内容，第一行必须是这个

		//代理类型
		Protocol Protocol `json:"protocol"`

		//监听地址
		Listens []ServerListen `json:"listens"`

		//监听域名
		Domains []string `json:"domains,omitempty"`

		//是否开启http
		SSL *ServerSSL `json:"ssl,omitempty"`

		//是否开启auth basic认证
		AuthBasic *AuthBasic `json:"authBasic,omitempty"`

		//是否开启自动转向mobile手机端
		RewriteMobile *RewriteMobile `json:"rewriteMobile,omitempty"`

		//监听路径
		Locations []ServerLocation `json:"locations,omitempty"`

		//tcp/upd转发需要
		ProxyPass string `json:"proxyPass,omitempty"`

		//额外参数
		Parameters []*config.Directive `json:"parameters"`

		//允许地址
		Allows []string `json:"allows,omitempty"`

		//禁用地址，允许地址在前，deny 最后
		Denys []string `json:"denys,omitempty"`
	}

	UpstreamServer struct {
		HostAndPort
		Weight      int    `json:"weight"`
		FailTimeout int    `json:"failTimeout"`
		MaxFails    int    `json:"maxFails"`
		Status      string `json:"status"` //down,backup
	}

	Upstream struct {
		//用于确定当前server,如果当前server是新添加的此字段可以为空，更新设删除使用这个字段
		Queries []string `json:"queries,omitempty"`
		Name    string   `json:"name"`
		Commit  string   `json:"commit"` //注释内容，第一行必须是这个
		//代理类型
		Protocol Protocol `json:"protocol"`
		//负载策略 sticky,ip_hash,last_conn,last_time
		LoadStrategy string              `json:"loadStrategy"`
		Servers      []UpstreamServer    `json:"servers"`
		Parameters   []*config.Directive `json:"parameters"`
	}

	Filter struct {
		//HTTP代理搜索域名，TCP代理搜索负载名字，因为tcp代理没有域名问题
		Name       string
		Commit     string
		Protocol   Protocol
		ExactMatch bool
	}
)

func (s *HostAndPort) String() string {
	if s.Port == 0 {
		return s.Host
	} else if s.Host == "" {
		return strconv.Itoa(s.Port)
	}
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

func ParseHostAndPort(str string) HostAndPort {
	idx := strings.Index(str, ":")
	if idx == -1 {
		port, _ := strconv.Atoi(str)
		return HostAndPort{Port: port}
	}
	if strings.HasPrefix(str, "unix://") {
		return HostAndPort{Host: str}
	}
	hap := HostAndPort{Host: str[0:idx]}
	hap.Port, _ = strconv.Atoi(str[idx+1:])
	return hap
}
