package file

import (
	"fmt"
	"github.com/ihaiker/aginx/v2/core/logs"
	"github.com/ihaiker/aginx/v2/core/util/errors"
	"github.com/ihaiker/aginx/v2/core/util/files"
	"github.com/ihaiker/aginx/v2/plugins/storage"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

var logger = logs.New("storage", "engine", "file")

type fileStorage struct {
	nginxConfig string
}

func LoadStorage() storage.Plugin {
	return &fileStorage{}
}
func (l *fileStorage) Scheme() string {
	return "file"
}

func (l *fileStorage) Name() string {
	return "本地文件存储"
}

func (l *fileStorage) Version() string {
	return "v2.0.0"
}

func (l *fileStorage) Help() string {
	return `本地存储，也是默认的nginx配置。
默认情况下系统会根据nginx自动查找配置位置并使用。
如果需要自定义。格式为：file://<nginx.conf路径> 例如：file://etc/nginx/nginx.conf.
需要注意:如果未提供nginx可执行程序并不能查找配置位置。`
}

func (l *fileStorage) Initialize(config url.URL) error {
	nginxConf := filepath.Join(config.Host, config.Path) //fixbug: 文件路径错误问题
	// 如果用户提供了 nginx.conf 全路径就检查文件是否存在，
	if files.IsDir(nginxConf) {
		nginxConf = filepath.Join(nginxConf, "nginx.conf")
	} else if !files.Exists(nginxConf) {
		return fmt.Errorf("%s not found", nginxConf)
	}
	l.nginxConfig = nginxConf
	return nil
}

//获取include文件的相对路径
func (fs *fileStorage) rel(path string) string {
	dir := filepath.Dir(fs.nginxConfig)
	if strings.HasPrefix(path, dir) {
		if rel, err := filepath.Rel(dir, path); err == nil {
			return rel
		}
	}
	return path
}

//获取文件的绝对路径
func (fs *fileStorage) abs(file string) string {
	//如果是绝对路径直接返回
	if strings.HasPrefix(file, "/") {
		return file
	}
	return filepath.Join(filepath.Dir(fs.nginxConfig), file)
}

func (fs *fileStorage) list() ([]*storage.File, error) {
	dir := filepath.Dir(fs.nginxConfig)
	fis, err := files.Walk(dir)
	for _, f := range fis {
		f.Name = fs.rel(f.Name)
	}
	return fis, err
}

func (fs *fileStorage) Search(args ...string) ([]*storage.File, error) {
	if len(args) == 0 {
		return fs.list()
	}

	matchers := make([]*storage.File, 0)
	for _, arg := range args {
		pattern := fs.abs(arg)
		fis, _ := filepath.Glob(pattern)
		for _, fName := range fis {
			if f, err := fs.Get(fName); err != nil {
				if err == errors.ErrNotFound {
					continue
				}
				return nil, err
			} else {
				matchers = append(matchers, f)
			}
		}
	}
	return matchers, nil
}

func (cs *fileStorage) Remove(file string) error {
	fp := cs.abs(file)
	if !files.Exists(fp) {
		return errors.ErrNotFound
	}
	logger.Info("remove ", file)
	return os.RemoveAll(fp)
}

func (fs *fileStorage) Get(file string) (reader *storage.File, err error) {
	path := fs.abs(file)
	if !files.Exists(path) || files.IsDir(path) {
		return nil, errors.Wrap(errors.ErrNotFound, "文件未发现")
	}
	rd, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return storage.NewFile(fs.rel(path), rd), nil
}

func (fs *fileStorage) Put(file string, content []byte) error {
	path := fs.abs(file)
	err := files.WriteFile(path, content)
	if err != nil {
		logger.Debug("new file: ", file)
	}
	return err
}

func (fs *fileStorage) Listener() <-chan storage.FileEvent {
	return nil
}
