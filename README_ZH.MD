# Aginx 

为 [NGINX][NGINX] 添加管理API的服务

## 简介

在项目中

## 安装 

### 源码安装
### 二进制安装
### Docker 安装

下载

## Features 功能列表

[X] restful api for nginx.
[X] cluster config use consul key/value.
[X] cluster config use etcd key/value.
[X] cluster config use zookeeper.

## 命令介绍

```shell script
api for nginx. 
Build: 2020-02-13 16:20:41, Go: go1.13.7, Commit: e427d7ca051f88719acbeaff982344a26a1e4602

Usage:
  aginx [flags]
  aginx [command]

Available Commands:
  cluster     Sync configuration files from nginx to cluster storage
  help        Help about any command
  server      the aginx server

Flags:
  -a, --api string        restful api port. (default :8081)
  -c, --cluster string    cluster config
                          for example. 
                                consul://127.0.0.1:8500/aginx[?token=authtoken]   config from consul.  
                                zk://127.0.0.1:2182/aginx[?scheme=&auth=]         config from zookeeper.
                                etcd://127.0.0.1:1234/aginx[?user=&password]      config from etcd.
                          
  -d, --debug             debug mode
  -e, --expose string     expose api use domain
  -h, --help              help for aginx
  -l, --level string      log level (default "info")
  -s, --security string   base auth for restful api, example: user:passwd
      --version           version for aginx

```

[NGINX]: http://nginx.org
[DEMO]: https://www.airpair.com/scalable-architecture-with-docker-consul-and-nginx
[DOWNLOAD]: https://api.github.com/repos/ihaiker/aginx/releases/latest