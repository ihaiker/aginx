package storage

import (
	"github.com/ihaiker/aginx/nginx/configuration"
	"github.com/ihaiker/aginx/util"
)

type Engine interface {

	//搜索文件
	Search(args ...string) ([]*util.NameReader, error)

	//找文件
	File(file string) (*util.NameReader, error)

	//存储文件内容
	Store(file string, content []byte) error

	//存储configuration
	StoreConfiguration(cfg *configuration.Configuration) error
}
