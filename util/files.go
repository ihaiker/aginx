package util

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

func WriterFile(path string, content []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}
	return ioutil.WriteFile(path, content, 0666)
}

func WriterReader(path string, reader io.Reader) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}
	bs, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, bs, 0666)
}
