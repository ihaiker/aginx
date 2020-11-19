package storage

import (
	"github.com/ihaiker/aginx/v2/core/config"
	"github.com/ihaiker/aginx/v2/core/nginx"
	"github.com/ihaiker/aginx/v2/core/storage/file"
	"github.com/ihaiker/aginx/v2/core/util/errors"
	"github.com/ihaiker/aginx/v2/plugins/storage"
	"net/url"
)

type fileChangeHandler struct {
	engine storage.Plugin
	hooks  []func(event storage.FileEvent)
}

func Changed(engine storage.Plugin, fn ...func(storage.FileEvent)) *fileChangeHandler {
	return &fileChangeHandler{
		engine: engine,
		hooks:  fn,
	}
}

func (f *fileChangeHandler) watch(local storage.Plugin) {
	eventC := f.engine.Listener()
	for {
		select {
		case event, ok := <-eventC:
			if ok {
				for _, path := range event.Paths {
					logger.Debugf("文件 %s %s", path.Name, event.Type)
					if event.Type == storage.FileEventTypeRemove {
						if err := local.Remove(path.Name); err != nil {
							logger.WithError(err).Warnf("删除 %s", path.Name)
						}
					} else {
						if err := local.Put(path.Name, path.Content); err != nil {
							logger.WithError(err).Warnf("更新 %s", path.Name)
						}
					}
				}
				for _, hook := range f.hooks {
					errors.Try(func() {
						hook(event)
					})
				}
			}
		}
	}
}
func (f *fileChangeHandler) Start() error {
	if f.engine.Scheme() == "file" {
		return nil
	}

	local := file.LoadStorage()
	if _, conf, err := nginx.HelpInfo(config.Config.Nginx); err != nil {
		return err
	} else {
		cfg, _ := url.Parse("file:/" + conf)
		if err = local.Initialize(*cfg); err != nil {
			return err
		}
	}

	//下载配置到本地
	if err := Sync(f.engine, local); err != nil {
		return errors.Wrap(err, "下载同步配置")
	}

	go f.watch(local)
	return nil
}

func (f *fileChangeHandler) Stop() error {
	if f.engine.Scheme() == "file" {
		return nil
	}
	return nil
}
