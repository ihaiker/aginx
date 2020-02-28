package plugins

import (
	"bytes"
	"net/url"
)

type LoadStorage func(config *url.URL) (StorageEngine, error)

const (
	PLUGIN_STORAGE = "LoadStorage"
)

type FileEventType string

const (
	FileEventTypeUpdate FileEventType = "update"
	FileEventTypeRemove FileEventType = "remove"
)

type ConfigurationFile struct {
	Name    string
	Content []byte
}

func (cfg *ConfigurationFile) String() string {
	return string(cfg.Content)
}

func NewFile(name string, content []byte) *ConfigurationFile {
	return &ConfigurationFile{Name: name, Content: content}
}

type FileEvent struct {
	Type  FileEventType
	Paths []ConfigurationFile
}

func (fe *FileEvent) String() string {
	out := bytes.NewBufferString(string(fe.Type))
	out.WriteString("(")
	for i, path := range fe.Paths {
		if i != 0 {
			out.WriteString(",")
		}
		out.WriteString(path.Name)
	}
	out.WriteString(")")
	return out.String()
}

type StorageEngine interface {
	IsCluster() bool

	StartListener() <-chan FileEvent

	//存储文件内容
	Put(file string, content []byte) error

	Remove(file string) error

	//搜索文件,如果args长度为空，则显示全部文件
	Search(pattern ...string) ([]*ConfigurationFile, error)

	//获取文件
	Get(file string) (*ConfigurationFile, error)
}
