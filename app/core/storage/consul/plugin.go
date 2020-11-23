package consul

import (
	"github.com/hashicorp/consul/api"
	"github.com/ihaiker/aginx/v2/core/logs"
	"github.com/ihaiker/aginx/v2/core/util"
	"github.com/ihaiker/aginx/v2/core/util/errors"
	"github.com/ihaiker/aginx/v2/plugins/storage"
	"net/url"
	"path/filepath"
	"time"
)

var logger = logs.New("storage", "engine", "consul")

type consulStorage struct {
	folder string
	client *api.Client

	index      uint64
	cacheFiles api.KVPairs
	events     chan storage.FileEvent

	closeChan chan struct{}
}

func LoadStorage() storage.Plugin {
	return &consulStorage{}
}

func (c *consulStorage) Scheme() string {
	return "consul"
}

func (c *consulStorage) Name() string {
	return "consul k/v 存储"
}

func (c *consulStorage) Version() string {
	return "v1.0"
}

func (c *consulStorage) Help() string {
	return `使用consul k/v作为配置存储器。
配置格式为：consul://host:port/aginx?param=value . 其中/aginx为k/v存储前缀。
可选参数说明：
	参数              说明
	token          连接consul所需要使用.
	tokenFile      token file 文件
	datacenter     参见 consul datacenter.
	namespace      参见 consul namespace.
	waitTime       连接consul超时时间，默认15秒。
	tls            true/false 是否启用https连接consul服务。
	ca             https ca证书路径
	cert           https cert 证书路径
	key            https key 证书路径
`
}

func (c *consulStorage) Initialize(cfg url.URL) (err error) {
	var config *api.Config
	if config, err = util.Consul(cfg); err != nil {
		return
	}
	if c.client, err = api.NewClient(config); err != nil {
		return
	}
	c.folder = cfg.EscapedPath()[1:]
	c.cacheFiles = api.KVPairs{}
	c.events = make(chan storage.FileEvent)
	return nil
}

func (cs *consulStorage) Listener() <-chan storage.FileEvent {
	return cs.events
}

func (cs *consulStorage) Put(file string, content []byte) error {
	key := filepath.Join(cs.folder, file)
	p := &api.KVPair{Key: key, Value: content}
	_, err := cs.client.KV().Put(p, nil)
	if err == nil {
		logger.Debug("store file ", key)
	}
	return err
}

func (cs *consulStorage) Search(args ...string) ([]*storage.File, error) {
	files := make([]*storage.File, 0)
	if kvPairs, _, err := cs.client.KV().List(cs.folder, nil); err != nil {
		return nil, err
	} else {
		for _, kv := range kvPairs {
			if len(kv.Value) != 0 {
				relPath, _ := filepath.Rel(cs.folder, kv.Key)
				if len(args) == 0 {
					files = append(files, &storage.File{
						Content: kv.Value, Name: relPath,
					})
				} else {
					for _, arg := range args {
						if matched, _ := filepath.Match(arg, relPath); matched {
							files = append(files, &storage.File{
								Content: kv.Value, Name: relPath,
							})
						}
					}
				}
			}
		}
	}
	return files, nil
}

func (cs *consulStorage) Remove(file string) error {
	key := filepath.Join(cs.folder, file)
	_, err := cs.client.KV().DeleteTree(key, nil)
	if err == nil {
		logger.Debug("remove ", key)
	}
	return err
}

func (cs *consulStorage) Get(file string) (*storage.File, error) {
	key := filepath.Join(cs.folder, file)
	if kvPair, _, err := cs.client.KV().Get(key, nil); err != nil {
		return nil, err
	} else if kvPair == nil {
		return nil, errors.ErrNotFound
	} else {
		reader := storage.NewFile(file, kvPair.Value)
		return reader, nil
	}
}

func (c *consulStorage) Start() error {
	go func() {
		defer errors.Catch(func(err error) {
			logger.WithError(err).Warn("watch")
		})
		for {
			select {
			case <-c.closeChan:
				return
			default:
				if err := c.watchChange(); err != nil {
					logger.WithError(err).Warn("watch")
					time.Sleep(time.Second * 3)
				}
			}
		}
	}()
	return nil
}

func (c *consulStorage) Stop() error {
	defer errors.Catch()
	close(c.closeChan)
	return nil
}

func (cs *consulStorage) watchChange() error {
	kvs, query, err := cs.client.KV().List(cs.folder, &api.QueryOptions{
		/*WaitTime: time.Second * 3,*/ WaitIndex: cs.index,
	})
	if err != nil {
		return err
	}

	//index为更新版本号，当更新以后版本会改变
	if cs.index != 0 && cs.index != query.LastIndex {
		//文件修改
		events := storage.FileEvent{
			Type:  storage.FileEventTypeUpdate,
			Paths: []storage.File{},
		}
		for _, kv := range kvs {
			if kv.Value == nil { //isDir
				continue
			}
			if cs.index == 0 || kv.ModifyIndex >= query.LastIndex {
				clusterPath, _ := filepath.Rel(cs.folder, kv.Key)
				events.Paths = append(events.Paths, storage.File{
					Name: clusterPath, Content: kv.Value,
				})
			}
		}
		if len(events.Paths) > 0 {
			cs.events <- events
		}

		events = storage.FileEvent{
			Type:  storage.FileEventTypeRemove,
			Paths: []storage.File{},
		}
		//删除文件
		for _, cacheFile := range cs.cacheFiles {
			has := false
			for _, kv := range kvs {
				if kv.Key == cacheFile.Key {
					has = true
					break
				}
			}
			if !has {
				clusterPath, _ := filepath.Rel(cs.folder, cacheFile.Key)
				events.Paths = append(events.Paths, storage.File{
					Name: clusterPath, Content: cacheFile.Value,
				})
			}
		}
		if len(events.Paths) > 0 {
			cs.events <- events
		}
	}
	cs.cacheFiles = kvs
	cs.index = query.LastIndex
	return nil
}
