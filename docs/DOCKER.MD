# docker

本章节将详细介绍如何使docker的服务发布到nginx. 

当容器或者服务发生变化后程序将使用模板发布配置到nginx上。

docker支持两种模式：
- 第一种： 标签嵌入模式.
- 第二重： 零耦合模式（推荐使用）.

## 标签嵌入模式

程序将会搜索服务包含相应label的服务使用模板发布到nginx，

**搜索的标签**

```goregexp
aginx.domain.0
aginx.domain.1
...
aginx.domain.N
```

**标签值：**使用url 方式。

使用https将自动申请TLS证书



**参数说明：**

- port 绑定的端口。如果在多开放端口或者没有开放端口的情况下必须指定
- weight 定义服务的权重，
- 如果设置使用https协议，将自动申请免费证书。
- internal 获取容器的内部地址发布到nginx上。
- networks 多网络绑定情况下，优先选择网络。例如：10.24.0,或者网络名称
- provider 申请免费TLS证书的提供商名称。 
- template  指定使用模板生成配置。

**实例：**

- 1、http://api.aginx.io
- 2、http://api.aginx.io?weight=2&port=8500
- 3、https://api.aginx.io?provider=lego&internal=true

**模板使用优先级：**

程序会查找（默认 ）文件夹下定义的模板。

- 0、template 参数指定的模板 。(路径：templates/${template}.tpl)
- 1、templates/${domain}.tpl 和域名相同名称的
- 2、templates/default.tpl

**系统默认模板 （templates/default.tpl）**

```gotemplate
upstream {{ .UpstreamName }} { {{range .Servers}}
	server {{.Host}}:{{.Port}} {{if ne .Weight 0}} weight={{.Weight}}{{end}};{{end}}
}

server { 
	server_name {{.Domain}};
	listen 80;

{{if .AutoSSL}}
	listen 443 ssl;
	ssl_certificate     {{.Cert.Certificate}};        
	ssl_certificate_key {{.Cert.PrivateKey}};
	ssl_session_timeout 5m;
	ssl_ciphers ECDHE-RSA-AES128-GCM-SHA256:ECDHE:ECDH:AES:HIGH:!NULL:!aNULL:!MD5:!ADH:!RC4;
	ssl_protocols TLSv1 TLSv1.1 TLSv1.2;
	ssl_prefer_server_ciphers on;

	if ( $scheme = 'http' ) {
       return 301 https://$host$request_uri;
    }
{{end}}

    location / {
        proxy_set_header        X-Scheme        $scheme;
        proxy_set_header        Host            $host;
        proxy_set_header        X-Real-IP       $remote_addr;
        proxy_set_header        X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_pass http://{{ .UpstreamName }};
    }
}
```



## 零耦合模式（推荐使用）

零耦合方式是使用外部文件定义每个服务的labels，此文件存储位置为：`<storage>`/registry/docker-labels.conf  。其中 <storage> 为存储方式。

此文件搜索container并给特定的container打labels。并且含有两种搜索模式：

1、容器名称匹配搜索：

```nginx
containerName {
  aginx.domain.0: http://test.aginx.io;
  aginx.domain.1: http://t2.aginx.io;
}
```

**containerName**:  容器名称



2、标签匹配搜索：

```nginx
aginx.labels labelName=labelValue {
   		标签: 标签值;
      aginx.domain.0: http://consul.aginx.io;
}
```

**labelName**: 容器现有标签名称，**labelValue: ** 容器现有标签匹配值。



此种方式可以不对任何container容器做任何处理就可以直接发布服务到nginx （零耦合），并且标签和标签值的写法和label匹配模式完全一致，服务也容易迁移。

