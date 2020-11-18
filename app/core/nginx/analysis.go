package nginx

import (
	"github.com/ihaiker/aginx/v2/core/nginx/config"
	"github.com/ihaiker/aginx/v2/core/util/errors"
	"github.com/ihaiker/aginx/v2/plugins/storage"
)

//从存储中读取配置
func Configuration(store storage.Plugin) (*config.Configuration, error) {
	reader, err := store.Get("nginx.conf")
	if err != nil {
		return nil, errors.Wrap(err, "get nginx.conf")
	}
	return parseConfig(store, reader)
}

func parseConfig(store storage.Plugin, cfgFile *storage.File) (cfg *config.Configuration, err error) {
	if cfg, err = config.ParseWith(cfgFile.Name, cfgFile.Content, nil); err == nil {
		err = virtual(store, cfg)
	}
	return
}

//处理 include 文件
func includes(store storage.Plugin, node *config.Directive) error {
	files, err := store.Search(node.Args...)
	if err != nil {
		return err
	}
	for _, file := range files {
		includeDirective := &config.Directive{Virtual: config.Include, Name: "file", Args: []string{file.Name}}
		if doc, err := parseConfig(store, file); err != nil {
			return err
		} else {
			includeDirective.Body = doc.Body
		}
		node.Body = append(node.Body, includeDirective)
	}
	return nil
}

//处理虚拟文件
func virtual(store storage.Plugin, directive *config.Directive) (err error) {
	if directive.Name == "include" {
		if err = includes(store, directive); err != nil {
			return
		}
	}
	if directive.Body != nil {
		for _, d := range directive.Body {
			if err = virtual(store, d); err != nil {
				return
			}
		}
	}
	return err
}
