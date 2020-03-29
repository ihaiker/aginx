package nginx

import (
	"fmt"
	"github.com/ihaiker/aginx/nginx/config"
	"strings"
)

func UpstreamName(domain string) string {
	name := strings.ReplaceAll(domain, ".", "_")
	name = strings.ReplaceAll(name, "*", "_x_")
	return name
}

func SimpleUpstream(name string, address ...string) *config.Directive {
	directive := config.NewDirective("upstream", name)
	if len(address) == 0 {
		directive.AddBody("server", "127.0.0.1:65535")
	} else {
		for _, s := range address {
			directive.AddBody("server", s)
		}
	}
	return directive
}

func SimpleUpstreamWithWeight(name string, address map[int]string) *config.Directive {
	directive := config.NewDirective("upstream", name)
	if len(address) == 0 {
		directive.AddBody("server", "127.0.0.1:65535")
	} else {
		for weight, server := range address {
			directive.AddBody("server", server, fmt.Sprintf("weight=%d", weight))
		}
	}
	return directive
}

func SimpleServer(domain string, address ...string) (upstream *config.Directive, server *config.Directive) {
	name := UpstreamName(domain)
	upstream = SimpleUpstream(name, address...)

	server = config.NewDirective("server")
	server.AddBody("listen", "80")
	server.AddBody("server_name", domain)
	location := server.AddBody("location", "/")
	{
		location.AddBody("proxy_pass", fmt.Sprintf("http://%s", name))
		location.AddBody("proxy_set_header", "Host", domain)
		location.AddBody("proxy_set_header", "X-Real-IP", "$remote_addr")
		location.AddBody("proxy_set_header", "X-Forwarded-For", "$proxy_add_x_forwarded_for")
	}
	return
}
