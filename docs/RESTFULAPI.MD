# restful api

>   为了跟好的解释restful api 的使用方式，我们提供了一个简单的nginx 配置这样可以更好的说明API

## NGINX 配置文件

###  **/etc/nginx/nginx.conf**

```nginx
worker_processes auto;
pid logs/nginx.pid;
events  {
    worker_connections 1024;
}

http  {
    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
    keepalive_timeout 65;
    types_hash_max_size 2048;
    include mime.types;
    default_type application/octet-stream;
 
    server  {
        listen 80;
        server_name portainer.aginx.io;
        location / {
            proxy_pass http://127.0.0.1:9000;
            proxy_set_header Host api.aginx.io;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        }
    }
  
  	include conf.d/*.conf;
    include hosts.d/*.conf;
}
```

### **/etc/nginx/conf.d/default.conf**

```nginx
server  {
    listen 80;
    server_name localhost;
    location / {
        root html;
    }
}
```

### **/etc/nginx/hosts.d/api.conf**

```nginx
upstream api_aginx_io {
    server 127.0.0.1:8011;
}
server  {
    listen 80;
    server_name api.aginx.io;
    location / {
        proxy_pass http://api_aginx_io;
        proxy_set_header Host api.aginx.io;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

### **/etc/nginx/hosts.d/b.conf**

```nginx
upstream a1_aginx_io {
    server 127.0.0.1:8011;
    server 127.0.0.1:8012;
}
server  {
    listen 80;
    server_name a1.aginx.io;
    location / {
        proxy_pass http://a1_aginx_io;
        proxy_set_header Host a1.aginx.io;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```



## API

### Directive API (指令API)

#### 查询

**地址 :** `GET /api/directive`

**参数：** 本API只包含一个参数`q`，此参数为字符数组。q参数详细用法查看 **定位参数语法**

实例：

```shell
curl 'http://127.0.0.1:8011/api/directive?q=http&q=server.server_name('portainer.aginx.io')
```

查询正确 **http status = 200** 返回：

```json
[
    {
        "name": "server",
        "body": [
            {
                "name": "listen",
                "args": [
                    "80"
                ]
            },
            {
                "name": "server_name",
                "args": [
                    "portainer.aginx.io"
                ]
            },
            {
                "name": "location",
                "args": [
                    "/"
                ],
                "body": [
                    {
                        "name": "proxy_pass",
                        "args": [
                            "http://127.0.0.1:9000"
                        ]
                    },
                    {
                        "name": "proxy_set_header",
                        "args": [
                            "Host",
                            "api.aginx.io"
                        ]
                    },
                    {
                        "name": "proxy_set_header",
                        "args": [
                            "X-Real-IP",
                            "$remote_addr"
                        ]
                    },
                    {
                        "name": "proxy_set_header",
                        "args": [
                            "X-Forwarded-For",
                            "$proxy_add_x_forwarded_for"
                        ]
                    }
                ]
            }
        ]
    }
]
```

如果查询的配置内容不存在 **http status = 500 ** 并且返回

```json
{
    "error": "InternalServerError", //错误标识
    "message": "file does not exist" //错误描述
}
```



#### 添加

**地址** : `PUT /api`

**参数：** 此API包含一个定位参数`q`，并把配置内容放入body中，此body的内容就是nginx的配置内容。

实例：添加一个配置

```shell
curl -X PUT 'http://127.0.0.1:8011/api/directive?q=http' \
--data-raw 'server  {
        listen 80;
        server_name ts.renzhen.la;
        location / {
            index index.html;
            root html;
        }
    }'
```

添加成功后台服务端 **http status = 204**

如果查询的配置内容不存在 **http status = 500 ** 并且返回

```json
{
    "error": "错误标识",
    "message": "错误描述"
}
```

当执行完添加命令后，配置文件**/etc/nginx/nginx.conf** 将会是一下结果

```nginx
worker_processes auto;
pid logs/nginx.pid;
events  {
    worker_connections 1024;
}

http  {
    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
    keepalive_timeout 65;
    types_hash_max_size 2048;
    include mime.types;
    default_type application/octet-stream;
 
    server  {
        listen 80;
        server_name portainer.aginx.io;
        location / {
            proxy_pass http://127.0.0.1:9000;
            proxy_set_header Host api.aginx.io;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        }
    }
  
  	include conf.d/*.conf;
    include hosts.d/*.conf;
    
    server  {
        listen 80;
        server_name ts.renzhen.la;
        location / {
            index index.html;
            root html;
        }
    }
}
```



#### 删除

**地址** : `DELETE /api/directive`

**参数：** 此API包含一个定位参数`q`。

```shell
curl -X DELETE 'http://127.0.0.1:8011/api?q=http&q=server.server_name('ts.renzhen.la')
```

此命令将删除了上面添加的配置。



#### 更新

**地址** : `POST /api/directive`

**参数：** 此方法的参数和**添加API**相同。运行结果也是一样的。

实例：修改之前通过添加的ts.renzhen.la服务，禁用外部访问

```shell
curl --location --request POST 'http://127.0.0.1:8011/api?q=http&q=server&q=server_name('ts.renzhen.la')' \
--data-raw 'server_name test.renzhen.la;'
```



####  定位参数详解

看完上面几个简单的示例，您一定对定位参数`q`有些疑惑，他到底值如何工作的。此小结我们将来介绍一下他的工作原理。

在解释参数`q`之前我能先来了解另外一个内容**NGINX配置**。

在对nginx配置分析后此软件将解析文件为：一个树形结构。并且这个树形结果的每个节点都是相同的结构。

分析的结果：

```nginx
directive [arg0 arg1 arg3 ... argN];
directive [arg0 arg1 ... argN ;] [{
      directive arg0 arg1 ... argN;
      directive arg0 arg1 ... argN;
}]
```

| 节点内容    | 说明                                                        |
| ----------- | ----------------------------------------------------------- |
| directive   | 指令，例如: http,server,server_name,tcp_nopush              |
| arg0...argN | 指令参数：例如：**listen 80**中的80，**sendfile on** 中的on |
| [{...}]     | 指令的下级指令。                                            |

**定位参数** 语法：

```javascript
[comparison]directive(comparison'arg1'[ operator comparison'arg2']+)[.[comparison]directive(comparison'arg1'[ operator comparison'arg2']+)]+
```

语法内容：

| 语法指令    | 可用值                         | 语法内容说明                                                 |
| ----------- | ------------------------------ | ------------------------------------------------------------ |
| comparison  | 指令或者参数的比较方式。       | 比较方式供五种<br />空：相等<br />! : 不相等<br />@: 包含<br />^: hasPrefix<br />$: hasSubffix |
| directive   | 指令名称                       | http,server等nginx的指令                                     |
| arg0...argN | 参数名称                       | nginx指令的参数                                              |
| operator    | 参数匹配的结果级的合并判断方式 | ”&“ 并且  ，”\|“ 或者                                        |

估计您看到这里这个语法会有些晦涩难懂，那我们来拆解一下这个语法。

上面的语法其实是 **query.subQuery.subQuery1** 这个query就是语法的一个最小单元：

```json
[comparison]directive(comparison'arg1'[ operator comparison'arg2']+)
```

那么看懂了这个语法，接下来给出一些简单的实例让你理解更佳深入一些。



**1、查询 http下面所有server 配置 参数如下，**

```json
q=http
q=server
```

**2、查询http下面 server_name 为 portainer.aginx.io的server** 

```
q=http
q=server.server_name('portainer.aginx.io')
```

**3、查询http下面 server_name 为 portainer.aginx.io 并且 使用了ssl 的 server** 

```
q=http
q=server.[server_name('portainer.aginx.io') & listen('80')]
```

**3、查询 http server下server_name 以www开头的server** 

```json
q=http
q=server.server_name(^'www')
```

**4、查询 api.aginx.io 配置内容**

```json
q=http
q=include
q=file('hosts.d/api.conf')
q=server.server_name('hosts.d/api.conf')
```

**5、查询所有的server**

```json
q=http
q=server
```

如果使用include使用下面的。

```json
q=http
q=include
q=*
q=server
```



###  文件API

#### 上传一个文件

 地址: `POST /api/file`

| 参数 | 说明                          |
| ---- | ----------------------------- |
| path | 文件路径.此路径必须为相对路径 |
| file | 文件                          |

#### 删除一个文件

地址：`DELETE /api/file?path=fileapth`

删除成功 **http status = 204**

#### 获取文件

地址：`GET /api/file/search?q=fileapth1&q=filepath2`

返回内容：

```json
{
  "filepath":"file_content1",
  "filepath2":"file_content2"
}
```



更多restapi使用帮助，您可以查看postman公开的文件定义。