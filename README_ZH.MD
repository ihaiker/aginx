# Aginx 

为 [NGINX][NGINX] 添加WEB管理和API管理的服务，
[![Try in PWD][PWDIMAGE]][PWD] 并打开8011端口查看

## 支持特征 

- [X] 提供 web console 管理 NGINX。
- [X] 提供 restful api SDK管理NGINX
- [X] 获取免费ssl证书和配置
- [X] 使用`consul k/v`、`etcd k/v`、`zookeeper`存储 NGINX 配置文件，多NGINX可以统一配置和管理。
- [X] 发布 `docker`、`consul` 服务到 nginx 配置。

## [安装文档 点击查阅](./docs/INSTALL.MD)

## [使用文档 点击查阅](./docs/USAGE.MD)

## 贡献

- 如果遇到错误，可以将其直接提交给dev分支。
- 如果遇到问题，可以通过问题反馈。
- 该项目正在开发中，仍有很大的改进空间。如果您可以贡献代码，请向PR分支提交PR。
- 如果有关于新功能的反馈，您可以通过问题或qq组、dingtalk组反馈。

## 技术支持群

![dingtalk](./docs/dingtalk2.png)
![qq](./docs/qq1.png)

[NGINX]: http://nginx.org
[PWDIMAGE]: https://raw.githubusercontent.com/play-with-docker/stacks/master/assets/images/button.png "Play With Docker Image"
[PWD]: https://labs.play-with-docker.com/?stack=https://raw.githubusercontent.com/ihaiker/aginx/master/demo/docker-compose.yml "Docker Stack YAML"
