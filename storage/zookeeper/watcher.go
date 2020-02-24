package zookeeper

import (
	"github.com/samuel/go-zookeeper/zk"
	"time"
)

type Watcher struct {
	C       chan zk.Event
	done    chan struct{}
	watched map[string]string
	keeper  *zk.Conn
}

func NewWatcher(keeper *zk.Conn) *Watcher {
	watcher := new(Watcher)
	watcher.C = make(chan zk.Event)
	watcher.done = make(chan struct{})
	watcher.keeper = keeper
	watcher.watched = make(map[string]string)
	return watcher
}

func (br *Watcher) Folder(name string) {
	br.watch(name, func(name string) (events <-chan zk.Event, err error) {
		_, _, events, err = br.keeper.ChildrenW(name)
		return
	})
}

func (br *Watcher) File(name string) {
	br.watch(name, func(name string) (events <-chan zk.Event, err error) {
		_, _, events, err = br.keeper.GetW(name)
		return
	})
}

func (br *Watcher) watch(name string, fn func(string) (<-chan zk.Event, error)) {
	br.watched[name] = name
	go func() {
		defer delete(br.watched, name)
		for {
			ec, err := fn(name)
			if err == zk.ErrNoNode {
				return
			} else if err != nil {
				logger.WithField("path", name).Debug("watch error ", err)
				time.Sleep(time.Second)
				continue
			}
			select {
			case <-br.done:
				return
			case event := <-ec:
				br.filter(event)
			}
		}
	}()
}

func (br *Watcher) filter(event zk.Event) {
	switch event.Type {
	case zk.EventNodeChildrenChanged:
		childless, _, _ := br.keeper.Children(event.Path)
		for _, children := range childless {
			path := event.Path + "/" + children
			if _, has := br.watched[path]; !has {
				data, _, _ := br.keeper.Get(path)
				if isDir(data) {
					br.Folder(path)
				} else {
					br.File(path)
					br.C <- zk.Event{Type: zk.EventNodeCreated, Path: path}
				}
			}
		}
	case zk.EventNodeDeleted:
		delete(br.watched, event.Path)
		br.C <- event
	case zk.EventNodeDataChanged:
		br.C <- event
	}
}

func (br *Watcher) Close() {
	close(br.done)
}
