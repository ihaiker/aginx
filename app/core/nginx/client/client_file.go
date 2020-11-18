package client

import (
	"github.com/ihaiker/aginx/v2/core/nginx"
	"github.com/ihaiker/aginx/v2/core/util/files"
	"github.com/ihaiker/aginx/v2/plugins/storage"
	"io/ioutil"
	"os"
	"path/filepath"
)

type clientFile struct {
	engine storage.Plugin
	daemon nginx.Daemon
}

func (f *clientFile) New(relativePath, localFilePath string) error {
	if content, err := ioutil.ReadFile(localFilePath); err != nil {
		return err
	} else {
		return f.NewWithContent(relativePath, content)
	}
}

func (f *clientFile) NewWithContent(relativePath string, content []byte) error {
	if err := f.daemon.Test(nil, func(testDir string) error {
		path := filepath.Join(testDir, relativePath)
		dir := filepath.Dir(path)
		if !files.Exists(dir) {
			if err := os.MkdirAll(dir, 0777); err != nil {
				return err
			}
		}
		return ioutil.WriteFile(path, content, 0666)
	}); err != nil {
		return err
	}
	err := f.engine.Put(relativePath, content)
	if err == nil {
		return f.daemon.Reload()
	}
	return err
}

func (f *clientFile) Remove(relativePath string) error {
	//测试文件是否正确
	if err := f.daemon.Test(nil, func(testDir string) error {
		path := filepath.Join(testDir, relativePath)
		return os.RemoveAll(path)
	}); err != nil {
		return err
	}

	err := f.engine.Remove(relativePath)
	if err == nil {
		return f.daemon.Reload()
	}
	return err
}

func (f *clientFile) Search(relativePaths ...string) ([]*storage.File, error) {
	return f.engine.Search(relativePaths...)
}

func (f *clientFile) Get(relativePath string) (*storage.File, error) {
	return f.engine.Get(relativePath)
}
