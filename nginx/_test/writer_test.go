package nginx_test

import (
	"github.com/ihaiker/aginx/nginx"
	"github.com/ihaiker/aginx/nginx/config"
	"github.com/ihaiker/aginx/storage"
	. "github.com/ihaiker/aginx/util"
	"path/filepath"
	"testing"
)

func TestName(t *testing.T) {
	defer Catch(func(err error) {
		t.Fatal(err)
	})
	engine := storage.FindStorage("")

	c, err := nginx.NewClient("", engine, nil, nil)
	PanicIfError(err)

	c.Add(nginx.Queries(), config.NewDirective("error_log", "logs/error.log"))
	c.Add(nginx.Queries("http", "include", "*", "server"), config.NewDirective("error_log", "logs/error.log"))

	_, conf, _ := nginx.GetInfo()
	path := filepath.Dir(conf)

	PanicIfError(nginx.Write(c.Configuration(), nginx.FileDiffer(path), nginx.FileWriter(path)))
}
