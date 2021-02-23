package tcloud

import (
	"fmt"
	"github.com/ihaiker/aginx/v2/api"
	"github.com/ihaiker/aginx/v2/core/nginx/config"
)

type verifiedProvider struct {
	aginx   api.Aginx
	queries []string
}

func newVerifiedProvider(aginx api.Aginx) *verifiedProvider {
	return &verifiedProvider{aginx: aginx}
}

func (self *verifiedProvider) present(domain, location, txt string) error {
	filter := &api.Filter{Name: domain, Protocol: api.ProtocolHTTP, ExactMatch: true}
	servers, err := self.aginx.GetServers(filter)
	if err != nil {
		return err
	}
	var server *api.Server
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
		if sl.Path == location {
			server.Locations = append(server.Locations[:i], server.Locations[i+1:]...)
			break
		}
	}
	server.Locations = append(server.Locations, api.ServerLocation{
		Type: api.ProxyCustom, Path: location,
		Parameters: []*config.Directive{
			config.New("add_header", "Content-Type", `"application/octet-stream"`),
			config.New("return", "200", fmt.Sprintf("'%s'", txt)),
		},
	})

	if self.queries, err = self.aginx.SetServer(server); err != nil {
		return err
	}
	if len(servers) != 0 {
		self.queries = append(server.Queries, fmt.Sprintf("location('%s')", location))
	}
	return nil
}

func (self *verifiedProvider) cleanUp() error {
	return self.aginx.Directive().Delete(self.queries...)
}
