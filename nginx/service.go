package nginx

import (
	"fmt"
	"github.com/ihaiker/aginx/api"
	"github.com/ihaiker/aginx/nginx/config"
	"github.com/ihaiker/aginx/util"
	"net/url"
	"path/filepath"
	"strings"
)

type Service struct {
	*Client
}

func NewService(client *Client) *Service {
	return &Service{Client: client}
}

func (service *Service) ListService() []*api.Server {
	servers := subServerFile(service.doc.Name, service.doc.Body)
	upstreams := subUpstreamFile(service.doc.Name, service.doc.Body)
	for _, server := range servers {
		for _, location := range server.Locations {
			if location.LoadBalance != nil && location.LoadBalance.Upstream != nil {
				location.LoadBalance.Upstream = upstreams.Get(location.LoadBalance.Upstream.Name)
			}
		}
	}
	return servers
}

func (service *Service) includeArgs(arg string) string {
	path := ""
	includes, err := service.Select("http", "include")
	util.PanicIfError(err)
	for _, i := range includes {
		if m, _ := filepath.Match(i.Args[0], arg); m {
			path = i.Args[0]
		}
	}
	return path
}

func (service *Service) DeleteService(name string) error {
	server := service.getServer(name)
	if server == nil {
		return fmt.Errorf("not found server %s", name)
	}

	paths := []string{"http"}
	if server.From != "nginx.conf" {
		path := service.includeArgs(server.From)
		if path == "" {
			return fmt.Errorf("not found server %s", name)
		}
		paths = append(paths, fmt.Sprintf("include('%s')", path),
			fmt.Sprintf("file('%s')", server.From))
	}

	for _, location := range server.Locations {
		if location.Type == "balance" {
			up := fmt.Sprintf("upstream('%s')", location.LoadBalance.Upstream.Name)
			if err := service.Delete(append(paths, up)...); err != nil {
				return err
			}
		}
	}
	return service.Delete(append(paths, fmt.Sprintf("server.server_name('%s')", name))...)
}

func (service *Service) makeServer(s *api.Server) []*config.Directive {
	var server *config.Directive = config.NewDirective("server")
	var upstream *config.Directive = nil
	for _, listen := range s.Listen {
		server.AddBody("listen", listen...)
	}
	server.AddBody("server_name", s.ServerName...)
	if s.Root != "" {
		server.AddBody("root", s.Root)
	}
	if s.Index != nil && len(s.Index) > 0 {
		server.AddBody("index", s.Index...)
	}
	for _, attr := range s.Attrs {
		server.AddBody(attr.Name, attr.Attrs...)
	}
	for _, location := range s.Locations {
		loc := server.AddBody("location", location.Path...)
		switch location.Type {
		case "root":
			loc.AddBody("root", location.Root)
			if len(location.Index) > 0 {
				loc.AddBody("index", location.Index...)
			}
		case "proxy":
			loc.AddBody("proxy_pass",
				fmt.Sprintf("%s://%s", location.LoadBalance.ProxyType, location.LoadBalance.ProxyAddress))
		case "balance":
			loc.AddBody("proxy_pass",
				fmt.Sprintf("%s://%s", location.LoadBalance.ProxyType, location.LoadBalance.Upstream.Name))
			upstream = config.NewDirective("upstream", location.LoadBalance.Upstream.Name)
			for _, attr := range location.LoadBalance.Upstream.Attrs {
				upstream.AddBody(attr.Name, attr.Attrs...)
			}
			for _, item := range location.LoadBalance.Upstream.Items {
				upstream.AddBody("server", append([]string{item.Server}, item.Attrs...)...)
			}
		}
		for _, attr := range location.Attrs {
			loc.AddBody(attr.Name, attr.Attrs...)
		}
	}
	if upstream != nil {
		return []*config.Directive{upstream, server}
	} else {
		return []*config.Directive{server}
	}
}

