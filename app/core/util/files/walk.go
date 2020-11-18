package files

import (
	"fmt"
	"github.com/ihaiker/aginx/v2/plugins/storage"
	"io/ioutil"
	"os"
	"path/filepath"
)

func Walk(root string) ([]*storage.File, error) {
	rootFi, err := os.Stat(root)
	if os.IsNotExist(err) {
		return nil, err
	}
	if !rootFi.IsDir() {
		return nil, fmt.Errorf("%s not dir", root)
	}

	dir, err := ioutil.ReadDir(root) //查看文件加下文件
	if err != nil {
		return nil, err
	}

	files := make([]*storage.File, 0)
	for _, fi := range dir {
		path := filepath.Join(root, fi.Name())
		fi, _ = os.Stat(path)
		if fi.IsDir() {
			if dirFiles, err := Walk(filepath.Join(root, fi.Name())); err != nil {
				return nil, err
			} else {
				files = append(files, dirFiles...)
			}
		} else if filepath.Ext(fi.Name()) == ".so" || filepath.Ext(fi.Name()) == ".dll" {
			//
		} else {
			bs, err := ioutil.ReadFile(path)
			if err != nil {
				return nil, err
			}
			files = append(files, storage.NewFile(path, bs))
		}
	}
	return files, err
}
