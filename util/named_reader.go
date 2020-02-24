package util

import (
	"bytes"
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

func (nr *NameReader) Bytes() []byte {
	if out, match := nr.Reader.(*bytes.Buffer); match {
		return out.Bytes()
	} else {
		bs, _ := ioutil.ReadAll(nr.Reader)
		return bs
	}
}

func MapNamedReader(list []*NameReader) map[string]*NameReader {
	out := map[string]*NameReader{}
	for _, reader := range list {
		out[reader.Name] = reader
	}
	return out
}
