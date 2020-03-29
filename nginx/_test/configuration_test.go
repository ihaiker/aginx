package nginx_test

import (
	"github.com/go-acme/lego/v3/challenge/http01"
	"github.com/ihaiker/aginx/nginx/config"
	"testing"
)

func TestQuery(t *testing.T) {

	location := config.NewDirective("location", http01.ChallengePath("test"))
	location.AddBody("add_header", "Content-Type", `"text/plain"`)
	location.AddBody("return", "200", `"test"`)

	t.Log(location)

	server := &config.Directive{
		Name: "server",
		Body: []*config.Directive{
			config.NewDirective("server_name", "shui.renzhen.la"),
			config.NewDirective("proxy_set_header", "Host", "$host"),
			config.NewDirective("proxy_set_header", "X-Real-IP", "$remote_addr"),
			config.NewDirective("proxy_set_header", "X-Forwarded-For", "$proxy_add_x_forwarded_for"),
			location,
		},
	}

	t.Log(server)

}
