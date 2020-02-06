package nginx

import (
	"fmt"
	"github.com/kr/pretty"
	"os"
	"testing"
)

var api Client

func init() {
	//cwd, _ := os.Getwd()
	//root, _ := filepath.Abs(cwd + "/../_test")
	//doc, err := AnalysisFromFile(root, "nginx.conf")
	doc, err := AnalysisNginx()
	if err != nil {
		os.Exit(0)
	}
	api = NewClient(doc)
}

func show(t *testing.T, query ...string) {
	directives, err := api.Select(query...)
	if err != nil {
		t.Fatal(err)
	}
	for _, server := range directives {
		fmt.Println(server.Pretty(0))
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
	access_log := NewDirective("access_log", "logs/domain2.access.log", "main")
	headers := []*Directive{
		access_log,
		NewDirective("proxy_set_header", "Host", "$host"),
		NewDirective("proxy_set_header", "X-Real-IP", "$remote_addr"),
		NewDirective("proxy_set_header", "X-Forwarded-For", "$proxy_add_x_forwarded_for"),
	}

	finder := Queries("http", "server.[server_name('domain1.com') & listen('443')]")

	err := api.Add(finder, headers...)
	if err != nil {
		t.Log(err)
	}
	TestClientHttpServerArray(t)
}

func TestClient_Delete(t *testing.T) {
	finder := Queries("http", "server.[server_name('domain1.com') & listen('443')]", "root")

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
	server := &Directive{
		Name: "server",
		Body: []*Directive{
			NewDirective("server_name", "shui.renzhen.la"),
			NewDirective("proxy_set_header", "Host", "$host"),
			NewDirective("proxy_set_header", "X-Real-IP", "$remote_addr"),
			NewDirective("proxy_set_header", "X-Forwarded-For", "$proxy_add_x_forwarded_for"),
		},
	}
	if err := api.Add(Queries("http"), server); err != nil {
		t.Fatal(err)
	}

	shui, _ := api.Select("http", "server.server_name('shui.renzhen.la')")
	_, _ = pretty.Println(shui)

	api.Delete("http", "server.server_name('shui.renzhen.la')", "proxy_set_header")

	shui, _ = api.Select("http", "server.server_name('shui.renzhen.la')")
	_, _ = pretty.Println(shui)

	modifyDirective := NewDirective("server_name", "who.renzhen.la")
	if err := api.Modify(Queries("http", "server", "server_name('shui.renzhen.la')"), modifyDirective); err != nil {
		t.Fatal(err)
	}

	shui, _ = api.Select("http", "server.server_name('who.renzhen.la')")
	_, _ = pretty.Println(shui)

}

func TestSelectInclude(t *testing.T) {
	servers, _ := api.Select("http", "include", "*", "server.server_name('localhost')")

	for _, server := range servers {
		fmt.Println(server.Pretty(0))
	}
}
