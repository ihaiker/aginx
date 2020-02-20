package util

import (
	"fmt"
	"io"
	"io/ioutil"
)

type NameReader struct {
	io.Reader
	Name string
}

func NamedReader(rd io.Reader, name string) *NameReader {
	return &NameReader{Reader: rd, Name: name}
}

func (nr *NameReader) String() string {
	return fmt.Sprintf("file(%s)", nr.Name)
}

func (nr *NameReader) ToString() string {
	bs, _ := ioutil.ReadAll(nr.Reader)
	return string(bs)
}

func MapNamedReader(list []*NameReader) map[string]*NameReader {
	out := map[string]*NameReader{}
	for _, reader := range list {
		out[reader.Name] = reader
	}
	return out
}
