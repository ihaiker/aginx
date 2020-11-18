package util

import (
	"github.com/ihaiker/aginx/v2/core/util/files"
	"io/ioutil"
	"path/filepath"
	"plugin"
)

func FindPlugins(path string) (map[string]*plugin.Plugin, error) {
	storagePlugins := make(map[string]*plugin.Plugin)
	if !files.Exists(path) {
		return storagePlugins, nil
	}
	fileInfos, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, file := range fileInfos {
		if file.IsDir() {
			continue
		}
		pug, err := plugin.Open(filepath.Join(path, file.Name()))
		if err != nil {
			return nil, err
		}
		name := file.Name()
		storagePlugins[name] = pug
	}
	return storagePlugins, nil
}
