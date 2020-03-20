package nginx_test

import (
	"github.com/go-acme/lego/v3/challenge/http01"
	"github.com/ihaiker/aginx/nginx"
	"testing"
)

func TestQuery(t *testing.T) {

	location := nginx.NewDirective("location", http01.ChallengePath("test"))
	location.AddBody("add_header", "Content-Type", `"text/plain"`)
	location.AddBody("return", "200", `"test"`)

	t.Log(location)

	server := &nginx.Directive{
		Name: "server",
		Body: []*nginx.Directive{
			nginx.NewDirective("server_name", "shui.renzhen.la"),
			nginx.NewDirective("proxy_set_header", "Host", "$host"),
			nginx.NewDirective("proxy_set_header", "X-Real-IP", "$remote_addr"),
			nginx.NewDirective("proxy_set_header", "X-Forwarded-For", "$proxy_add_x_forwarded_for"),
			location,
		},
	}

	t.Log(server)

}
