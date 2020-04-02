# docker

本章节将详细介绍如何使docker的服务发布到nginx. 

当容器或者服务发生变化后程序将使用模板发布配置到nginx上。

docker支持两种模式：
- 第一种： 标签模式.
- 第二重： 全局模板模式.

## 标签模式（默认方式）

程序将会搜索服务包含相应label的服务使用模板发布到nginx，

**label的正则**
```goregexp
aginx.domain(\.(\d+))?
```

`\d`为需要发布到nginx上作为代理的端口，此端口可以不写。
但是需要符合以下条件中的一个：
- 容器仅仅只公开一个端口
- 容器`EXPOSE`只定义了一个端口

**value的正则**
```goregexp
([a-zA-Z0-9-_\.]*)(,(weight=(\d+)))?(,(internal))?(,(ssl))?(,(virtual))?(,(nodes))?(,(networks=(\\d))
```

**内容说明：**

- weight 定义服务的权重，
- ssl 此服务是否自动申请免费证书并且部署。（如果设置了使用免费证书http监听将会被自动转向https）。
- internal 获取容器的内部地址发布到nginx上。
- virtual 使用vip地址发布到nginx上。（仅在swarm service起效）
- nodes 使用swarm所有节点地址发布到nginx上。（仅在swarm service起效）
- networks 多网络绑定情况下，优先选择网络。例如：10.24.0

**实例：**

- 1、aginx.domain=api.aginx.io
- 2、aginx.domain.8500=api.aginx.io,weight=2
- 3、aginx.domain=api.aginx.io,ssl
- 3、aginx.domain=api.aginx.io,weight=1,ssl
- 4、aginx.domain=api.aginx.io,internal
- 5、aginx.domain=api.aginx.io,virtual
- 6、aginx.domain=api.aginx.io,virtual;api2.aginx.io,internal (1.1.3后版本支持，多域名绑定)

**模板使用：**

程序会查找`--docker-labels-template-dir`（默认 templates/docker）文件夹下定义的模板。并且查找模板存在优先级
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
	server_name {{.Data.Domain}};	
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
        Containers []types.ContainerJSON
        Services   []swarm.Service
        PublishIP  string
        Nodes      []swarm.Node
        Docker     *dockerClient.Client
    }
}
```
