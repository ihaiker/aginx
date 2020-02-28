package storage

import (
	"errors"
	"github.com/ihaiker/aginx/plugins"
	"github.com/ihaiker/aginx/storage/consul"
	"github.com/ihaiker/aginx/storage/etcd"
	"github.com/ihaiker/aginx/storage/file"
	"github.com/ihaiker/aginx/storage/zookeeper"
	. "github.com/ihaiker/aginx/util"
	"net/url"
)

func FindStorage(cluster string) (storage plugins.StorageEngine) {
	if cluster == "" {
		storage = file.MustSystem()
	} else {
		config, err := url.Parse(cluster)
		if err == nil {
			switch config.Scheme {
			case "consul":
				storage, err = consul.New(config)
			case "etcd":
				storage, err = etcd.New(config)
			case "zk":
				storage, err = zookeeper.New(config)
			default:
				storagePlugins := FindPlugins("storage")
				if storagePlugin, has := storagePlugins[config.Scheme]; has {
					if fn, err := storagePlugin.Lookup(plugins.PLUGIN_INIT_METHOD_NAME); err == nil {
						if storagePluginMethod, match := fn.(plugins.StoragePluginMethod); match {
							storage, err = storagePluginMethod(config)
						}
					}
				}
				if storage == nil {
					err = errors.New("storage plugin not support: " + cluster)
				}
			}
		}
		PanicIfError(err)
	}
	return
}
