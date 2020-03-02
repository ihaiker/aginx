# Server 启动参数



## 现有参数

你可以使用 `aginx server -h` 命令直接查阅说明.

| 参数                         | 默认值               | 说明                                                         |
| ---------------------------- | -------------------- | ------------------------------------------------------------ |
| -i, --api                    | 127.0.0.1:8011       | restful api 绑定地址                                         |
| -s, --security               | -                    | 访问restful api的 base auth访问配置. 实例：user:passwd       |
| -u, --email                  | aginx@renzhen.la     | 注册免费ssl证书使用的邮箱账户。                              |
| -e, --expose                 | -                    | 暴露API服务服务使用的域名。例如: api.aginx.io                |
|                              |                      |                                                              |
| -S, --storage                | -                    | 使用第三方存储，存储nginx配置。<br />consul://127.0.0.1:8500/aginx[?token=authtoken]<br />zk://127.0.0.1:2182/aginx[?scheme=&auth=]<br />etcd://127.0.0.1:2379/aginx[?user=&password] |
| --disable-watcher            | False                | 禁用文件变化监听，程序默认开大了程序文件变化，重启`nginx`。并且如果您开启了第三方存储也将自动同步到第三方上。 |
|                              |                      |                                                              |
| --server                     | -                    | 自动添加一个代理配置。此代理配置使用最简单配置方式。<br/>example: --server 'a1.aginx.io=172.0.0.1:8080' --server 'a2.aginx.io=172.0.0.1:8083,127.0.0.1:8084' |
| -c, --conf                   | -                    | 使用配置文件，例如：/etc/nginx/aginx.conf                    |
|                              |                      |                                                              |
|                              |                      |                                                              |
| -C, --consul                 | -                    | Automatically obtain consul registered services and publish them to NGINX. |
| --consul-cacert              |                      | the CA file to use for talking to Consul over TLS.           |
| --consul-capath              |                      | sets the path to a directory of CA certs to use for talking to Consul over TLS. |
| --consul-client-cert         |                      | sets the client cert file to use for talking to Consul over TLS. |
| --consul-client-key          |                      | the client key file to use for talking to Consul over TLS.   |
| --consul-http-addr           |                      | Address is the address of the Consul server.                 |
| --consul-http-auth           |                      | sets the HTTP authentication header.                         |
| --consul-http-ssl            |                      | sets whether or not to use HTTPS.                            |
| --consul-http-ssl-verify     |                      | h sets whether or not to disable certificate checking.       |
| --consul-http-token          |                      | Token is used to provide a per-request ACL token. which overrides the agent's default token. |
| --consul-http-token-file     |                      | defines an environment variable name which sets the HTTP token file. |
| --consul-tls-server-name     |                      | defines an environment variable name which sets the server name to use as the SNI host when connecting via TLS |
| --consul-datacenter          | dc1                  | Datacenter to use                                            |
| --consul-template-mode       | false                | 使用模板模式                                                 |
| --consul-filter              | .*                   | 使用模板模式时，关注的服务过滤器，只有当符合此过滤器的服务改变后才会发送通知。 |
| --consul-labels-template-dir | templates/consul     | labels 模式下，生成配置的模板文件搜索位置。 <br />模板文件搜索优先顺序为  **${domain}.ngx.tpl** 、**default.tpl**、系统默认模板。<br />更详尽的说明查阅[CONSUL.MD](./CONSUL.MD) |
| --consul-template            | templates/consul.tpl | 模板模式的模板， 和 `--consul-template-mode true`配合使用    |
| --consul-template-funcmap    |                      | 添加模板方法                                                 |
|                              |                      |                                                              |
|                              |                      |                                                              |
| -D, --docker                 |                      | 发布配置consul服务到nginx                                    |
| --docker-host                |                      | Set the url to the docker server                             |
| --docker-ip                  |                      | IP for ports mapped to the host                              |
| --docker-api-version         | 1.40                 | Set the version of the API to reach, leave empty for latest (1.40). |
| -docker-cert-path            | -                    | Load the TLS certificates from.                              |
| --docker-tls-verify          | -                    | to enable or disable TLS verification, off by default.       |
| --docker-labels-template-dir | templates/docker     | label模式下模板文件搜索路径。<br />模板文件搜索优先顺序为  **${domain}.ngx.tpl** 、**default.tpl**、系统默认模板。<br />更详尽的说明查阅：[DOCKER.MD](./DOCKER.MD) |
| --docker-template            | templates/docker.tpl | 使用模板模式的模板文件。                                     |
| --docker-template-mode       | False                | 是否使用模板模式。                                           |
| --docker-container-filter    | .*                   | Filtering containers that need attention, see regexp (default [.*]) |
| --docker-service-filter      | .*                   | Filter services that need attention, see regexp (default [.*]) |
| --docker-template-funcmap    |                      | 添加模板方法                                                 |
|                              |                      |                                                              |



## 环境变量

所有配置文件都可以使用系统环境变量来配置。

例如：要配置 --api 和 --security 

```powershell
AGINX_API=127.0.0.1:8012 AGINX_SECURITY=test:test aginx server 
```

环境变量和参数就是 **AGINX_ + toUpler(replace(key,'-','_'))**





## 插件参数



如果您使用了第三方register插件，那么插件的参数依然可以使用 `aginx server -h` 查看或者查阅插件介绍，

此插件参数同样符合环境变量设置规则。此规则为系统设置方式，并非插件功能。

