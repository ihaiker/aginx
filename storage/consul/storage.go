package consul

import (
	consulApi "github.com/hashicorp/consul/api"
	"github.com/ihaiker/aginx/logs"
	"github.com/ihaiker/aginx/plugins"
	"github.com/ihaiker/aginx/util"
	"net/url"
	"path/filepath"
)

var logger = logs.New("storage", "engine", "consul")

type consulStorage struct {
	address string
	folder  string
	client  *consulApi.Client
}

func New(clusterConfig *url.URL) (cs *consulStorage, err error) {
	address := clusterConfig.Host
	folder := clusterConfig.EscapedPath()[1:]
	token := clusterConfig.Query().Get("token")

	cs = &consulStorage{
		address: address, folder: folder,
	}
	config := consulApi.DefaultConfig()
	config.Address = cs.address
	config.Token = token
	if cs.client, err = consulApi.NewClient(config); err != nil {
		return
	}
	return
}

func (cs *consulStorage) IsCluster() bool {
	return true
}

func (cs *consulStorage) StartListener() <-chan plugins.FileEvent {
	return NewWatcher(cs.folder, cs.client)
}

func (cs *consulStorage) Search(args ...string) ([]*plugins.ConfigurationFile, error) {
	files := make([]*plugins.ConfigurationFile, 0)
	if kvPairs, _, err := cs.client.KV().List(cs.folder, nil); err != nil {
		return nil, err
	} else {
		for _, kv := range kvPairs {
			if len(kv.Value) != 0 {
				relPath, _ := filepath.Rel(cs.folder, kv.Key)
				if len(args) == 0 {
					files = append(files, &plugins.ConfigurationFile{
						Content: kv.Value, Name: relPath,
					})
				} else {
					for _, arg := range args {
						if matched, _ := filepath.Match(arg, relPath); matched {
							files = append(files, &plugins.ConfigurationFile{
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
	logger.Debug("remove ", key)

	if kvs, _, err := cs.client.KV().List(key, nil); err != nil {
		return err
	} else {
		for _, kv := range kvs {
			_, _ = cs.client.KV().Delete(kv.Key, nil)
		}
	}
	_, err := cs.client.KV().Delete(key, nil)
	return err
}

func (cs *consulStorage) Get(file string) (*plugins.ConfigurationFile, error) {
	key := filepath.Join(cs.folder, file)

	if kvPair, _, err := cs.client.KV().Get(key, nil); err != nil {
		return nil, err
	} else if kvPair == nil {
		return nil, util.ErrNotFound
	} else {
		reader := plugins.NewFile(file, kvPair.Value)
		return reader, nil
	}
}

func (cs *consulStorage) store(file string, content []byte) error {
	logger.Debug("store file ", file)
	p := &consulApi.KVPair{Key: file, Value: content}
	if _, err := cs.client.KV().Put(p, nil); err != nil {
		logger.WithError(err).Debug("store file: ", file)
		return err
	}
	return nil
}

func (cs *consulStorage) Put(file string, content []byte) error {
	return cs.store(cs.folder+"/"+file, content)
}
