package nginx_test

import (
	"fmt"
	"github.com/ihaiker/aginx/nginx"
	"github.com/ihaiker/aginx/storage/file"
	"path/filepath"
	"testing"
)

func TestNginxFull(t *testing.T) {
	conf, _ := filepath.Abs("../../bin/nginx/nginx.conf")
	t.Log(conf)
	fileStore := file.New(conf)
	doc, err := nginx.Readable(fileStore)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(doc)
}

func TestSystem(t *testing.T) {
	fileStore, err := file.System()
	if err != nil {
		t.Fatal(err)
	}
	conf, err := nginx.Readable(fileStore)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(conf)
}
