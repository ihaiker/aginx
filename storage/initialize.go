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
				err = errors.New("not support: " + config.Scheme)
			}
		}
		PanicIfError(err)
	}
	return
}
