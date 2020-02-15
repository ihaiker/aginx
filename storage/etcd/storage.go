package etcd

import (
	"bytes"
	v3 "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/ihaiker/aginx/logs"
	"github.com/ihaiker/aginx/nginx/configuration"
	ig "github.com/ihaiker/aginx/server/ignore"
	"github.com/ihaiker/aginx/storage/file"
	"github.com/ihaiker/aginx/util"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var logger = logs.New("storage", "engine", "etcd")

type etcdV3Storage struct {
	closeChan chan struct{}
	wg        *sync.WaitGroup

	etcdApi *v3.Client
	//client2 *v2.Client

	folder  string
	rootDir string

	ignore ig.Ignore
}

func New(clusterConfig *url.URL, ignore ig.Ignore) (*etcdV3Storage, error) {
	address := clusterConfig.Host
	folder := clusterConfig.EscapedPath()[1:]
	username := clusterConfig.Query().Get("user")
	password := clusterConfig.Query().Get("password")

	cs := &etcdV3Storage{
		closeChan: make(chan struct{}),
		wg:        new(sync.WaitGroup),
		folder:    folder, ignore: ignore,
	}
	if !strings.HasPrefix(cs.folder, "/") {
		cs.folder = "/" + cs.folder
	}
	if _, conf, err := file.GetInfo(); err != nil {
		return nil, err
	} else {
		cs.rootDir = filepath.Dir(conf)
	}

	config := v3.Config{
		Endpoints: []string{address},
		Username:  username, Password: password,
	}
	if client, err := v3.New(config); err != nil {
		return nil, err
	} else {
		cs.etcdApi = client
		return cs, nil
	}

}

func (cs *etcdV3Storage) IsCluster() bool {
	return true
}

func (cs *etcdV3Storage) watchChanged() {
	cs.wg.Add(1)
	defer cs.wg.Done()

	watch := cs.etcdApi.Watch(cs.etcdApi.Ctx(), cs.folder, v3.WithPrefix(), v3.WithPrevKV())
	for {
		select {
		case <-cs.closeChan:
			return
		case resp := <-watch:
			for _, event := range resp.Events {
				if event.Type == mvccpb.DELETE {
					filePath := cs.rootDir + strings.Replace(string(event.Kv.Key), cs.folder, "", 1)
					var err error
					if isDir(event.Kv.Value) {
						err = os.RemoveAll(filePath)
					} else {
						err = os.Remove(filePath)
					}
					logger.WithError(err).Info("remove local file : ", string(event.Kv.Key))
				} else if event.IsCreate() || event.IsModify() {
					cs.localFile(event.Kv.Key, event.Kv.Value)
				}
			}
			logger.Info("publish: ", util.StorageFileChanged)
			util.EBus.Publish(util.StorageFileChanged)
		}
	}
}

func isDir(content []byte) bool {
	return bytes.HasPrefix(content, []byte("etcdv3_dir_"))
}

func (cs *etcdV3Storage) localFile(file, content []byte) {
	filePath := cs.rootDir + strings.Replace(string(file), cs.folder, "", 1)

	//is folder
	if isDir(content) {
		_ = os.MkdirAll(filePath, os.ModePerm)
		logger.Debug("mkdir local ", filePath)
		return
	}

	err := util.WriterFile(filePath, content)
	logger.WithError(err).Debug("down file ", string(file))
}

func (cs *etcdV3Storage) Start() error {
	if resp, err := cs.etcdApi.Get(cs.etcdApi.Ctx(), cs.folder, v3.WithPrefix()); err != nil {
		return err
	} else {
		for _, kv := range resp.Kvs {
			cs.localFile(kv.Key, kv.Value)
		}
	}
	go cs.watchChanged()
	return nil
}

func (cs *etcdV3Storage) Stop() error {
	if cs.closeChan != nil {
		close(cs.closeChan)
	}
	cs.wg.Wait()
	return nil
}

func (cs *etcdV3Storage) Search(args ...string) ([]*util.NameReader, error) {
	readers := make([]*util.NameReader, 0)
	if resp, err := cs.etcdApi.Get(cs.etcdApi.Ctx(), cs.folder, v3.WithPrefix()); err != nil {
		return nil, err
	} else {
		for _, kv := range resp.Kvs {
			key := string(kv.Key)
			if isDir(kv.Value) {
				continue
			}
			name := strings.ReplaceAll(key, cs.folder+"/", "")
			for _, arg := range args {
				if matched, _ := filepath.Match(arg, name); matched {
					reader := util.NamedReader(bytes.NewBuffer(kv.Value), name)
					readers = append(readers, reader)
				}
			}
		}
	}
	return readers, nil
}

func (cs *etcdV3Storage) Remove(file string) error {
	key := cs.folder + "/" + file
	resp, err := cs.etcdApi.Delete(cs.etcdApi.Ctx(), key, v3.WithPrefix())
	logger.Debug("delete cluster file ", resp.Deleted)
	return err
}

func (cs *etcdV3Storage) File(file string) (*util.NameReader, error) {
	path := cs.folder + "/" + file
	if response, err := cs.etcdApi.Get(cs.etcdApi.Ctx(), path); err != nil {
		return nil, err
	} else if response.Count == 0 {
		return nil, os.ErrNotExist
	} else {
		reader := util.NamedReader(bytes.NewBuffer(response.Kvs[0].Value), file)
		return reader, nil
	}
}

func (cs *etcdV3Storage) store(file string, content []byte) error {
	logger.Debug("store cluster ", file)
	_, err := cs.etcdApi.Put(cs.etcdApi.Ctx(), file, string(content))
	return err
}

func (cs *etcdV3Storage) Store(file string, content []byte) error {
	return cs.store(cs.folder+"/"+file, content)
}

func (cs *etcdV3Storage) StoreConfiguration(cfg *configuration.Configuration) error {
	logger.Debug("store configuration")
	return configuration.DownWriter(cs.folder, cfg, cs.store)
}
