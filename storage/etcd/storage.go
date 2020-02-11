package etcd

import (
	"bytes"
	v3 "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/ihaiker/aginx/nginx/configuration"
	"github.com/ihaiker/aginx/storage/file"
	"github.com/ihaiker/aginx/util"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type etcdV3Storage struct {
	closeChan chan struct{}
	wg        *sync.WaitGroup

	client3 *v3.Client
	//client2 *v2.Client

	folder  string
	rootDir string
}

func New(address, folder, username, password string) (*etcdV3Storage, error) {
	cs := &etcdV3Storage{
		closeChan: make(chan struct{}),
		wg:        new(sync.WaitGroup),
		folder:    folder,
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
		cs.client3 = client
		return cs, nil
	}

}

func (cs *etcdV3Storage) watchChanged() {
	cs.wg.Add(1)
	defer cs.wg.Done()

	watch := cs.client3.Watch(cs.client3.Ctx(), cs.folder, v3.WithPrefix())
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
					logrus.WithField("file", string(event.Kv.Key)).WithError(err).Info("remove file")
				} else if event.IsCreate() || event.IsModify() {
					cs.localFile(event.Kv.Key, event.Kv.Value)
				}
			}
			logrus.Info("publish: ", util.NginxReload)
			util.EBus.Publish(util.NginxReload)
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
		return
	}

	err := util.WriterFile(filePath, content)
	logrus.WithField("engine", "consul").WithField("file", string(file)).
		WithError(err).Debug("store the configuration.")
}

func (cs *etcdV3Storage) Start() error {
	if resp, err := cs.client3.Get(cs.client3.Ctx(), cs.folder, v3.WithPrefix()); err != nil {
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
	if resp, err := cs.client3.Get(cs.client3.Ctx(), cs.folder, v3.WithPrefix()); err != nil {
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

func (cs *etcdV3Storage) File(file string) (*util.NameReader, error) {
	path := cs.folder + "/" + file
	if response, err := cs.client3.Get(cs.client3.Ctx(), path); err != nil {
		return nil, err
	} else if response.Count == 0 {
		return nil, os.ErrNotExist
	} else {
		reader := util.NamedReader(bytes.NewBuffer(response.Kvs[0].Value), file)
		return reader, nil
	}
}

func (cs *etcdV3Storage) store(file string, content []byte) error {
	logrus.WithField("engine", "etcd").Debug("store ", file)
	_, err := cs.client3.Put(cs.client3.Ctx(), file, string(content))
	return err
}

func (cs *etcdV3Storage) Store(file string, content []byte) error {
	return cs.store(cs.folder+"/"+file, content)
}

func (cs *etcdV3Storage) StoreConfiguration(cfg *configuration.Configuration) error {
	logrus.WithField("engine", "etcd").Debug("store configuration")
	return configuration.DownWriter(cs.folder, cfg, cs.store)
}
