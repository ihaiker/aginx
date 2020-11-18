package storage

import (
	"fmt"
	"github.com/ihaiker/aginx/v2/core/logs"
	"github.com/ihaiker/aginx/v2/core/storage/consul"
	"github.com/ihaiker/aginx/v2/core/storage/etcd"
	"github.com/ihaiker/aginx/v2/core/storage/file"
	"github.com/ihaiker/aginx/v2/core/storage/zookeeper"
	"github.com/ihaiker/aginx/v2/plugins/storage"
	"net/url"
)

var Plugins = map[string]storage.Plugin{
	"file":   file.LoadStorage(),
	"consul": consul.LoadStorage(),
	"etcd":   etcd.LoadStorage(),
	"zk":     zookeeper.LoadStorage(),
}

func Get(urlConfig string) (storage.Plugin, error) {
	config, err := url.Parse(urlConfig)
	if err != nil {
		return nil, err
	}
	for _, p := range Plugins {
		if p.Scheme() == config.Scheme {
			logs.Infof("存储插件选择：%s (%s) ", p.Name(), p.Scheme())
			err = p.Initialize(*config)
			return p, err
		}
	}
	return nil, fmt.Errorf("未发现存储插件：%s", urlConfig)
}
