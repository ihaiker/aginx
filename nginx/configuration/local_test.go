package configuration_test

import (
	"bytes"
	"fmt"
	"github.com/ihaiker/aginx/nginx"
	"github.com/ihaiker/aginx/nginx/client"
	"github.com/ihaiker/aginx/nginx/configuration"
	"github.com/ihaiker/aginx/storage/file"
	. "github.com/ihaiker/aginx/util"
	"io/ioutil"
	"path/filepath"
	"testing"
)

func writer(root string) configuration.Writer {
	return func(file string, content []byte) error {
		fp := filepath.Join(root, file)
		fmt.Println("write file ", fp)
		return WriteFile(fp, content)
	}
}

func differ(root string) configuration.Differ {
	return func(file string, content []byte) bool {
		if bs, err := ioutil.ReadFile(filepath.Join(root, file)); err == nil {
			return !bytes.Equal(bs, content)
		}
		return true
	}
}

func TestName(t *testing.T) {
	defer Catch(func(err error) {
		t.Fatal(err)
	})
	engine, err := file.System()
	PanicIfError(err)

	c, err := client.NewClient(engine)
	PanicIfError(err)

	c.Add(client.Queries(), nginx.NewDirective("error_log", "logs/error.log"))
	c.Add(client.Queries("http", "include", "*", "server"), nginx.NewDirective("error_log", "logs/error.log"))
	_, conf, _ := file.GetInfo()
	path := filepath.Dir(conf)
	configuration.DownWriterDiffer(c.Configuration(), writer(path), differ(path))
}
