package file

import (
	"github.com/ihaiker/aginx/logs"
	"github.com/ihaiker/aginx/util"
	"os"
	"path/filepath"
	"strings"
)

var logger = logs.New("storage", "engine", "file")

type fileStorage struct {
	conf string
}

func (fs *fileStorage) Abs(file string) string {
	if strings.HasPrefix(file, "/") {
		return file
	} else {
		return filepath.Join(filepath.Dir(fs.conf), file)
	}
}

func System() (*fileStorage, error) {
	_, file, err := GetInfo()
	if err != nil {
		return nil, err
	}
	return New(file), nil
}

func New(conf string) *fileStorage {
	return &fileStorage{conf: conf}
}

func (fs *fileStorage) IsCluster() bool {
	return false
}

func (fs *fileStorage) Search(args ...string) ([]*util.NameReader, error) {
	readers := make([]*util.NameReader, 0)
	for _, arg := range args {

		pattern := fs.Abs(arg)
		files, _ := filepath.Glob(pattern)

		for _, f := range files {
			if reader, err := fs.Get(f); os.IsNotExist(err) {
				continue
			} else if err != nil {
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

func (cs *fileStorage) Remove(file string) error {
	fp := filepath.Join(filepath.Dir(cs.conf), file)
	if fileInfo, err := os.Stat(fp); err != nil {
		return err
	} else if fileInfo.IsDir() {
		return os.RemoveAll(fp)
	} else {
		return os.Remove(fp)
	}
}

func (fs *fileStorage) Get(file string) (reader *util.NameReader, err error) {
	path := fs.Abs(file)
	if stat, err := os.Stat(path); err == nil && stat.IsDir() {
		return nil, os.ErrNotExist
	}

	rd, err := os.OpenFile(path, os.O_RDONLY, os.ModeTemporary)
	if err != nil {
		return nil, err
	}
	rel, _ := filepath.Rel(filepath.Dir(fs.conf), path)
	return util.NamedReader(rd, rel), nil
}

func (fs *fileStorage) Put(file string, content []byte) error {
	path := fs.Abs(file)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	_ = os.Remove(path)

	if fio, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0666); err != nil {
		return err
	} else {
		defer func() { _ = fio.Close() }()
		_, _ = fio.Write(content)
		return nil
	}
}
