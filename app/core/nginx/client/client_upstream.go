package client

import (
	"fmt"
	"github.com/ihaiker/aginx/v2/api"
	"github.com/ihaiker/aginx/v2/core/nginx/config"
	"github.com/ihaiker/aginx/v2/core/nginx/query"
	"github.com/ihaiker/aginx/v2/core/util"
	"github.com/ihaiker/aginx/v2/core/util/errors"
	"strconv"
	"strings"
	"time"
)

func Upstream2Config(u *api.Upstream) *config.Directive {
	conf := config.New("upstream", u.Name)
	if u.Commit != "" {
		conf.AddBody("#", u.Commit)
	}
	if u.LoadStrategy != "" {
		conf.AddBody(u.LoadStrategy)
	}
	for _, server := range u.Servers {
		s := conf.AddBody("server", server.HostAndPort.String())
		if server.Weight != 0 {
			s.AddArgs(fmt.Sprintf("weight=%d", server.Weight))
		}
		if server.FailTimeout != 0 {
			s.AddArgs(fmt.Sprintf("fail_timeout=%ds", server.FailTimeout))
		}
		if server.MaxFails != 0 {
			s.AddArgs(fmt.Sprintf("max_fails=%d", server.MaxFails))
		}
		if server.Status != "" {
			s.AddArgs(server.Status)
		}
	}
	conf.AddBodyDirective(u.Parameters...)
	return conf
}

func (a *client) analysisUpstream(conf *config.Directive) *api.Upstream {
	up := new(api.Upstream)
	up.Name = conf.Args[0]
	up.Parameters = make([]*config.Directive, 0)
	for idx, d := range conf.Body {
		switch d.Name {
		case "#":
			if idx == 0 {
				up.Commit = d.Args[0]
			}
		case "server":
			server := new(api.UpstreamServer)
			server.HostAndPort = api.ParseHostAndPort(d.Args[0])
			for _, param := range d.Args[1:] {
				name, value := util.Split2(param, "=")
				switch name {
				case "weight":
					server.Weight, _ = strconv.Atoi(value)
				case "max_fails":
					server.MaxFails, _ = strconv.Atoi(value)
				case "fail_timeout":
					t, _ := time.ParseDuration(value)
					server.FailTimeout, _ = strconv.Atoi(strconv.FormatFloat(t.Seconds(), 'f', 0, 64))
				default:
					server.Status = name
				}
			}
			up.Servers = append(up.Servers, *server)
		case "ip_hash", "fair", "url_hash", "least_conn", "least_time", "hash", "sticky":
			if len(d.Body) == 0 && len(d.Args) == 0 {
				up.LoadStrategy = d.Name
				continue
			}
			fallthrough
		default:
			up.Parameters = append(up.Parameters, d)
		}
	}
	return up
}

func (a *client) filterStream(filter *api.Filter, upstream *api.Upstream) bool {
	if filter == nil {
		return true
	}
	matched := filter.Name == ""
	if filter.Name != "" {
		if filter.ExactMatch {
			matched = upstream.Name == filter.Name
		} else {
			matched = strings.Contains(upstream.Name, filter.Name)
		}
	}
	if filter.Commit != "" {
		matched = matched && strings.Contains(upstream.Commit, filter.Commit)
	}
	if filter.Protocol != "" {
		matched = matched && upstream.Protocol == filter.Protocol
	}
	return matched
}
func (a *client) GetUpstream(filter *api.Filter) (upstreams []*api.Upstream, err error) {
	var conf *config.Configuration
	if conf, err = a.Configuration(); err != nil {
		return
	}
	upstreams = make([]*api.Upstream, 0)

	protocols := map[string]api.Protocol{
		"http":   api.ProtocolHTTP,
		"stream": api.ProtocolTCP,
	}

	for name, protocol := range protocols {
		//http.upstream下查找
		if selects, err := query.Selects(conf, name, "upstream"); err == nil {
			for _, directive := range selects {
				upstream := a.analysisUpstream(directive)
				upstream.Queries = []string{name, fmt.Sprintf("upstream('%s')", upstream.Name)}
				upstream.Protocol = protocol
				if a.filterStream(filter, upstream) {
					upstreams = append(upstreams, upstream)
				}
			}
		}

		//http include
		if includes, err := query.Selects(conf, name, "include"); err == nil {
			for _, include := range includes {
				match := include.Args[0]
				for _, file := range include.Body {
					fileName := file.Args[0]
					queries := []string{
						name, fmt.Sprintf("include('%s')", match),
						fmt.Sprintf("file('%s')", fileName), "upstream",
					}
					if selects, err := query.Selects(conf, queries...); err == nil {
						for _, directive := range selects {
							upstream := a.analysisUpstream(directive)
							upstream.Queries = append(queries[:3], fmt.Sprintf("upstream('%s')", upstream.Name))
							upstream.Protocol = protocol
							if a.filterStream(filter, upstream) {
								upstreams = append(upstreams, upstream)
							}
						}
					}
				}
			}
		}
	}
	return
}

func (a *client) SetUpstream(upstream *api.Upstream) ([]string, error) {
	conf := Upstream2Config(upstream)
	//更新
	if upstream.Queries != nil && len(upstream.Queries) != 0 {
		return upstream.Queries, a.Directive().Modify(upstream.Queries, conf)
	}

	path := ""
	if upstream.Protocol == api.ProtocolHTTP {
		if _, err := a.Directive().Select("http", "include('hosts.d/*.conf')"); err != nil {
			include := config.New("include", "hosts.d/*.conf")
			if err = a.Directive().Add([]string{"http"}, include); err != nil {
				return nil, err
			}
		}
		path = fmt.Sprintf("hosts.d/%s.conf", upstream.Name)
		upstream.Queries = []string{
			"http", "include('hosts.d/*.conf')",
			fmt.Sprintf("file('%s')", path),
			fmt.Sprintf("upstream('%s')", upstream.Name),
		}
	} else {
		_, err := a.Directive().Select("stream")
		if errors.IsNotFound(err) {
			stream := config.New("stream")
			stream.AddBody("include", "stream.d/*.conf")
			if err = a.Directive().Add([]string{}, stream); err != nil {
				return nil, err
			}
		} else {
			_, err = a.Directive().Select("stream", "include('stream.d/*.conf')")
			if errors.IsNotFound(err) {
				include := config.New("include", "stream.d/*.conf")
				if err = a.Directive().Add([]string{"stream"}, include); err != nil {
					return nil, err
				}
			}
		}
		path = fmt.Sprintf("stream.d/%s.conf", upstream.Name)

		upstream.Queries = []string{
			"http", "include('stream.d/*.conf')",
			fmt.Sprintf("file('%s')", path),
			fmt.Sprintf("upstream('%s')", upstream.Name),
		}
	}

	return upstream.Queries, a.Files().NewWithContent(path, []byte(conf.String()))
}
