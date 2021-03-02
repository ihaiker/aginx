package etcd

import (
	"bytes"
	v3 "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/ihaiker/aginx/v2/core/logs"
	"github.com/ihaiker/aginx/v2/core/util"
	"github.com/ihaiker/aginx/v2/core/util/errors"
	"github.com/ihaiker/aginx/v2/plugins/storage"
	"net/url"
	"path/filepath"
)

var logger = logs.New("storage", "engine", "etcd")

type etcdV3Storage struct {
	api    *v3.Client
	folder string
	closeC chan struct{}
	config url.URL
}

func LoadStorage() storage.Plugin {
	return &etcdV3Storage{}
}

func (e *etcdV3Storage) Scheme() string {
	return "etcd"
}

func (e *etcdV3Storage) Name() string {
	return "etcd k/v 存储"
}

func (e *etcdV3Storage) Version() string {
	return "v2.0.0"
}

func (e *etcdV3Storage) GetConfig() url.URL {
	return e.config
}

func (e *etcdV3Storage) Help() string {
	return `etcd存储nginx配置文件。
配置格式：etcd://host:port?param=value
参数说明：
	参数                          值
	username                        
	password                    
	autoSyncInterval            自动同步间隔
	dialTimeout                 拨号超时
	dialKeepAliveTime           拨打“保持活动时间”
	dialKeepAliveTimeout        拨打“保持活动超时”
	tls                         true/false 是否启用https连接consul服务。
	ca                          https ca证书
	cert                        https cert 证书
	key                         https key 证书
`
}

func (e *etcdV3Storage) Initialize(clusterConfig url.URL) (err error) {
	var config *v3.Config
	if config, err = util.Etcd(clusterConfig); err != nil {
		return
	}
	if e.api, err = v3.New(*config); err != nil {
		return
	}
	e.folder = clusterConfig.EscapedPath()
	e.closeC = make(chan struct{})
	return
}

func isDir(content []byte) bool {
	return bytes.HasPrefix(content, []byte("etcdv3_dir_"))
}

func (cs *etcdV3Storage) Put(file string, content []byte) error {
	path := filepath.Join(cs.folder, file)
	_, err := cs.api.Put(cs.api.Ctx(), path, string(content))
	if err == nil {
		logger.Debug("store cluster ", path)
	}
	return err
}

func (cs *etcdV3Storage) Get(file string) (*storage.File, error) {
	path := filepath.Join(cs.folder, file)
	if response, err := cs.api.Get(cs.api.Ctx(), path); err != nil {
		return nil, err
	} else if response.Count == 0 {
		return nil, errors.ErrNotFound
	} else {
		file := storage.NewFile(file, response.Kvs[0].Value)
		return file, nil
	}
}

func (cs *etcdV3Storage) Remove(file string) error {
	key := filepath.Join(cs.folder, file)
	_, err := cs.api.Delete(cs.api.Ctx(), key, v3.WithPrefix())
	if err == nil {
		logger.Debug("delete cluster file ", key)
	}
	return err
}

func (cs *etcdV3Storage) Search(args ...string) ([]*storage.File, error) {
	files := make([]*storage.File, 0)
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
				files = append(files, storage.NewFile(name, kv.Value))
			} else {
				for _, arg := range args {
					if matched, _ := filepath.Match(arg, name); matched {
						reader := storage.NewFile(name, kv.Value)
						files = append(files, reader)
					}
				}
			}
		}
	}
	return files, nil
}

func (cs *etcdV3Storage) Listener() <-chan storage.FileEvent {
	events := make(chan storage.FileEvent)
	go func() {
		defer errors.Catch()
		watch := cs.api.Watch(cs.api.Ctx(), cs.folder, v3.WithPrefix(), v3.WithPrevKV())
		for {
			select {
			case <-cs.closeC:
				return
			case resp := <-watch:
				for _, event := range resp.Events {
					file, _ := filepath.Rel(cs.folder, string(event.Kv.Key))
					if event.Type == mvccpb.DELETE {
						events <- storage.FileEvent{
							Type:  storage.FileEventTypeRemove,
							Paths: []storage.File{{Name: file, Content: event.Kv.Value}},
						}
					} else if event.IsCreate() || event.IsModify() {
						if !isDir(event.Kv.Value) {
							events <- storage.FileEvent{
								Type:  storage.FileEventTypeUpdate,
								Paths: []storage.File{{Name: file, Content: event.Kv.Value}},
							}
						}
					}
				}
			}
		}
	}()
	return events
}

func (es *etcdV3Storage) Stop() error {
	errors.Try(func() {
		close(es.closeC)
	})
	return nil
}

func (es *etcdV3Storage) Start() error {
	return nil
}
