package nginx_test

import (
	"github.com/ihaiker/aginx/nginx"
	"github.com/ihaiker/aginx/storage"
	. "github.com/ihaiker/aginx/util"
	"path/filepath"
	"testing"
)

func TestName(t *testing.T) {
	defer Catch(func(err error) {
		t.Fatal(err)
	})
	engine := storage.FindStorage("", false)

	c, err := nginx.NewClient(engine)
	PanicIfError(err)

	c.Add(nginx.Queries(), nginx.NewDirective("error_log", "logs/error.log"))
	c.Add(nginx.Queries("http", "include", "*", "server"), nginx.NewDirective("error_log", "logs/error.log"))

	_, conf, _ := nginx.GetInfo()
	path := filepath.Dir(conf)

	PanicIfError(nginx.Write(c.Configuration(), nginx.FileDiffer(path), nginx.FileWriter(path)))
}
