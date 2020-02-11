package consul

import (
	"bytes"
	consulApi "github.com/hashicorp/consul/api"
	"github.com/ihaiker/aginx/nginx/configuration"
	"github.com/ihaiker/aginx/storage/file"
	"github.com/ihaiker/aginx/util"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type consulStorage struct {
	closeChan chan struct{}
	wg        *sync.WaitGroup
	address   string
	folder    string
	client    *consulApi.Client
	index     uint64
	rootDir   string
}

func New(address, folder, token string) (cs *consulStorage, err error) {
	cs = new(consulStorage)
	cs.address = address
	cs.folder = folder
	cs.closeChan = make(chan struct{})
	cs.wg = new(sync.WaitGroup)
	cs.index = 0

	if _, conf, err := file.GetInfo(); err != nil {
		return nil, err
	} else {
		cs.rootDir = filepath.Dir(conf)
	}

	config := consulApi.DefaultConfig()
	config.Address = cs.address
	config.Token = token
	if cs.client, err = consulApi.NewClient(config); err != nil {
		return
	}
	return
}

func (cs *consulStorage) downloadFile() (changed bool) {
	kvs, query, err := cs.client.KV().List(cs.folder, &consulApi.QueryOptions{
		WaitTime: time.Second * 3, WaitIndex: cs.index,
	})
	if err != nil {
		return
	}

	if cs.index != query.LastIndex {

		_ = filepath.Walk(cs.rootDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return err
			}
			return os.Remove(path)
		})

		for _, kv := range kvs {
			filePath := cs.rootDir + strings.Replace(kv.Key, cs.folder, "", 1)
			err := util.WriterFile(filePath, kv.Value)
			if cs.index == 0 || kv.ModifyIndex >= query.LastIndex {
				changed = true
				logrus.WithField("engine", "consul").WithField("file", kv.Key).
					WithError(err).Debug("the configuration has changed.")
			}
		}
	}

	cs.index = query.LastIndex
	return
}

func (cs *consulStorage) watchChanged() {
	cs.wg.Add(1)
	defer cs.wg.Done()
	for {
		select {
		case <-cs.closeChan:
			return
		default:
			if cs.downloadFile() {
				logrus.Info("publish: ", util.NginxReload)
				util.EBus.Publish(util.NginxReload)
			}
		}
	}
}

func (cs *consulStorage) Start() error {
	cs.downloadFile()
	go cs.watchChanged()
	return nil
}

func (cs *consulStorage) Stop() error {
	if cs.closeChan != nil {
		close(cs.closeChan)
	}
	cs.wg.Wait()
	return nil
}

func (cs *consulStorage) Search(args ...string) ([]*util.NameReader, error) {
	readers := make([]*util.NameReader, 0)
	if keys, _, err := cs.client.KV().Keys(cs.folder, "", nil); err != nil {
		return nil, err
	} else {
		for _, key := range keys {
			if strings.HasSuffix(key, "/") {
				continue
			}
			name := strings.ReplaceAll(key, cs.folder+"/", "")
			for _, arg := range args {
				if matched, _ := filepath.Match(arg, name); matched {
					reader, _ := cs.File(name)
					readers = append(readers, reader)
				}
			}
		}
	}
	return readers, nil
}

func (cs *consulStorage) File(file string) (*util.NameReader, error) {
	key := cs.folder + "/" + file
	if value, _, err := cs.client.KV().Get(key, nil); err != nil {
		return nil, err
	} else if value == nil {
		return nil, os.ErrNotExist
	} else {
		reader := util.NamedReader(bytes.NewBuffer(value.Value), file)
		return reader, nil
	}
}

func (cs *consulStorage) store(file string, content []byte) error {
	p := &consulApi.KVPair{Key: file, Value: content}
	if _, err := cs.client.KV().Put(p, nil); err != nil {
		return err
	}
	return nil
}

func (cs *consulStorage) Store(file string, content []byte) error {
	return cs.store(cs.folder+"/"+file, content)
}

func (cs *consulStorage) StoreConfiguration(cfg *configuration.Configuration) error {
	return configuration.DownWriter(cs.folder, cfg, cs.store)
}
