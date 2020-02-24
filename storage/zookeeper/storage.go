package zookeeper

import (
	"bytes"
	"github.com/ihaiker/aginx/logs"
	"github.com/ihaiker/aginx/plugins"
	"github.com/ihaiker/aginx/util"
	"github.com/samuel/go-zookeeper/zk"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

var logger = logs.New("storage", "engine", "zk")
var zkDirData = []byte("zkdir")

type zkStorage struct {
	folder  string
	keeper  *zk.Conn
	watcher *Watcher
}

func New(clusterConfig *url.URL) (zks *zkStorage, err error) {
	address := clusterConfig.Host
	folder := clusterConfig.EscapedPath()
	scheme := clusterConfig.Query().Get("scheme")
	auth := clusterConfig.Query().Get("auth")

	zks = &zkStorage{folder: folder}
	zks.keeper, _, err = zk.Connect([]string{address}, time.Second*3)
	zks.keeper.SetLogger(logger)
	if scheme != "" {
		if err = zks.keeper.AddAuth(scheme, []byte(auth)); err != nil {
			return nil, err
		}
	}
	zks.watcher = NewWatcher(zks.keeper)
	err = zks.zkMkdir(zks.folder)
	return
}

func (zks *zkStorage) IsCluster() bool {
	return true
}

func (zks *zkStorage) zkList(path string, onlyFile bool) ([]*plugins.ConfigurationFile, error) {
	readers := make([]*plugins.ConfigurationFile, 0)
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
					readers = append(readers, plugins.NewFile(file, nil))
				}
				if files, err := zks.zkList(file, onlyFile); err != nil {
					return nil, err
				} else {
					readers = append(readers, files...)
				}
			} else {
				relPath, _ := filepath.Rel(zks.folder, file)
				readers = append(readers, plugins.NewFile(relPath, data))
			}
		}
	}
	return readers, err
}

func isDir(data []byte) bool {
	return bytes.Equal(data, zkDirData) || len(data) == 0
}

func (zks *zkStorage) Search(args ...string) ([]*plugins.ConfigurationFile, error) {
	matched := make([]*plugins.ConfigurationFile, 0)

	zkFiles, err := zks.zkList(zks.folder, true)
	if err == nil {
	READERS:
		for _, zkFile := range zkFiles {
			if len(args) == 0 {
				matched = append(matched, &plugins.ConfigurationFile{
					Name:    zkFile.Name,
					Content: zkFile.Content,
				})
			} else {
				for _, arg := range args {
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
	logger.WithError(err).Debug("remove cluster ", file)
	return err
}

func (zks *zkStorage) Get(file string) (*plugins.ConfigurationFile, error) {
	path := zks.folder + "/" + file
	if data, _, err := zks.keeper.Get(path); err != nil {
		if err.Error() == "zk: node does not exist" {
			err = os.ErrNotExist
		}
		return nil, err
	} else if bytes.Equal(data, zkDirData) {
		return nil, os.ErrNotExist
	} else {
		return plugins.NewFile(file, data), nil
	}
}

func (zks *zkStorage) zkMkdir(file string) error {
	dir := filepath.Dir(file)
	exists, _, err := zks.keeper.Exists(dir)

	if err == zk.ErrNoNode || !exists {
		if err = zks.zkMkdir(dir); err != nil {
			return err
		}
		logger.Info("create dir ", dir)
		if _, err = zks.keeper.Create(dir, zkDirData, 0, zk.WorldACL(zk.PermAll)); err != nil {
			return err
		}
	}
	return err
}

func (zks *zkStorage) zkStore(file string, content []byte) error {
	if err := zks.zkMkdir(file); err != nil {
		return err
	}
	if exists, stat, err := zks.keeper.Exists(file); err != nil {
		return err
	} else if exists {
		_, err := zks.keeper.Set(file, content, stat.Version)
		logger.WithError(err).Debug("store cluster file ", file)
		return err
	} else {
		_, err := zks.keeper.Create(file, content, 0, zk.WorldACL(zk.PermAll))
		logger.WithError(err).Debug("store cluster file ", file)
		return err
	}
}

func (zks *zkStorage) Put(file string, content []byte) error {
	path := zks.folder + "/" + file
	return zks.zkStore(path, content)
}

func (zks *zkStorage) StartListener() <-chan plugins.FileEvent {
	events := make(chan plugins.FileEvent)
	if zkFiles, err := zks.zkList(zks.folder, false); err == nil {
		zks.watcher.Folder(zks.folder)
		for _, zkFile := range zkFiles {
			if zkFile.Content != nil { //file
				zks.watcher.File(zkFile.Name)
			} else {
				zks.watcher.Folder(zkFile.Name)
			}
		}
		go func() {
			defer util.Catch()

			for event := range zks.watcher.C {
				relPath, _ := filepath.Rel(zks.folder, event.Path)
				switch event.Type {
				case zk.EventNodeCreated, zk.EventNodeDataChanged:
					data, _, _ := zks.keeper.Get(event.Path)
					if !isDir(data) {
						events <- plugins.FileEvent{
							Type: plugins.FileEventTypeUpdate,
							Paths: []plugins.ConfigurationFile{{
								Name: relPath, Content: data,
							}},
						}
					}
				case zk.EventNodeDeleted:
					events <- plugins.FileEvent{
						Type: plugins.FileEventTypeRemove,
						Paths: []plugins.ConfigurationFile{{
							Name: relPath, Content: []byte{},
						}},
					}
				}
			}
		}()
	}
	return events
}

func (zks *zkStorage) Start() error {
	return nil
}

func (zks *zkStorage) Stop() error {
	if zks.watcher != nil {
		zks.watcher.Close()
	}
	if zks.keeper != nil {
		zks.keeper.Close()
	}
	return nil
}
