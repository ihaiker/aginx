package configuration

import (
	"github.com/go-acme/lego/v3/challenge/http01"
	"testing"
)

func TestQuery(t *testing.T) {

	location := NewDirective("location", http01.ChallengePath("test"))
	location.AddBody("add_header", "Content-Type", `"text/plain"`)
	location.AddBody("return", "200", `"test"`)

	server := &Directive{
		Name: "server",
		Body: []*Directive{
			NewDirective("server_name", "shui.renzhen.la"),
			NewDirective("proxy_set_header", "Host", "$host"),
			NewDirective("proxy_set_header", "X-Real-IP", "$remote_addr"),
			NewDirective("proxy_set_header", "X-Forwarded-For", "$proxy_add_x_forwarded_for"),
			location,
		},
	}

	t.Log(location.Query())

	t.Log(server.Query())

}
