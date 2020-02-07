package client

import (
	"fmt"
	"github.com/ihaiker/aginx/storage/file"
	"path/filepath"
	"testing"
)

func TestNginxFull(t *testing.T) {
	conf, _ := filepath.Abs("../../_test/nginx.conf")
	fileStore := file.New(conf)
	doc, err := Readable(fileStore)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(doc.Directive().Json())
}

func TestSystem(t *testing.T) {
	fileStore, err := file.System()
	if err != nil {
		t.Fatal(err)
	}
	conf, err := Readable(fileStore)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(conf.Directive().Json())
}
