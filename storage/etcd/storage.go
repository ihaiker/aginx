package etcd

import (
	"bytes"
	v3 "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/ihaiker/aginx/logs"
	"github.com/ihaiker/aginx/plugins"
	"github.com/ihaiker/aginx/util"
	"net/url"
	"path/filepath"
)

var logger = logs.New("storage", "engine", "etcd")

type etcdV3Storage struct {
	api    *v3.Client
	folder string
}

func New(clusterConfig *url.URL) (*etcdV3Storage, error) {
	address := clusterConfig.Host
	folder := clusterConfig.EscapedPath()
	username := clusterConfig.Query().Get("user")
	password := clusterConfig.Query().Get("password")

	cs := &etcdV3Storage{folder: folder}

	config := v3.Config{
		Endpoints: []string{address},
		Username:  username, Password: password,
	}
	if client, err := v3.New(config); err != nil {
		return nil, err
	} else {
		cs.api = client
		return cs, nil
	}
}

func (cs *etcdV3Storage) IsCluster() bool {
	return true
}

func isDir(content []byte) bool {
	return bytes.HasPrefix(content, []byte("etcdv3_dir_"))
}

func (cs *etcdV3Storage) Search(args ...string) ([]*plugins.ConfigurationFile, error) {
	files := make([]*plugins.ConfigurationFile, 0)
	if resp, err := cs.api.Get(cs.api.Ctx(), cs.folder, v3.WithPrefix()); err != nil {
		return nil, err
	} else {
		for _, kv := range resp.Kvs {
			key := string(kv.Key)
			if isDir(kv.Value) {
				continue
			}
			name, _ := filepath.Rel(cs.folder, key)
			if len(args) == 0 {
				files = append(files, plugins.NewFile(name, kv.Value))
			} else {
				for _, arg := range args {
					if matched, _ := filepath.Match(arg, name); matched {
						reader := plugins.NewFile(name, kv.Value)
						files = append(files, reader)
					}
				}
			}
		}
	}
	return files, nil
}

func (cs *etcdV3Storage) Remove(file string) error {
	key := cs.folder + "/" + file
	resp, err := cs.api.Delete(cs.api.Ctx(), key, v3.WithPrefix())
	logger.Debug("delete cluster file ", resp.Deleted)
	return err
}

func (cs *etcdV3Storage) Get(file string) (*plugins.ConfigurationFile, error) {
	path := cs.folder + "/" + file
	if response, err := cs.api.Get(cs.api.Ctx(), path); err != nil {
		return nil, err
	} else if response.Count == 0 {
		return nil, util.ErrNotFound
	} else {
		reader := plugins.NewFile(file, response.Kvs[0].Value)
		return reader, nil
	}
}

func (cs *etcdV3Storage) store(file string, content []byte) error {
	logger.Debug("store cluster ", file)
	_, err := cs.api.Put(cs.api.Ctx(), file, string(content))
	return err
}

func (cs *etcdV3Storage) Put(file string, content []byte) error {
	return cs.store(cs.folder+"/"+file, content)
}

func (cs *etcdV3Storage) StartListener() <-chan plugins.FileEvent {
	events := make(chan plugins.FileEvent)
	go func() {
		defer util.Catch()
		watch := cs.api.Watch(cs.api.Ctx(), cs.folder, v3.WithPrefix(), v3.WithPrevKV())
		for {
			select {
			case resp := <-watch:
				for _, event := range resp.Events {
					file, _ := filepath.Rel(cs.folder, string(event.Kv.Key))
					if event.Type == mvccpb.DELETE {
						events <- plugins.FileEvent{
							Type:  plugins.FileEventTypeRemove,
							Paths: []plugins.ConfigurationFile{{Name: file, Content: event.Kv.Value}},
						}
					} else if event.IsCreate() || event.IsModify() {
						if !isDir(event.Kv.Value) {
							events <- plugins.FileEvent{
								Type:  plugins.FileEventTypeUpdate,
								Paths: []plugins.ConfigurationFile{{Name: file, Content: event.Kv.Value}},
							}
						}
					}
				}
			}
		}
	}()
	return events
}
