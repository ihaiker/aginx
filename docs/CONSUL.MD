# consul 

本章节将详细介绍如何使consul注册的服务发布到nginx.

consul支持两种模式：
- 第一种： 标签模式
- 第二重： 全局模板模式，此模式有点像 `consul-template`

## 标签模式（默认方式）

程序将会搜索服务中meta中包含`aginx.domain`的服务使用模板发布到nginx，

meta值需要符合下面的正则:
```regexp
([a-zA-Z0-9-_\.]*)(,(weight=(\d+)))?(,(ssl))?
```
meta之内容说明：

- weight 定义服务的权重，如果未定将使用consul定义的权重值
- ssl 此服务是否自动申请免费证书，并且部署。（如果设置了使用免费证书http监听将会被自动转向https）。

实例：
- 1、aginx.domain=api.aginx.io
- 2、aginx.domain=api.aginx.io,weight=2
- 3、aginx.domain=api.aginx.io,ssl
- 3、aginx.domain=api.aginx.io,weight=1,ssl


模板使用：
程序会查找`--consul-labels-template-dir`（默认 templates/consul）文件夹下定义的模板。并且查找模板存在优先级
- 1、${domain}.ngx.tpl 和域名相同名称的
- 2、default.tpl
- 3、label模式下默认模板。

掺入模板数据
```go
type TemplateDate struct {
    Aginx api.Aginx
    Data  map[string]interface{}{
        "Domain":  "demo.aginx.io",
        "AutoSSL": false,
        "Servers": []*struct {
            ID      string
            Domain  string
            Address string
            Weight  int
            AutoSSL bool
            Attrs   map[string]string
        },
    }
}
```


系统默认模板
```gotemplate
upstream {{ upstreamName .Data.Domain }} { {{range .Data.Servers}}
	server {{.Address}} {{if ne .Weight 0}} weight={{.Weight}}{{end}};{{end}}
}
{{if .Data.AutoSSL}}server {
	listen       80;
	server_name {{.Domain}};	
	return 301 https://$host$request_uri;
}{{end}}
server { {{if .Data.AutoSSL}}
	listen 443 ssl;
	ssl_certificate     {{.Data.SSL.Certificate}};        
	ssl_certificate_key {{.Data.SSL.PrivateKey}};
	ssl_session_timeout 5m;
	ssl_ciphers ECDHE-RSA-AES128-GCM-SHA256:ECDHE:ECDH:AES:HIGH:!NULL:!aNULL:!MD5:!ADH:!RC4;
	ssl_protocols TLSv1 TLSv1.1 TLSv1.2;
	ssl_prefer_server_ciphers on; {{else}}
	listen 80; {{end}}

    server_name {{.Data.Domain}};
    try_files $uri @tornado;

    location @tornado {
        proxy_set_header        X-Scheme        $scheme;
        proxy_set_header        Host            $host;
        proxy_set_header        X-Real-IP       $remote_addr;
        proxy_set_header        X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_pass http://{{ upstreamName .Data.Domain }};
    }
}
```

## 模板模式

当使用模板模式后程序会在service改变后直接使用定义模板生成配置文件到nginx。

传入模板数据内容
```go
type TemplateDate struct {
    Aginx api.Aginx
    Data struct {
        Services map[string][]*consul.ServiceEntry
        Keys     map[string]*consul.KVPair
        Consul   *consul.Client
    }
}
```
