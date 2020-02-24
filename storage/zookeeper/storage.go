package zookeeper

import (
	"bytes"
	"github.com/ihaiker/aginx/logs"
	ig "github.com/ihaiker/aginx/server/ignore"
	fileStorage "github.com/ihaiker/aginx/storage/file"
	"github.com/ihaiker/aginx/util"
	"github.com/samuel/go-zookeeper/zk"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var logger = logs.New("storage", "engine", "zk")
var zkDirData = []byte("zkdir")

type zkStorage struct {
	address string
	folder  string
	keeper  *zk.Conn
	watcher *Watcher
	ignore  ig.Ignore
}

func New(clusterConfig *url.URL, ignore ig.Ignore) (zks *zkStorage, err error) {
	address := clusterConfig.Host
	folder := clusterConfig.EscapedPath()[1:]
	scheme := clusterConfig.Query().Get("scheme")
	auth := clusterConfig.Query().Get("auth")

	zks = &zkStorage{address: address, folder: folder, ignore: ignore}
	if !strings.HasPrefix(zks.folder, "/") {
		zks.folder = "/" + zks.folder
	}
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

func (zks *zkStorage) zkList(path string, onlyFile bool) ([]*util.NameReader, error) {
	readers := make([]*util.NameReader, 0)
	keys, _, err := zks.keeper.Children(path)
	if err != nil {
		return nil, err
	}
	for _, key := range keys {
		file := path + "/" + key
		if data, st, err := zks.keeper.Get(file); err != nil {
			return nil, err
		} else {
			if bytes.Equal(data, zkDirData) || st.NumChildren > 0 || len(data) == 0 {
				if !onlyFile {
					readers = append(readers, util.NamedReader(nil, file))
				}
				if files, err := zks.zkList(file, onlyFile); err != nil {
					return nil, err
				} else {
					readers = append(readers, files...)
				}
			} else {
				readers = append(readers, util.NamedReader(bytes.NewBuffer(data), file))
			}
		}
	}
	return readers, err
}

func (zks *zkStorage) Search(args ...string) ([]*util.NameReader, error) {
	matched := make([]*util.NameReader, 0)

	readers, err := zks.zkList(zks.folder, true)
	if err == nil {
	READERS:
		for _, reader := range readers {
			for _, arg := range args {
				pattern := zks.folder + "/" + arg
				if match, _ := filepath.Match(pattern, reader.Name); match {
					reader.Name = strings.ReplaceAll(reader.Name, zks.folder+"/", "")
					matched = append(matched, reader)
					continue READERS
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

func (zks *zkStorage) Get(file string) (*util.NameReader, error) {
	path := zks.folder + "/" + file
	if data, _, err := zks.keeper.Get(path); err != nil {
		return nil, err
	} else if bytes.Equal(data, zkDirData) {
		return nil, os.ErrNotExist
	} else {
		return util.NamedReader(bytes.NewBuffer(data), file), nil
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

func (zks *zkStorage) publishFileChangedEvent() {
	logger.Info("publish: ", util.StorageFileChanged)
	util.EBus.Publish(util.StorageFileChanged)
}

func (zks *zkStorage) watchEvent(rootDir string) {
	for event := range zks.watcher.C {
		logger.Debug("event ", event.Type.String(), " ", event.Path)
		localFile := rootDir + "/" + strings.Replace(event.Path, zks.folder, "", 1)
		switch event.Type {
		case zk.EventNodeCreated:
			data, _, _ := zks.keeper.Get(event.Path)
			isDir := bytes.Equal(data, zkDirData) || len(data) == 0
			if isDir {
				if err := os.MkdirAll(localFile, os.ModePerm); err != nil {
					logger.Warn("mkdir ", localFile, " error ", err)
				}
			} else {
				if err := ioutil.WriteFile(localFile, data, 0666); err != nil {
					logger.Warn("open file ", localFile, " error ", err)
				}
				zks.publishFileChangedEvent()
			}

		case zk.EventNodeDeleted:
			if fileInfo, err := os.Stat(localFile); err != nil {
				if !os.IsNotExist(err) {
					logger.Warn("open file ", localFile, " error ", err)
				}
			} else if fileInfo.IsDir() {
				logger.Warn("delete folder ", localFile, " error ", os.RemoveAll(localFile))
			} else {
				logger.Warn("delete file ", localFile, " error ", os.Remove(localFile))
			}
			zks.publishFileChangedEvent()

		case zk.EventNodeDataChanged:
			data, _, _ := zks.keeper.Get(event.Path)
			if !(bytes.Equal(data, zkDirData) || len(data) == 0) {
				err := ioutil.WriteFile(localFile, data, 0666)
				logger.Warn("write file changed ", localFile, " error ", err)
			}
			zks.publishFileChangedEvent()
		}
	}
}

func (zks *zkStorage) Start() error {
	_, conf, err := fileStorage.GetInfo()
	if err != nil {
		return err
	}

	//clear file
	rootDir := filepath.Dir(conf)
	if err := os.RemoveAll(rootDir); err != nil {
		return err
	}

	zkFiles, err := zks.zkList(zks.folder, false)
	if err != nil {
		return err
	}

	zks.watcher.Folder(zks.folder)

	for _, zkFile := range zkFiles {
		if zkFile.Reader != nil { //file
			filePath := rootDir + strings.Replace(zkFile.Name, zks.folder, "", 1)
			logger.Debug("sync file ", zkFile.Name)
			if err := util.WriteReader(filePath, zkFile); err != nil {
				return err
			}
			zks.watcher.File(zkFile.Name)
		} else {
			zks.watcher.Folder(zkFile.Name)
		}
	}
	go zks.watchEvent(rootDir)
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
