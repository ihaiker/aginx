package configuration

import (
	"github.com/ihaiker/aginx/nginx"
	"github.com/ihaiker/aginx/util"
)

type Writer func(file string, content []byte) error
type Differ func(file string, content []byte) bool

func Down(root string, cfg *nginx.Configuration) error {
	return DownWriter(root, cfg, util.WriteFile)
}

func DownWriter(root string, cfg *nginx.Configuration, writerFn Writer) (err error) {
	content := cfg.String()
	if err = writerFn(root+"/nginx.conf", []byte(content)); err != nil {
		return
	}
	//if err = writeVirtual(root, cfg.Directive(), writerFn, func(file string, content []byte) bool {
	//	return false
	//}); err != nil {
	//	return
	//}
	return
}
