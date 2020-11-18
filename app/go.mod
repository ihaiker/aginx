module github.com/ihaiker/aginx/v2

go 1.14

replace github.com/docker/docker => github.com/docker/engine v1.4.2-0.20200204220554-5f6d6f3f2203

require (
	github.com/Microsoft/go-winio v0.4.14 // indirect
	github.com/alecthomas/participle v0.6.0
	github.com/containerd/containerd v1.4.1 // indirect
	github.com/coreos/etcd v3.3.13+incompatible
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v0.0.0-00010101000000-000000000000
	github.com/docker/go-connections v0.4.0
	github.com/docker/go-units v0.4.0 // indirect
	github.com/go-acme/lego/v3 v3.9.0
	github.com/go-ping/ping v0.0.0-20201020221959-265e7c64b33b
	github.com/hashicorp/consul/api v1.7.0
	github.com/ihaiker/cobrax v1.2.2
	github.com/kataras/iris/v12 v12.1.8
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/rs/cors v1.7.0
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/cobra v1.1.1
	github.com/tencentcloud/tencentcloud-sdk-go v1.0.50
)
