package storage

import (
	"bytes"
	"github.com/ihaiker/aginx/plugins"
	"github.com/ihaiker/aginx/storage/file"
	"github.com/ihaiker/aginx/util"
	"os"
	"path/filepath"
)

type bridge struct {
	plugins.StorageEngine
	LocalStorageEngine plugins.StorageEngine

	watcher   bool
	configDir string

	localWatcher, clusterWatcher <-chan plugins.FileEvent
	closeC                       chan struct{}
}

func NewBridge(cluster string, watcher bool, conf string) *bridge {
	b := &bridge{
		StorageEngine: FindStorage(cluster),
		watcher:       watcher,
		configDir:     filepath.Dir(conf),
		closeC:        make(chan struct{}),
	}
	if b.StorageEngine.IsCluster() {
		b.LocalStorageEngine = file.New(conf)
	}
	b.Initalize()
	return b
}

//更新配置文件，如果是非本地存储调用才有效果
func (sb *bridge) Initalize() {
	if sb.IsCluster() {
		util.PanicIfError(Sync(sb.StorageEngine, sb.LocalStorageEngine))
	}
	sb.clusterWatcher = sb.StorageEngine.StartListener()

	if sb.watcher && sb.LocalStorageEngine != nil {
		sb.localWatcher = sb.LocalStorageEngine.StartListener()
	} else {
		sb.localWatcher = make(chan plugins.FileEvent)
	}
	go sb.StartWatcher()
}

func (sb *bridge) StartWatcher() {
	for {
		select {
		case <-sb.closeC:
			return
		case event, has := <-sb.clusterWatcher:
			if has {
				changed := false
				if event.Type == plugins.FileEventTypeRemove {
					for _, path := range event.Paths {
						absPath := filepath.Join(sb.configDir, path.Name)
						if util.Exists(absPath) {
							changed = true
							err := os.RemoveAll(absPath)
							logger.Info("sync cluster, remove ", path.Name, " ", err)
						}
					}
				} else if event.Type == plugins.FileEventTypeUpdate {
					for _, path := range event.Paths {
						absPath := filepath.Join(sb.configDir, path.Name)
						if write, _ := util.DiffWriteFile(absPath, path.Content); write {
							changed = true
							logger.Info("sync cluster, file ", path.Name)
						}
					}
				}
				if changed {
					util.PublishFileChanged()
				}
			}
		case event, has := <-sb.localWatcher:
			if has {
				changed := false
				if event.Type == plugins.FileEventTypeRemove {
					for _, path := range event.Paths {
						_ = sb.StorageEngine.Remove(path.Name)
					}
				} else if event.Type == plugins.FileEventTypeUpdate {
					for _, path := range event.Paths {
						if file, err := sb.StorageEngine.Get(path.Name); err == os.ErrNotExist {
							_ = sb.StorageEngine.Put(path.Name, path.Content)
							changed = true
						} else if err == nil {
							if !bytes.Equal(file.Content, path.Content) {
								_ = sb.StorageEngine.Put(path.Name, path.Content)
								changed = true
							}
						} else {
							logger.Warn("sync file ", path.Name, " error ", err)
						}
					}
				}
				if changed {
					util.PublishFileChanged()
				}
			}
		}
	}
}

func (sb *bridge) Start() error {
	return util.StartService(sb.StorageEngine)
}

func (sb *bridge) Stop() error {
	close(sb.closeC)
	return util.StopService(sb.StorageEngine)
}

//双向操作,put
func (sb *bridge) Put(file string, content []byte) error {
	if sb.LocalStorageEngine != nil {
		if err := sb.LocalStorageEngine.Put(file, content); err != nil {
			return err
		}
	}
	return sb.StorageEngine.Put(file, content)
}

//双向操作,remove
func (sb *bridge) Remove(file string) error {
	if sb.LocalStorageEngine != nil {
		if err := sb.LocalStorageEngine.Remove(file); err != nil {
			return err
		}
	}
	return sb.StorageEngine.Remove(file)
}
