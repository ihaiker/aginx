package storage

import (
	"github.com/ihaiker/aginx/util"
)

type Engine interface {
	IsCluster() bool

	//存储文件内容
	Put(file string, content []byte) error

	Remove(file string) error

	//搜索文件
	Search(args ...string) ([]*util.NameReader, error)

	//获取文件
	Get(file string) (*util.NameReader, error)
}
