package lego

import (
	"fmt"
	"github.com/go-acme/lego/v3/challenge/http01"
	"github.com/ihaiker/aginx/v2/api"
	"github.com/ihaiker/aginx/v2/core/nginx/config"
	"time"
)

type sslProvider struct {
	aginx   api.Aginx
	queries []string
}

func newProvider(aginx api.Aginx) *sslProvider {
	return &sslProvider{aginx: aginx}
}

func (self *sslProvider) Present(domain, token, keyAuth string) error {
	filter := &api.Filter{Name: domain, Protocol: api.ProtocolHTTP, ExactMatch: true}
	servers, err := self.aginx.GetServers(filter)
	if err != nil {
		return err
	}
	var server *api.Server

	locationPath := http01.ChallengePath(token)
	if len(servers) == 0 {
		server = new(api.Server)
		server.Domains = []string{domain}
		server.Listens = []api.ServerListen{
			{HostAndPort: api.HostAndPort{Port: 80}},
		}
		server.Protocol = api.ProtocolHTTP
	} else {
		server = servers[0]
	}
	for i, sl := range server.Locations {
		if sl.Path == locationPath {
			server.Locations = append(server.Locations[:i], server.Locations[i+1:]...)
			break
		}
	}
	server.Locations = append(server.Locations, api.ServerLocation{
		Type: api.ProxyCustom, Path: locationPath,
		Parameters: []*config.Directive{
			config.New("add_header", "Content-Type", `"text/plain"`),
			config.New("return", "200", fmt.Sprintf("'%s'", keyAuth)),
		},
	})
	if self.queries, err = self.aginx.SetServer(server); err != nil {
		return err
	}
	if len(servers) != 0 {
		self.queries = append(server.Queries, fmt.Sprintf("location('%s')", locationPath))
	}
	return nil
}

func (self *sslProvider) CleanUp(domain, token, keyAuth string) error {
	return self.aginx.Directive().Delete(self.queries...)
}

func (self *sslProvider) Timeout() (timeout, interval time.Duration) {
	return time.Minute, time.Second * 3
}
