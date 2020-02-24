package storage

import (
	"bytes"
	"github.com/ihaiker/aginx/logs"
	"github.com/ihaiker/aginx/plugins"
)

var logger = logs.New("storage", "engine", "bridge")

func contains(args []*plugins.ConfigurationFile, cur *plugins.ConfigurationFile) (*plugins.ConfigurationFile, bool) {
	for _, arg := range args {
		if arg.Name == cur.Name {
			return arg, true
		}
	}
	return nil, false
}

func Sync(from plugins.StorageEngine, to plugins.StorageEngine) error {
	fromFiles, err := from.Search()
	if err != nil {
		return err
	}

	toFiles, err := to.Search()
	if err != nil {
		return err
	}

	for _, formFile := range fromFiles {
		if toFile, has := contains(toFiles, formFile); has {
			if bytes.Equal(toFile.Content, formFile.Content) {
				logger.Debug("ignore update ", formFile.Name)
				continue
			}
		}
		logger.Info("update file ", formFile.Name)
		err := to.Put(formFile.Name, formFile.Content)
		if err != nil {
			logger.Warn("update file ", formFile.Name, " error ", err)
			return err
		}
	}

	for _, toFile := range toFiles {
		if _, has := contains(fromFiles, toFile); !has {
			logger.Info("remove file ", toFile.Name)
			err := to.Remove(toFile.Name)
			if err != nil {
				logger.Warn("remove file ", toFile.Name, " error ", err)
				return err
			}
		}
	}
	return nil
}
