package file

import (
	"github.com/ihaiker/aginx/logs"
	"github.com/ihaiker/aginx/nginx"
	"github.com/ihaiker/aginx/plugins"
	"github.com/ihaiker/aginx/util"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var logger = logs.New("storage", "engine", "file")

type fileStorage struct {
	conf        string
	fileWatcher *FileWatcher
}

func (fs *fileStorage) Abs(file string) string {
	if strings.HasPrefix(file, "/") {
		return file
	} else {
		return filepath.Join(filepath.Dir(fs.conf), file)
	}
}

func System() (*fileStorage, error) {
	_, conf, err := nginx.GetInfo()
	if err != nil {
		return nil, err
	}
	return New(conf), nil
}

func MustSystem() *fileStorage {
	engine, err := System()
	util.PanicIfError(err)
	return engine
}

func New(conf string) *fileStorage {
	return &fileStorage{conf: conf}
}

func (fs *fileStorage) IsCluster() bool {
	return false
}

func (fs *fileStorage) Search(args ...string) ([]*plugins.ConfigurationFile, error) {
	if len(args) == 0 {
		return fs.List()
	}
	readers := make([]*plugins.ConfigurationFile, 0)
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
		logger.Info("remove dir ", file)
		return os.RemoveAll(fp)
	} else {
		logger.Info("remove file ", file)
		return os.Remove(fp)
	}
}

func (fs *fileStorage) Get(file string) (reader *plugins.ConfigurationFile, err error) {
	path := fs.Abs(file)
	if stat, err := os.Stat(path); err == nil && stat.IsDir() {
		return nil, os.ErrNotExist
	}

	rd, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	rel, _ := filepath.Rel(filepath.Dir(fs.conf), path)
	return plugins.NewFile(rel, rd), nil
}

func (fs *fileStorage) Put(file string, content []byte) error {
	path := fs.Abs(file)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}
	logger.Info("put file ", file)
	if fio, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666); err != nil {
		return err
	} else {
		defer func() { _ = fio.Close() }()
		_, _ = fio.Write(content)
		return nil
	}
}

func (fs *fileStorage) List() ([]*plugins.ConfigurationFile, error) {
	dir := filepath.Dir(fs.conf)
	return walkFile(dir, "")
}

func (fs *fileStorage) StartListener() <-chan plugins.FileEvent {
	dir := filepath.Dir(fs.conf)
	fs.fileWatcher = NewFileWatcher(dir)
	_ = fs.fileWatcher.Start()
	return fs.fileWatcher.Listener
}

func walkFile(root, appendRelativeDir string) ([]*plugins.ConfigurationFile, error) {
	files := make([]*plugins.ConfigurationFile, 0)
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		if filepath.Ext(path) == ".so" || filepath.Ext(path) == ".dll" {
			return nil
		}

		if info.Mode()&os.ModeSymlink == os.ModeSymlink {
			linkPath, _ := filepath.EvalSymlinks(path)
			if linkInfo, err := os.Stat(linkPath); err != nil {
				return err
			} else if linkInfo.IsDir() {
				relative, _ := filepath.Rel(root, path)
				if dirFiles, err := walkFile(linkPath, relative); err == nil {
					files = append(files, dirFiles...)
				}
				return nil
			}
		}

		file, _ := filepath.Rel(root, path)
		if appendRelativeDir != "" {
			file = filepath.Join(appendRelativeDir, file)
		}
		bs, _ := ioutil.ReadFile(path)

		files = append(files, &plugins.ConfigurationFile{Name: file, Content: bs})
		return nil
	})
	return files, err
}
