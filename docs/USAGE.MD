# 使用介绍

## 启动aginx

>   aginx服务分为四个节点：api节点，web 控制台节点，registry注册中心发现节点和nginx托管运行 。



### 1、AllinOne 启动

```shell
aginx
```

此命令提供api管理服务，运行此服务需要注意`nginx`命令可以找到，不然会报错，程序会托管`nginx`一起启动，并不需要单独启动。
如果你已经启动了`nginx`此程序会尝试先行关闭它并且托管启动。



### 2、分布式运行

>   注意：分布式运行需要提供统一的存储方式。并且同步配置到集中存储

同步命令 `aginx sync consul://192.168.1.9:8500/aginx`



| 节点说明          | 节点IP       | 启动命令                                                     |
| ----------------- | ------------ | ------------------------------------------------------------ |
| consol数据存储    | 192.168.1.9  |                                                              |
| API节点           | 192.168.1.10 | aginx --bind 192.168.1.10:8011 --disable-admin --disable-daemon --storage consul://192.168.1.123:8500/aginx |
| web控制台节点     | 192.168.1.11 | aginx --bind 192.168.1.11:8012 --auth aginx=aginx --disable-api --disable-daemon |
| registry注册节点  | 192.168.1.12 | aginx --disable-admin --disable-api --disable-daemon --registry  docker://var/run/docker.sock --api  192.168.1.10:8011<br />aginx --disable-admin --disable-api --disable-daemon --registry docker://192.168.1.xx:2375  --api  192.168.1.10:8011<br />aginx --disable-admin --disable-api --disable-daemon --registry consul://192.168.1.9:8500  --api  192.168.1.10:8011 |
| nginx托管启动节点 | 192.168.1.13 | aginx --disable-admin --disable-api  --storage consul://192.168.1.123:8500/aginx |



>    测试

```shell script
curl http://128.0.0.1:8011/info
```



## 特性说明

### 一、restful api

Restful api 提供了管理和配置`nginx`的命令。有关全部restful api 内容查阅相关章节：[restful api](./RESTFULAPI.MD)



### 二、sdk for aginx

本程序提供了`sdk`，所使用此SDK你可以自己开发配置管理nginx的服务，当然使用sdk您需要启动`aginx`主服务，此sdk只是调用`restful api`。

下载sdk
```shell script
go get github.com/ihaiker/aginx/v2/api
```

使用sdk
```go
package test

import (
	"github.com/ihaiker/aginx/api"
	"testing"
)

func TestAginx_Add(t *testing.T) {
    api := New("http://127.0.0.1:8011","user","password")
	{
		err := api.Directive().Delete("worker_rlimit_nofilem")
		t.Log(err)
	}
	{
		err := api.Directive().Add(Queries(), nginx.NewDirective("worker_rlimit_nofile", "8192"))
		t.Log(err)
	}
}
```

### 三、使用第三方储统一配置多nginx

本程序提供了`consul k/v`、`etcd k/v`、`zookeeper`三种存储`nginx`配置。

通过启动参数[--storage](./FLAGS.MD)或者[-S](./FLAGS.MD)启用第三方存储。

使用`consul k/v`存储实例。

1、在使用第三方存储，必须首先同步本地配置文件到存储。程序提供了简单的同步方式。

```shell script
aginx sync consul://127.0.0.1:8500/aginx
```
注：运行本命令后，本地的配置文件将被同步到 `consul k/v`、 `/aginx` 目录下。

<p style="color:red"><b>特别注意：如果使用了第三方配置存储后，您的配置文件需要使用相对路径处理include不然程序将无法发现include文件。</b></p>

### 四、自定义存储插件

如果本程序提供的存储方式不满足您的需求，同样本程序也提供了扩展插件，相关章节查阅 [engine.go][engine.go]

### 五、简单的申请免费ssl证书

有关本章节内容，你可以查阅 [restful api](./RESTFULAPI.MD)

### 六、发布docker, consul的服务到nginx

docker发布服务到nginx 详情查阅：[DOCKER.MD](./DOCKER.MD)

consul发布服务到nginx和docker高度兼容（零兼容模式搜索文件为：`<storage>`/registry/consul-labels.conf），可以查阅：[DOCKER.MD](./DOCKER.MD) 了解更多。

### 七、自定义注册中心服务发布到nginx（插件编写）
查阅：[app/plugins/registry/registry.go][registry] 文件定义


# 更多帮助

>   本软件采用[cobra](https://github.com/spf13/cobra)命令行启动方式，如果本章节伟介绍详尽的内容您可以自己查询help命令。

```shell script
$ aginx -h

Usage:
  aginx [flags]
  aginx [command]

Available Commands:
  completion  输出命令帮助
  help        Help about any command
  plugin      显示插件配置的帮助信息
  sync        同步存储配置

Flags:
      --allow-ip strings      api服务允许调用的IP地址 (env: AGINX_ALLOW_IP)  (default [*])
      --auth stringToString   管理台认证用户 (env: AGINX_AUTH)  (default [aginx=aginx])
      --bind string           api服务开放地址 (env: AGINX_BIND)  (default "127.0.0.1:8011")
      --api  string 					连接api节点的地址 （env: AGINX_API） 
  -C, --cert strings          使用 aginx help certs <provider> 查询更新帮助信息 (env: AGINX_CERT)  (default [lego://aginx@renzhen.la/certs/lego,custom://certs/custom])
      --cert-def string       默认cert使用名字 (env: AGINX_CERT_DEF)  (default "lego")
  -f, --conf strings          the global config file path.(env: AGINX_CONF)
      --disable-admin         禁用管理控制台 (env: AGINX_DISABLE_ADMIN) 
      --disable-api           禁用API服务 (env: AGINX_DISABLE_API) 
      --disable-daemon        禁用nginx托管，禁用后将不会托管启动nginx (env: AGINX_DISABLE_DAEMON) 
      --expose string         为API服务暴露一个域名。例如: api.aginx.io 或 api.aginx.io,ssl (env: AGINX_EXPOSE) 
  -h, --help                  help for aginx
      --log-file string       日志输出到文件的位置，默认输出到控制台 (env: AGINX_LOG_FILE) 
  -L, --log-level string      日志级别 (env: AGINX_LOG_LEVEL)  (default "info")
      --nginx string          nginx 可执行程序的位置，默认将自动搜索.如果搜索不到并且未指定将报错 (env: AGINX_NGINX) 
  -P, --plugins string        插件文件夹 (env: AGINX_PLUGINS)  (default "./plugins")
  -R, --registry strings      配置注册管理器 (env: AGINX_REGISTRY) 
  -S, --storage string        集中存储配置方式. 
                                        例如:
                                                consul://127.0.0.1:8500/aginx[?token=authtoken]   consul k/v配置.
                                                zk://127.0.0.1:2182/aginx[?scheme=&auth=]         zookeeper 配置.
                                                etcd://127.0.0.1:2379/aginx[?user=&password]      
                                                file://etc/nginx/nginx.conf                       本机配置
                                 (env: AGINX_STORAGE) 
  -v, --version               version for aginx

```

[registry.go]: https://github.com/ihaiker/aginx/blob/master/app/plugins/registry/registry.go
[engine.go]: https://github.com/ihaiker/aginx/blob/master/app/plugins/storage/engine.go