func (service *Service) ModifyService(names []string, s *api.Server) {
	util.AssertTrue(len(s.ServerName) > 0, "server name not found!")
	for _, name := range names {
		_ = service.DeleteService(name)
	}

	from := fmt.Sprintf("hosts.d/%s.conf", s.ServerName[0])
	includeDir := "hosts.d/*.conf"
	if s.From != "" && s.From != "nginx.conf" {
		from = s.From
		includeDir = service.includeArgs(s.From)
	}
	service.hostsd(includeDir)

	file := config.NewDirective("file", from)
	file.Virtual = config.Include
	if files, err := service.Select("http",
		fmt.Sprintf("include('%s')", includeDir),
		fmt.Sprintf("file('%s')", from)); err == nil && len(files) > 0 {
		file = files[0]
	} else {
		dir := service.MustSelect("http", fmt.Sprintf("include('%s')", includeDir))[0]
		dir.AddBodyDirective(file)
	}
	file.AddBodyDirective(service.makeServer(s)...)
}

func (service *Service) getServer(name string) *api.Server {
	services := service.ListService()
	for _, s := range services {
		for _, sn := range s.ServerName {
			if sn == name {
				return s
			}
		}
	}
	return nil
}

func (service *Service) ListUpstream() api.Upstreams {
	return subUpstreamFile(service.doc.Name, service.doc.Body)
}

func subUpstreamFile(filename string, items config.Directives) api.Upstreams {
	upstreams := make([]*api.Upstream, 0)
	for _, item := range items {
		switch item.Name {
		case "upstream":
			upstream := &api.Upstream{
				From: filename, Name: item.Args[0],
			}
			for _, d := range item.Body {
				if d.Name == "server" {
					upstream.Items = append(upstream.Items, &api.UpstreamItem{
						Server: d.Args[0],
						Attrs:  d.Args[1:],
					})
				} else {
					upstream.Attrs = append(upstream.Attrs, &api.Attr{
						Name: d.Name, Attrs: d.Args,
					})
				}
			}
			upstreams = append(upstreams, upstream)
		case "server":
		case "include":
			for _, include := range item.Body {
				ups := subUpstreamFile(include.Args[0], include.Body)
				upstreams = append(upstreams, ups...)
			}
		default:
			ups := subUpstreamFile(filename, item.Body)
			upstreams = append(upstreams, ups...)
		}
	}
	return upstreams
}

func subServerFile(filename string, items config.Directives) []*api.Server {
	ss := make([]*api.Server, 0)
	for _, item := range items {
		switch item.Name {
		case "upstream":

		case "server":
			ss = append(ss, server(filename, item))
		case "include":
			for _, include := range item.Body {
				ss = append(ss, subServerFile(include.Args[0], include.Body)...)
			}
		default:
			ss = append(ss, subServerFile(filename, item.Body)...)
		}
	}
	return ss
}

func serverLocation(item *config.Directive) *api.ServerLocation {
	l := &api.ServerLocation{}
	l.Path = item.Args
	for _, ld := range item.Body {
		switch ld.Name {
		default:
			l.Attrs = append(l.Attrs, &api.Attr{
				Name: ld.Name, Attrs: ld.Args,
			})
		case "root":
			l.Root = ld.Args[0]
		case "index":
			l.Index = ld.Args
		case "proxy_pass":
			u, _ := url.Parse(ld.Args[0])
			l.LoadBalance = &api.ServerLocationLoadBalance{
				ProxyType: u.Scheme,
			}
			if strings.Contains(u.Host, ":") {
				l.LoadBalance.ProxyAddress = u.Host
			} else {
				l.LoadBalance.Upstream = &api.Upstream{
					Name: u.Host,
				}
			}
		}
	}
	if l.Root != "" {
		l.Type = "root"
	} else if l.LoadBalance != nil {
		if l.LoadBalance.Upstream != nil {
			l.Type = "balance"
		} else {
			l.Type = "proxy"
		}
	} else {
		l.Type = "empty"
	}
	return l
}

func server(filename string, item *config.Directive) *api.Server {
	s := &api.Server{From: filename}
	for _, d := range item.Body {
		switch d.Name {
		case "#":
		case "server_name":
			s.ServerName = d.Args
		case "listen":
			s.Listen = append(s.Listen, d.Args)
		case "location":
			s.Locations = append(s.Locations, serverLocation(d))
		default:
			s.Attrs = append(s.Attrs, &api.Attr{
				Name: d.Name, Attrs: d.Args,
			})
		}
	}
	return s
}
