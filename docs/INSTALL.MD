# 安装

## 源码安装

本软件采用golang语言开发，在源码安装本软件之前需要先行安装golang编译环境。

```shell script
git clone https://github.com/ihaiker/aginx
cd aginx/app
go mod download
go build -o aginx aginx.go
```
执行上述命令后您可以在当前文件夹下找到编译结果`aginx`文件。

## 二进制安装

下载地址：https://github.com/ihaiker/aginx/releases

## Docker 安装

```shell script
docker pull ihaiker/aginx:latest
```

# [使用文档](./USAGE.MD)
