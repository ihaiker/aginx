package nginx

import (
	"github.com/ihaiker/aginx/nginx/config"
	"github.com/ihaiker/aginx/plugins"
	"github.com/ihaiker/aginx/util"
	"path/filepath"
	"strings"
)

func Readable(store plugins.StorageEngine) (*config.Configuration, error) {
	reader, err := store.Get("nginx.conf")
	if err != nil {
		return nil, util.Wrap(err, "get nginx.conf")
	}
	return ReaderReadable(store, reader)
}

func ReaderReadable(store plugins.StorageEngine, cfgFile *plugins.ConfigurationFile) (cfg *config.Configuration, err error) {
	if cfg, err = config.ParseWith(cfgFile.Name, cfgFile.Content); err == nil {
		err = virtual(store, cfg)
	}
	return
}

func includes(store plugins.StorageEngine, node *config.Directive) error {
	files, err := store.Search(node.Args...)
	if err != nil {
		return err
	}
	for _, file := range files {
		includeDirective := &config.Directive{Virtual: config.Include, Name: "file", Args: Queries(file.Name)}
		if doc, err := ReaderReadable(store, file); err != nil {
			return err
		} else {
			includeDirective.Body = doc.Body
		}
		node.Body = append(node.Body, includeDirective)
	}
	return nil
}

func virtual(store plugins.StorageEngine, directive *config.Directive) (err error) {
	if directive.Name == "include" {
		configDir := MustConfigDir()
		for i, arg := range directive.Args {
			if strings.HasPrefix(arg, configDir) {
				directive.Args[i], _ = filepath.Rel(configDir, arg)
			}
		}
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
