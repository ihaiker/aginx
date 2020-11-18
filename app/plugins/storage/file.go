package storage

//存储插件存储文件内容
type File struct {
	Name    string `json:"name"`
	Content []byte `json:"content"`
}

func (cfg *File) String() string {
	return string(cfg.Content)
}

func NewFile(name string, content []byte) *File {
	return &File{Name: name, Content: content}
}
