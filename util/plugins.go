package util

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"plugin"
	"strings"
)

func FindPlugins(module string) map[string]*plugin.Plugin {
	wd, err := os.Getwd()
	PanicIfError(err)

	pluginDir, err := filepath.Abs(filepath.Join(wd, "plugins", module))
	PanicIfError(err)

	storagePlugins := make(map[string]*plugin.Plugin)

	if info, err := os.Stat(pluginDir); os.IsNotExist(err) {
		return storagePlugins
	} else if !info.IsDir() {
		panic(errors.New(pluginDir + " not a plugins folder"))
	} else {
		PanicIfError(err)
	}

	files, err := ioutil.ReadDir(pluginDir)
	PanicIfError(err)

	for _, file := range files {
		pug, err := plugin.Open(filepath.Join(pluginDir, file.Name()))
		if err != nil {
			continue
		}
		name := file.Name()
		if ext := filepath.Ext(name); ext != "" {
			name = strings.Replace(name, "."+ext, "", 1)
		}
		storagePlugins[name] = pug
	}
	return storagePlugins
}
