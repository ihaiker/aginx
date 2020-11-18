package storage

import (
	"bytes"
)

type FileEventType string

const (
	//文件更新（创建、更新）
	FileEventTypeUpdate FileEventType = "update"
	//文件删除
	FileEventTypeRemove FileEventType = "remove"
)

type FileEvent struct {
	Type  FileEventType
	Paths []File
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
