package nginx_test

import (
	"fmt"
	"github.com/ihaiker/aginx/logs"
	"github.com/ihaiker/aginx/nginx"
	"github.com/ihaiker/aginx/nginx/config"
	"github.com/ihaiker/aginx/storage/file"
	"github.com/kr/pretty"
	"github.com/sirupsen/logrus"
	"testing"
)

func init() {
	logs.SetLevel(logrus.DebugLevel)
}

var api = nginx.MustClient("", file.New(nginx.MustConf()), nil, nil)

func show(t *testing.T, query ...string) {

	directives, err := api.Select(query...)
	if err != nil {
		t.Fatal(err)
	}
	for _, server := range directives {
		fmt.Println(server)
	}
}

func TestClient(t *testing.T) {
	show(t)
}

func TestClientSignal(t *testing.T) {
	show(t, "user")
}

func TestClientHttpServer(t *testing.T) {
	show(t, "http", "server", "server_name")
}

func TestClientHttpSelectName(t *testing.T) {
	show(t, "http", "server", "server_name('domain1.com')")
}

func TestClientHttpServerByName(t *testing.T) {
	show(t, "http", "server.server_name('domain1.com')")
}

func TestClientHttpServerArray(t *testing.T) {
	show(t, "http", "server.[server_name(@'domain') & listen('443')]")
}

func TestClientHttpServer3Array(t *testing.T) {
	show(t, "http", "server.[server_name('domain1.com') & listen('443')]")
}

func TestClientStream(t *testing.T) {
	show(t, "stream", "upstream('backend')")
}

func TestClient_Add(t *testing.T) {
	access_log := config.NewDirective("access_log", "logs/domain2.access.log", "main")
	headers := []*config.Directive{
		access_log,
		config.NewDirective("proxy_set_header", "Host", "$host"),
		config.NewDirective("proxy_set_header", "X-Real-IP", "$remote_addr"),
		config.NewDirective("proxy_set_header", "X-Forwarded-For", "$proxy_add_x_forwarded_for"),
	}

	finder := nginx.Queries("http", "server.[server_name('domain1.com') & listen('443')]")

	err := api.Add(finder, headers...)
	if err != nil {
		t.Log(err)
	}
	TestClientHttpServerArray(t)
}

func TestClient_Delete(t *testing.T) {
	finder := nginx.Queries("http", "server.[server_name('domain1.com') & listen('443')]", "root")

	if err := api.Delete(finder...); err != nil {
		t.Fatal(err)
	}

	out, err := api.Select(finder[0 : len(finder)-1]...)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = pretty.Println(out)
}

func TestClientAll(t *testing.T) {
	server := &config.Directive{
		Name: "server",
		Body: []*config.Directive{
			config.NewDirective("server_name", "shui.renzhen.la"),
			config.NewDirective("proxy_set_header", "Host", "$host"),
			config.NewDirective("proxy_set_header", "X-Real-IP", "$remote_addr"),
			config.NewDirective("proxy_set_header", "X-Forwarded-For", "$proxy_add_x_forwarded_for"),
		},
	}
	if err := api.Add(nginx.Queries("http"), server); err != nil {
		t.Fatal(err)
	}

	shui, _ := api.Select("http", "server.server_name('shui.renzhen.la')")
	_, _ = pretty.Println(shui)

	api.Delete("http", "server.server_name('shui.renzhen.la')", "proxy_set_header")

	shui, _ = api.Select("http", "server.server_name('shui.renzhen.la')")
	_, _ = pretty.Println(shui)

	modifyDirective := config.NewDirective("server_name", "who.renzhen.la")
	if err := api.Modify(nginx.Queries("http", "server", "server_name('shui.renzhen.la')"), modifyDirective); err != nil {
		t.Fatal(err)
	}

	shui, _ = api.Select("http", "server.server_name('who.renzhen.la')")
	_, _ = pretty.Println(shui)

}

func TestSelectInclude(t *testing.T) {
	servers, err := api.Select("http", "include('reg.d/*.conf')", "file('reg.d/api.aginx.io.ngx.conf')")
	if err != nil {
		t.Fatal(err)
	}
	for _, server := range servers {
		fmt.Println(server)
	}
}

func TestServers(t *testing.T) {
	servers, err := api.Select("http", "include", "*", "server")
	if err != nil {
		t.Fatal(err)
	}
	for _, server := range servers {
		fmt.Println(server)
	}
}

func TestClient_SimpleServer(t *testing.T) {
	err := api.SimpleServer("test.aginx.io", false, "127.0.0.1:8002")
	if err != nil {
		t.Fatal(err)
	}
}
