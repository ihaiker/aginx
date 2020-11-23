package zookeeper

import (
	"bytes"
	"github.com/ihaiker/aginx/v2/core/logs"
	"github.com/ihaiker/aginx/v2/core/util/errors"
	"github.com/ihaiker/aginx/v2/plugins/storage"
	"github.com/samuel/go-zookeeper/zk"
	"net/url"
	"path/filepath"
	"time"
)

var logger = logs.New("storage", "engine", "zk")
var zkDir = []byte("zkdir")

type zkStorage struct {
	folder string
	keeper *zk.Conn
}

func LoadStorage() *zkStorage {
	return &zkStorage{}
}

func (zks *zkStorage) Scheme() string {
	return "zk"
}

func (zks *zkStorage) Name() string {
	return "Zookeeper 管理器"
}

func (zks *zkStorage) Version() string {
	return "v2.0.0"
}

func (zks *zkStorage) Help() string {
	return `Zookeeper 存储管理器。
配置格式为：zk://host:port?param=value
可选参数说明：
	参数              说明
	scheme          认证信息scheme
	auth            认证信息值
`
}

func (zks *zkStorage) Initialize(config url.URL) (err error) {
	address := config.Host
	folder := config.EscapedPath()
	scheme := config.Query().Get("scheme")
	auth := config.Query().Get("auth")

	zks.folder = folder
	zks.keeper, _, err = zk.Connect([]string{address}, time.Second*3)
	zks.keeper.SetLogger(logger)
	if scheme != "" {
		if err = zks.keeper.AddAuth(scheme, []byte(auth)); err != nil {
			return
		}
	}
	if zks.folder != "" {
		err = zks.zkMkdir(zks.folder)
	}
	return
}

func (zks *zkStorage) zkMkdir(dir string) error {
	exists, _, err := zks.keeper.Exists(dir)
	if err == zk.ErrNoNode || !exists {
		if err = zks.zkMkdir(filepath.Dir(dir)); err != nil {
			return err
		}
		logger.Debugf("创建文件夹: %s", dir)
		if _, err = zks.keeper.Create(dir, zkDir, 0, zk.WorldACL(zk.PermAll)); err != nil {
			return err
		}
	}
	return err
}

func (zks *zkStorage) Listener() <-chan storage.FileEvent {
	panic("implement me")
}

func (zks *zkStorage) Put(file string, content []byte) error {
	path := filepath.Join(zks.folder, file)

	if err := zks.zkMkdir(filepath.Dir(path)); err != nil {
		return err
	}
	if exists, stat, err := zks.keeper.Exists(path); err != nil {
		return err
	} else if exists {
		_, err := zks.keeper.Set(path, content, stat.Version)
		logger.Debug("存储文件 ", file)
		return err
	} else {
		_, err := zks.keeper.Create(path, content, 0, zk.WorldACL(zk.PermAll))
		logger.Debug("存储文件 ", file)
		return err
	}
}

func (zks *zkStorage) Remove(file string) error {
	path := zks.folder + "/" + file
	err := zks.keeper.Delete(path, -1)
	if err == zk.ErrNotEmpty {
		childless, _, _ := zks.keeper.Children(path)
		for _, children := range childless {
			if err = zks.Remove(file + "/" + children); err != nil {
				return err
			}
		}
		return zks.keeper.Delete(path, -1)
	}
	logger.Debug("删除文件 ", file)
	return err
}

func isDir(data []byte) bool {
	return bytes.Equal(data, zkDir) || len(data) == 0
}

func (zks *zkStorage) zkList(path string, onlyFile bool) ([]*storage.File, error) {
	readers := make([]*storage.File, 0)
	keys, _, err := zks.keeper.Children(path)
	if err != nil {
		return nil, err
	}
	for _, key := range keys {
		file := path + "/" + key
		if data, st, err := zks.keeper.Get(file); err != nil {
			return nil, err
		} else {
			if isDir(data) || st.NumChildren > 0 {
				if !onlyFile {
					readers = append(readers, storage.NewFile(file, nil))
				}
				if files, err := zks.zkList(file, onlyFile); err != nil {
					return nil, err
				} else {
					readers = append(readers, files...)
				}
			} else {
				relPath, _ := filepath.Rel(zks.folder, file)
				readers = append(readers, storage.NewFile(relPath, data))
			}
		}
	}
	return readers, err
}

func (zks *zkStorage) Search(pattern ...string) ([]*storage.File, error) {
	matched := make([]*storage.File, 0)

	zkFiles, err := zks.zkList(zks.folder, true)
	if err == nil {
	READERS:
		for _, zkFile := range zkFiles {
			if len(pattern) == 0 {
				matched = append(matched, &storage.File{
					Name:    zkFile.Name,
					Content: zkFile.Content,
				})
			} else {
				for _, arg := range pattern {
					if match, _ := filepath.Match(arg, zkFile.Name); match {
						matched = append(matched, zkFile)
						continue READERS
					}
				}
			}
		}
	}
	return matched, err
}

func (zks *zkStorage) Get(file string) (*storage.File, error) {
	path := filepath.Join(zks.folder, file)
	if data, _, err := zks.keeper.Get(path); err != nil {
		if err.Error() == "zk: node does not exist" {
			err = errors.ErrNotFound
		}
		return nil, err
	} else if bytes.Equal(data, zkDir) {
		return nil, errors.ErrNotFound
	} else {
		return storage.NewFile(file, data), nil
	}
}
