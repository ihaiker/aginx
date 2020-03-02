# 使用介绍

前言：本软件采用[cobra](https://github.com/spf13/cobra)命令行启动方式，如果本章节伟介绍详尽的内容您可以自己查询help命令。


```shell script
$ aginx -h

Usage:
  aginx [command]

Available Commands:
  help        Help about any command
  registry    the AGINX registry server
  server      the AGINX server
  sync        Sync configuration files from nginx to cluster storage

Flags:
  -d, --debug          debug mode
  -h, --help           help for aginx
  -l, --level string   log level (default "info")
      --version        version for aginx

Use "aginx [command] --help" for more information about a command.
```

## Aginx服务

```shell
aginx server
```

此命令提供api管理服务，运行此服务需要注意`nginx`命令可以找到，不然会报错，程序会托管`nginx`一起启动，并不需要单独启动。
如果你已经启动了`nginx`此程序会尝试先行关闭它并且托管启动。

[FLAGS.MD](./FLAGS.MD)可以查看全部参数和配置。

*测试restful api*

```shell script
curl http://128.0.0.1:8011/api
```
执行成功说明服务启动成功。

### 特性说明

#### 一、restful api

Restful api 提供了管理和配置`nginx`的命令。有关全部restful api 内容查阅相关章节：[restful api](./RESTFULAPI.MD)



#### 二、sdk for aginx

本程序提供了`sdk`，所使用此SDK你可以自己开发配置管理nginx的服务，当然使用sdk您需要启动`aginx`主服务，此sdk只是调用`restful api`。

下载sdk
```shell script
go get github.com/ihaiker/aginx/api
```

使用sdk
```go
package test

import (
	"github.com/ihaiker/aginx/api"
	"testing"
)

func TestAginx_Add(t *testing.T) {
    api := New("http://127.0.0.1:8011")
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

#### 三、使用第三方储统一配置多nginx

本程序提供了`consul k/v`、`etcd k/v`、`zookeeper`三种存储`nginx`配置。

通过启动参数[--storage](./FLAGS.MD)或者[-S](./FLAGS.MD)启用第三方存储。

使用`consul k/v`存储实例。

1、在使用第三方存储，必须首先同步本地配置文件到存储。程序提供了简单的同步方式。
```shell script
aginx sync consul://127.0.0.1:8500/aginx
```
注：运行本命令后，本地的配置文件将被同步到 `consul k/v`、 `/aginx` 目录下。

<p style="color:red"><b>特别注意：如果使用了第三方配置存储后，您的配置文件需要使用相对路径处理include不然程序将无法发现include文件。</b></p>

#### 四、第三方存储插件

如果本程序提供的存储方式不满足您的需求，同样本程序也提供了扩展插件，相关章节查阅 [存储插件开发](./plugins/STORAGE.MD)



#### 五、简单的申请免费ssl证书

有关本章节内容，你可以查阅 [restful api](./RESTFULAPI.MD)



#### 六、发布docker的服务到nginx

详情查阅 [DOCKER.MD](./DOCKER.MD)



#### 七、发布consul的服务到nginx

详情查阅 [CONSUL.MD](./CONSUL.MD)



#### 八、其他注册中心服务发布到nginx插件

详情查阅：[REGISTER.MD](./plugins/REGISTER.MD)

