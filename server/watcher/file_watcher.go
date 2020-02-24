package watcher

import (
	"github.com/ihaiker/aginx/logs"
	ig "github.com/ihaiker/aginx/server/ignore"
	"github.com/ihaiker/aginx/storage"
	"github.com/ihaiker/aginx/storage/file"
	"github.com/ihaiker/aginx/util"
	"github.com/radovskyb/watcher"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var logger = logs.New("watcher")

type FileWatcher struct {
	engine  storage.Engine
	wr      *watcher.Watcher
	ignore  ig.Ignore
	rootDir string
}

func NewFileWatcher(engine storage.Engine, ignore ig.Ignore) *FileWatcher {
	return &FileWatcher{engine: engine, ignore: ignore}
}

//合并rename,move..等操作是文件夹的时候的批量操，查找最小操作执行
func (fw *FileWatcher) mergeEvents(events []watcher.Event) []watcher.Event {
	if len(events) == 1 {
		return events
	}

	switch events[0].Op {
	case watcher.Remove: //删除
		for i := 0; i < len(events); i++ {
			min := i
			for j := i; j < len(events); j++ {
				if len(events[i].Path) > len(events[j].Path) {
					min = j
				}
			}
			if min != i {
				events[min], events[i] = events[i], events[min]
			}
		}
		return events
	case watcher.Rename, watcher.Move, watcher.Create:
		mergeEvents := make([]watcher.Event, 0)

		for i := 0; i < len(events); i++ {

			//搜索最小化操作
			minEvent := &events[i]
			for j := 0; j < len(events); j++ {
				if i == j {
					continue
				}
				if strings.HasPrefix(minEvent.OldPath, events[j].OldPath) && strings.HasPrefix(minEvent.Path, events[j].Path) {
					minEvent = &events[j]
				}
			}

			//是否包含更小的操作
			has := false
			for _, event := range mergeEvents {
				if strings.HasPrefix(minEvent.OldPath, event.OldPath) && strings.HasPrefix(minEvent.Path, event.Path) {
					has = true
					break
				}
			}
			if !has {
				mergeEvents = append(mergeEvents, *minEvent)
			}
		}
		return mergeEvents
	default:
		return events
	}
}

func (fw *FileWatcher) handlerEvent(events []watcher.Event) {
	if len(events) == 0 {
		return
	}

	events = fw.mergeEvents(events) //合并至最小化操作
	for _, event := range events {
		clusterPath := strings.Replace(event.Path, fw.rootDir+"/", "", 1)

		switch event.Op {
		case watcher.Create, watcher.Write:
			if event.IsDir() {
				continue
			}
			if fw.ignore.IfNotIsAdd(clusterPath) {
				continue
			}

			logger.Debug("write ", event.Path)
			bs, _ := ioutil.ReadFile(event.Path)
			if err := fw.engine.Put(clusterPath, bs); err != nil {
				logger.Warn("store file ", clusterPath, ", error:", err)
			}

		case watcher.Remove:

			if fw.ignore.IfNotIsAdd(clusterPath) {
				continue
			}

			logger.Debug("remove ", event.Path)
			if err := fw.engine.Remove(clusterPath); err != nil {
				logger.Warn("store remove file error", err)
			}

		case watcher.Rename, watcher.Move:
			logger.Debug("not support move ", event.OldPath, " to ", event.Path, " move back")
			_ = fw.wr.RemoveRecursive(fw.rootDir)
			_ = os.Rename(event.Path, event.OldPath)
			_ = fw.wr.AddRecursive(fw.rootDir)
		}
	}

	logger.Info("publish: ", util.StorageFileChanged)
	util.EBus.Publish(util.StorageFileChanged)
}

func (fw *FileWatcher) Start() error {
	fw.wr = watcher.New()
	fw.wr.IgnoreHiddenFiles(true)
	/*fw.wr.AddFilterHook(func(info os.FileInfo, fullPath string) error {
		if info.Mode()&os.ModeSymlink == os.ModeSymlink {
			return http.ErrNotSupported
		}
		return nil
	})*/
	_, conf, _ := file.GetInfo()
	fw.rootDir = filepath.Dir(conf)
	if err := fw.wr.AddRecursive(fw.rootDir); err != nil {
		return err
	}

	fw.wr.FilterOps(watcher.Create, watcher.Write, watcher.Rename, watcher.Remove, watcher.Move)
	go func() {
		events := make([]watcher.Event, 0) //移动，删除，重命名等操作会产生批量时间，把这些操作转成一次执行
		timer := time.NewTimer(time.Millisecond * 200)
		for {
			timer.Reset(time.Millisecond * 200)

			select {
			case event := <-fw.wr.Event:
				if event.IsDir() && event.Op == watcher.Write {
					continue
				}
				events = append(events, event)
			case <-timer.C:
				fw.handlerEvent(events)
				events = events[0:0]
			case <-fw.wr.Closed:
				timer.Stop()
				fw.handlerEvent(events)
				events = events[0:0]
				return
			}
		}
	}()
	go func() {
		err := fw.wr.Start(time.Millisecond * 300)
		logrus.WithError(err).Warn("start file watcher")
	}()
	return nil
}

func (fw *FileWatcher) Stop() error {
	fw.wr.Close()
	return nil
}
