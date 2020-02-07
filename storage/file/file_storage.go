package file

import (
	"github.com/ihaiker/aginx/nginx/configuration"
	"github.com/ihaiker/aginx/util"
	"os"
	"path/filepath"
	"strings"
)

type FileStorage struct {
	conf string
}

func (fs *FileStorage) Abs(file string) string {
	if strings.HasPrefix(file, "/") {
		return file
	} else {
		path, _ := filepath.Abs(filepath.Dir(fs.conf) + "/" + file)
		return path
	}
}

func System() (*FileStorage, error) {
	_, file, err := GetInfo()
	if err != nil {
		return nil, err
	}
	return New(file), nil
}

func New(conf string) *FileStorage {
	return &FileStorage{conf: conf}
}

func (fs *FileStorage) Search(args ...string) ([]*util.NameReader, error) {
	readers := make([]*util.NameReader, 0)
	for _, arg := range args {

		pattern := fs.Abs(arg)
		files, _ := filepath.Glob(pattern)

		for _, f := range files {
			if reader, err := fs.File(f); err != nil {
				return nil, err
			} else {
				if !strings.HasPrefix(arg, "/") {
					prefix := strings.Replace(pattern, arg, "", 1)
					reader.Name = strings.Replace(reader.Name, prefix, "", 1)
				}
				readers = append(readers, reader)
			}
		}
	}
	return readers, nil
}

func (fs *FileStorage) File(file string) (reader *util.NameReader, err error) {
	path := fs.Abs(file)
	rd, err := os.OpenFile(path, os.O_RDONLY, os.ModeTemporary)
	if err != nil {
		if os.IsNotExist(err) {
			dir := filepath.Dir(fs.conf)
			return fs.File(dir + "/" + file)
		}
		return nil, err
	}
	return util.NamedReader(rd, path), nil
}

func (fs *FileStorage) Store(file string, content []byte) error {
	path := fs.Abs(file)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, os.ModeDir); err != nil {
		return err
	}
	if err := os.Remove(path); err != nil {
		return err
	}
	if fio, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0666); err != nil {
		return err
	} else {
		defer func() { _ = fio.Close() }()
		_, _ = fio.Write(content)
		return nil
	}
}

func (fs *FileStorage) StoreConfiguration(cfg *configuration.Configuration) error {
	return configuration.Down(filepath.Dir(fs.conf), cfg)
}
