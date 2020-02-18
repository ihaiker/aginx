package api

import (
	"github.com/ihaiker/aginx/nginx/configuration"
	"github.com/kr/pretty"
	"testing"
)

//var api = New("http://api.aginx.io")
var api = New("http://127.0.0.1:8011")

func init() {
	api.Auth("aginx", "aginx")
}

func TestAginx_Select(t *testing.T) {
	ds, err := api.Select("http", "server")
	t.Log(ds, err)
}

func TestAginx_Configuration(t *testing.T) {
	ds, err := api.Configuration()
	t.Log(ds, err)
}

func TestAginx_Add(t *testing.T) {
	{
		err := api.Delete("worker_rlimit_nofilem")
		t.Log(err)
	}
	{
		err := api.Add(Queries(), configuration.NewDirective("worker_rlimit_nofile", "8192"))
		t.Log(err)
	}

	{
		ds, err := api.Select("worker_rlimit_nofile")
		t.Log(ds)
		t.Log(err)
	}
	{
		err := api.Modify(Queries("worker_rlimit_nofile"), configuration.NewDirective("worker_rlimit_nofile", "1024"))
		t.Log(err)
	}
	{
		ds, err := api.Select("worker_rlimit_nofile")
		t.Log(ds)
		t.Log(err)
	}
}

func TestAginx_HttpUpstream(t *testing.T) {
	ds, err := api.HttpUpstream()
	t.Log(ds)
	t.Log(err)
}

func TestAginx_HttpServer(t *testing.T) {
	ds, err := api.HttpServer()
	t.Log(ds)
	t.Log(err)
}

func TestAginx_StreamUpstream(t *testing.T) {
	ds, err := api.StreamUpstream()
	t.Log(ds)
	t.Log(err)
}

func TestAginx_StreamServer(t *testing.T) {
	ds, err := api.StreamServer()
	t.Log(ds)
	t.Log(err)
}

func TestUpload(t *testing.T) {
	err := api.NewFile("password.conf", "/etc/nginx/fastcgi.conf")

	pretty.Println(err)
}
