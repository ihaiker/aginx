package api

import (
	"github.com/ihaiker/aginx/nginx"
	"github.com/kr/pretty"
	"testing"
)

//var api = New("http://api.aginx.io")
var api = New("http://127.0.0.1:8011")

func init() {
	api.Auth("aginx", "aginx")
}

func TestAginx_Select(t *testing.T) {
	ds, err := api.Directive().Select("http", "server")
	t.Log(ds, err)
}

func TestAginx_Configuration(t *testing.T) {
	ds, err := api.Configuration()
	t.Log(ds, err)
}

func TestAginx_Add(t *testing.T) {
	{
		err := api.Directive().Delete("worker_rlimit_nofilem")
		t.Log(err)
	}
	{
		err := api.Directive().Add(Queries(), nginx.NewDirective("worker_rlimit_nofile", "8192"))
		t.Log(err)
	}

	{
		ds, err := api.Directive().Select("worker_rlimit_nofile")
		t.Log(ds)
		t.Log(err)
	}
	{
		err := api.Directive().Modify(Queries("worker_rlimit_nofile"), nginx.NewDirective("worker_rlimit_nofile", "1024"))
		t.Log(err)
	}
	{
		ds, err := api.Directive().Select("worker_rlimit_nofile")
		t.Log(ds)
		t.Log(err)
	}
}

func TestAginx_HttpUpstream(t *testing.T) {
	ds, err := api.Directive().HttpUpstream()
	t.Log(ds)
	t.Log(err)
}

func TestAginx_HttpServer(t *testing.T) {
	ds, err := api.Directive().HttpServer()
	t.Log(ds)
	t.Log(err)
}

func TestAginx_StreamUpstream(t *testing.T) {
	ds, err := api.Directive().StreamUpstream()
	t.Log(ds)
	t.Log(err)
}

func TestAginx_StreamServer(t *testing.T) {
	ds, err := api.Directive().StreamServer()
	t.Log(ds)
	t.Log(err)
}

func TestUpload(t *testing.T) {
	err := api.File().New("password.conf", "/etc/nginx/fastcgi.conf")
	pretty.Println(err)
}
