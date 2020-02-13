package ignore

import (
	"github.com/sirupsen/logrus"
	"strings"
)

type clusterIgnore struct {
	files []string
}

func Cluster() *clusterIgnore {
	return &clusterIgnore{files: make([]string, 0)}
}

func (ignore *clusterIgnore) Add(files ...string) {
	ignore.files = append(ignore.files, files...)
	logrus.WithField("module", "ignore").Debug("add ignore files: ", strings.Join(files, ","))
}

func (ignore *clusterIgnore) Is(path string) bool {
	for idx := 0; idx < len(ignore.files); idx++ {
		file := ignore.files[idx]
		if file == path {
			logrus.WithField("module", "ignore").Debug("is ignore file ", file)
			ignore.files = append(ignore.files[0:idx], ignore.files[idx+1:]...)
			return true
		}
	}
	return false
}

func (ignore *clusterIgnore) IfNotIsAdd(path string) bool {
	if ignore.Is(path) {
		return true
	} else {
		ignore.Add(path)
		return false
	}
}
