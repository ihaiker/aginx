package consul

import (
	consulApi "github.com/hashicorp/consul/api"
	"github.com/ihaiker/aginx/plugins"
	"github.com/ihaiker/aginx/util"
	"path/filepath"
	"time"
)

type watcher struct {
	folder string
	client *consulApi.Client

	index      uint64
	cacheFiles consulApi.KVPairs

	Listener chan plugins.FileEvent
}

func NewWatcher(folder string, client *consulApi.Client) chan plugins.FileEvent {
	w := watcher{
		folder:     folder,
		client:     client,
		index:      0,
		cacheFiles: consulApi.KVPairs{},
		Listener:   make(chan plugins.FileEvent),
	}
	go func() {
		defer util.Catch(func(err error) {
			logger.Info("water error: ", err)
		})
		for {
			w.watchChange()
		}
	}()
	return w.Listener
}

func (cs *watcher) watchChange() {
	kvs, query, err := cs.client.KV().List(cs.folder, &consulApi.QueryOptions{
		WaitTime: time.Second * 3, WaitIndex: cs.index,
	})
	if err != nil {
		return
	}

	if cs.index != 0 && cs.index != query.LastIndex {
		//文件修改
		events := plugins.FileEvent{
			Type:  plugins.FileEventTypeUpdate,
			Paths: []plugins.ConfigurationFile{},
		}
		for _, kv := range kvs {
			if cs.index == 0 || kv.ModifyIndex >= query.LastIndex {
				clusterPath, _ := filepath.Rel(cs.folder, kv.Key)
				events.Paths = append(events.Paths, plugins.ConfigurationFile{
					Name: clusterPath, Content: kv.Value,
				})
			}
		}
		if len(events.Paths) > 0 {
			cs.Listener <- events
		}

		events = plugins.FileEvent{
			Type:  plugins.FileEventTypeRemove,
			Paths: []plugins.ConfigurationFile{},
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
				events.Paths = append(events.Paths, plugins.ConfigurationFile{
					Name: clusterPath, Content: cacheFile.Value,
				})
			}
		}
		if len(events.Paths) > 0 {
			cs.Listener <- events
		}
	}
	cs.cacheFiles = kvs
	cs.index = query.LastIndex
}
