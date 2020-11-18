package client

import (
	"fmt"
	"github.com/ihaiker/aginx/v2/api"
	"github.com/ihaiker/aginx/v2/core/nginx/config"
	"github.com/ihaiker/aginx/v2/core/nginx/query"
	"github.com/ihaiker/aginx/v2/core/util/errors"
	"net/url"
	"strings"
)

func containUpstream(upstreams []*api.Upstream, upstreamName string) bool {
	for _, upstream := range upstreams {
		if upstream.Name == upstreamName {
			return true
		}
	}
	return false
}

func analysisServer(selects *config.Directive, upstreams []*api.Upstream) *api.Server {
	server := new(api.Server)
	server.Parameters = make([]*config.Directive, 0)
	for i, d := range selects.Body {
		switch d.Name {
		case "#":
			if i == 0 {
				server.Commit = strings.Join(d.Args, " ")
			}
		case "listen":
			listen := api.ServerListen{}
			listen.HostAndPort = api.ParseHostAndPort(d.Args[0])
			for _, arg := range d.Args[1:] {
				if arg == "default" {
					listen.Default = true
				} else if arg == "ssl" {
					listen.SSL = true
				} else if arg == "http2" {
					listen.HTTP2 = true
				} else if arg == "udp" {
					server.Protocol = api.ProtocolUDP
				}
			}
			server.Listens = append(server.Listens, listen)
		case "server_name":
			server.Domains = append(server.Domains, d.Args...)
		case "auth_basic", "auth_basic_user_file":
			if server.AuthBasic == nil {
				server.AuthBasic = new(api.AuthBasic)
			}
			if d.Name == "auth_basic" {
				server.AuthBasic.Switch = d.Args[0]
			} else if d.Name == "auth_basic_user_file" {
				server.AuthBasic.UserFile = d.Args[0]
			}
		case "ssl_certificate", "ssl_certificate_key", "ssl_protocols":
			if server.SSL == nil {
				server.SSL = new(api.ServerSSL)
			}
			if d.Name == "ssl_certificate" {
				server.SSL.Certificate = d.Args[0]
			} else if d.Name == "ssl_certificate_key" {
				server.SSL.CertificateKey = d.Args[0]
			} else {
				server.SSL.Protocols = strings.Join(d.Args, " ")
			}
		case "proxy_pass": //tcp/udp
			server.ProxyPass = strings.Join(d.Args, " ")
		case "location":
			location := api.ServerLocation{}
			location.Path = strings.Join(d.Args, " ")
			location.Type = api.ProxyCustom
			location.Parameters = make([]*config.Directive, 0)
		LOCATION:
			for j, lbc := range d.Body {
				switch lbc.Name {
				case "#":
					if j == 0 {
						location.Commit = lbc.Args[0]
					}
				case "root", "alias", "index":
					location.Type = api.ProxyHTML
					if location.HTML == nil {
						location.HTML = new(api.ServerLocationHTML)
					}
					if lbc.Name == "root" {
						location.HTML.Model = "root"
						location.HTML.Path = lbc.Args[0]
					} else if lbc.Name == "alias" {
						location.HTML.Model = "alias"
						location.HTML.Path = lbc.Args[0]
					} else {
						location.HTML.Indexes = strings.Join(lbc.Args, " ")
					}
				case "auth_basic", "auth_basic_user_file":
					if location.AuthBasic == nil {
						location.AuthBasic = new(api.AuthBasic)
					}
					if lbc.Name == "auth_basic_user_file" {
						location.AuthBasic.UserFile = lbc.Args[0]
					} else if lbc.Name == "auth_basic" {
						location.AuthBasic.Switch = lbc.Args[0]
					}
				case "proxy_pass":
					if proxyPass, err := url.Parse(lbc.Args[0]); err != nil {
						location.Type = api.ProxyUpstream
						location.Upstream = &api.ServerLocationUpstream{Name: lbc.Args[0]}
					} else {
						host := proxyPass.Host
						if containUpstream(upstreams, host) {
							location.Type = api.ProxyUpstream
							location.Upstream = &api.ServerLocationUpstream{Name: host}
							if proxyPass.Path != "" && proxyPass.Path != "/" {
								location.Upstream.Path = proxyPass.Path
							}
						} else {
							location.Type = api.ProxyHTTP
							location.HTTP = &api.ServerLocationHTTP{To: lbc.Args[0]}
						}
					}
				case "allow":
					location.Allows = append(location.Allows, d.Args...)
				case "deny":
					location.Denys = append(location.Denys, d.Args...)
				case "proxy_set_header":
					switch lbc.Args[0] {
					case "Upgrade", "Connection":
						location.WebSocket = true
						continue LOCATION
					case "X-Scheme", "Host", "X-Real-IP", "X-Forwarded-For":
						location.BasicHeader = true
						continue LOCATION
					}
					fallthrough
				default:
					location.Parameters = append(location.Parameters, lbc)
				}
			}
			server.Locations = append(server.Locations, location)
		case "allow":
			server.Allows = append(server.Allows, d.Args...)
		case "deny":
			server.Denys = append(server.Denys, d.Args...)
		case "if":
			if strings.Join(d.Args, " ") == "( $scheme = 'http' )" {
				if server.SSL == nil {
					server.SSL = new(api.ServerSSL)
				}
				server.SSL.HTTPRedirect = true
				break
			} else if len(d.Args) > 4 && d.Args[1] == "$http_user_agent" && d.Args[2] == "~" {
				if rewrite := d.Body.Get("rewrite"); rewrite != nil {
					server.RewriteMobile = new(api.RewriteMobile)
					server.RewriteMobile.Agents = d.Args[3]
					server.RewriteMobile.Domain = rewrite.Args[1]
					break
				}
			}
			fallthrough
		default:
			server.Parameters = append(server.Parameters, d)
		}
	}
	return server
}

func Server2Config(s *api.Server) *config.Directive {
	conf := config.New("server")
	if s.Commit != "" {
		conf.AddBody("#", s.Commit)
	}
	for _, listen := range s.Listens {
		body := conf.AddBody("listen", listen.String())
		if listen.Default {
			body.AddArgs("default")
		}
		if listen.SSL {
			body.AddArgs("ssl")
		}
		if listen.HTTP2 {
			body.AddArgs("http2")
		}
		if s.Protocol == api.ProtocolUDP {
			body.AddArgs("udp")
		}
	}

	if s.Protocol == api.ProtocolHTTP && (s.Domains != nil || len(s.Domains) != 0) {
		conf.AddBody("server_name", s.Domains...)
	}

	if s.Protocol == api.ProtocolHTTP && s.RewriteMobile != nil {
		if s.RewriteMobile.Agents == "" {
			s.RewriteMobile.Agents = "'(MIDP)|(WAP)|(UP.Browser)|(Smartphone)|(Obigo)|(Mobile)|(AU.Browser)|(wxd.Mms)|" +
				"(WxdB.Browser)|(CLDC)|(UP.Link)|(KM.Browser)|(UCWEB)|(SEMC-Browser)|" +
				"(Mini)|(Symbian)|(Palm)|(Nokia)|(Panasonic)|(MOT-)|(SonyEricsson)|" +
				"(NEC-)|(Alcatel)|(Ericsson)|(BENQ)|(BenQ)|(Amoisonic)|(Amoi-)|" +
				"(Capitel)|(PHILIPS)|(SAMSUNG)|(Lenovo)|(Mitsu)|(Motorola)|(SHARP)|" +
				"(WAPPER)|(LG-)|(LG/)|(EG900)|(CECT)|(Compal)|(kejian)|(Bird)|(BIRD)|(G900/V1.0)|" +
				"(Arima)|(CTL)|(TDG)|(Daxian)|(DAXIAN)|(DBTEL)|(Eastcom)|(EASTCOM)|(PANTECH)|" +
				"(Dopod)|(Haier)|(HAIER)|(KONKA)|(KEJIAN)|(LENOVO)|(Soutec)|(SOUTEC)|(SAGEM)|" +
				"(SEC-)|(SED-)|(EMOL-)|(INNO55)|(ZTE)|(iPhone)|(Android)|(Windows CE)|(Wget)|" +
				"(Java)|(curl)|(Opera)'"
		}
		if s.RewriteMobile.Agents[0] != '\'' && s.RewriteMobile.Agents[0] != '"' {
			s.RewriteMobile.Agents = "\"" + s.RewriteMobile.Agents + "\""
		}
		rewrite := conf.AddBody("if", "(", "$http_user_agent", "~", s.RewriteMobile.Agents, ")")
		rewrite.AddBody("rewrite", "^/(.*)$", fmt.Sprintf("%s", s.RewriteMobile.Domain), "permanent")
	}

	if s.AuthBasic != nil {
		if s.AuthBasic.Switch != "" {
			conf.AddBody("auth_basic", s.AuthBasic.Switch)
		}
		if s.AuthBasic.UserFile != "" {
			conf.AddBody("auth_basic_user_file", s.AuthBasic.UserFile)
		}
	}
	if s.SSL != nil {
		conf.AddBody("ssl_certificate", s.SSL.Certificate)
		conf.AddBody("ssl_certificate_key", s.SSL.CertificateKey)
		conf.AddBody("ssl_protocols", s.SSL.Protocols)
		if s.SSL.HTTPRedirect {
			conf.AddBody("if", "( $scheme = 'http' ) ").AddBody("return", "301", "https://$host$request_uri")
		}
	}
	conf.AddBodyDirective(s.Parameters...)

	if s.Protocol != api.ProtocolHTTP {
		conf.AddBody("proxy_pass", s.ProxyPass)
	} else {
		for _, location := range s.Locations {
			lc := conf.AddBody("location", location.Path)
			if location.Commit != "" {
				lc.AddBody("#", location.Commit)
			}
			switch location.Type {
			case api.ProxyHTTP:
				lc.AddBody("proxy_pass", location.HTTP.To)
			case api.ProxyUpstream:
				path := fmt.Sprintf("http://%s", location.Upstream.Name)
				if location.Upstream.Path != "" {
					path += location.Upstream.Path
				}
				lc.AddBody("proxy_pass", path)
			case api.ProxyHTML:
				lc.AddBody(location.HTML.Model, location.HTML.Path)
				if location.HTML.Indexes != "" {
					lc.AddBody("index", location.HTML.Indexes)
				}
			}
			if location.BasicHeader {
				lc.AddBody("proxy_set_header", "X-Scheme", "$scheme")
				lc.AddBody("proxy_set_header", "Host", "$host")
				lc.AddBody("proxy_set_header", "X-Real-IP", "$remote_addr")
				lc.AddBody("proxy_set_header", "X-Forwarded-For", "$proxy_add_x_forwarded_for")
			}
			if location.WebSocket {
				lc.AddBody("proxy_set_header", "Upgrade", "$http_upgrade")
				lc.AddBody("proxy_set_header", "Connection", "'Upgrade'")
			}
			if location.AuthBasic != nil {
				if location.AuthBasic.Switch != "" {
					lc.AddBody("auth_basic", location.AuthBasic.Switch)
				}
				if location.AuthBasic.UserFile != "" {
					lc.AddBody("auth_basic_user_file", location.AuthBasic.UserFile)
				}
			}
			for _, allow := range location.Allows {
				lc.AddBody("allow", allow)
			}
			for _, deny := range location.Denys {
				lc.AddBody("deny", deny)
			}
			lc.AddBodyDirective(location.Parameters...)
		}
	}
	for _, allow := range s.Allows {
		conf.AddBody("allow", allow)
	}
	for _, deny := range s.Denys {
		conf.AddBody("deny", deny)
	}
	return conf
}

func (a *client) filterServer(filter *api.Filter, server *api.Server) bool {
	if filter == nil {
		return true
	}
	matched := filter.Name == ""
	if filter.Name != "" {
		//tcp/udp搜索
		if server.Protocol != api.ProtocolHTTP {
			if filter.ExactMatch {
				matched = filter.Name == server.ProxyPass
			} else {
				matched = strings.Contains(server.ProxyPass, filter.Name)
			}
		} else {
			for _, domain := range server.Domains {
				if filter.ExactMatch {
					if domain == filter.Name {
						matched = true
						break
					}
				} else if strings.Contains(domain, filter.Name) {
					matched = true
					break
				}
			}
		}
	}
	if filter.Commit != "" {
		matched = matched && strings.Contains(server.Commit, filter.Commit)
	}
	if filter.Protocol != "" {
		matched = matched && server.Protocol == filter.Protocol
	}
	return matched
}

func (a *client) GetServers(filter *api.Filter) (ss []*api.Server, err error) {
	var conf *config.Configuration
	if conf, err = a.Configuration(); err != nil {
		return
	}
	ss = make([]*api.Server, 0)

	var upstreams []*api.Upstream
	if upstreams, err = a.GetUpstream(nil); err != nil {
		return
	}

	protocols := map[string]api.Protocol{
		"http":   api.ProtocolHTTP,
		"stream": api.ProtocolTCP,
	}
	for name, protocol := range protocols {
		//http.server下查找
		if selects, err := query.Selects(conf, name, "server"); err == nil {
			for _, directive := range selects {
				server := analysisServer(directive, upstreams)
				if server.Protocol == "" {
					server.Protocol = protocol
				}
				if server.Protocol == api.ProtocolHTTP {
					server.Queries = []string{name, fmt.Sprintf("server.server_name('%s')", server.Domains[0])}
				} else {
					server.Queries = []string{name, fmt.Sprintf("server.proxy_pass('%s')", server.ProxyPass)}
				}
				if a.filterServer(filter, server) {
					ss = append(ss, server)
				}
			}
		}
		if includes, err := query.Selects(conf, name, "include"); err == nil {
			for _, include := range includes {
				match := include.Args[0]
				for _, file := range include.Body {
					fileName := file.Args[0]
					queries := []string{
						name,
						fmt.Sprintf("include('%s')", match),
						fmt.Sprintf("file('%s')", fileName),
						"server",
					}
					if selects, err := query.Selects(conf, queries...); err == nil {
						for _, directive := range selects {
							server := analysisServer(directive, upstreams)
							if server.Protocol == "" {
								server.Protocol = protocol
							}
							if server.Protocol == api.ProtocolHTTP {
								server.Queries = append(queries[:3], fmt.Sprintf("server.server_name('%s')", server.Domains[0]))
							} else {
								server.Queries = append(queries[:3], fmt.Sprintf("server.proxy_pass('%s')", server.ProxyPass))
							}
							if a.filterServer(filter, server) {
								ss = append(ss, server)
							}
						}
					}
				}
			}
		}
	}
	return
}

func (a *client) SetServer(server *api.Server) ([]string, error) {
	conf := Server2Config(server)
	//更新
	if server.Queries != nil && len(server.Queries) != 0 {
		return server.Queries, a.Directive().Modify(server.Queries, conf)
	}

	path := ""
	if server.Protocol == api.ProtocolHTTP {
		if _, err := a.Directive().Select("http", "include('hosts.d/*.conf')"); err != nil {
			include := config.New("include", "hosts.d/*.conf")
			if err = a.Directive().Add([]string{"http"}, include); err != nil {
				return nil, err
			}
		}
		path = fmt.Sprintf("hosts.d/%s.conf", server.Domains[0])
		server.Queries = []string{
			"http", "include('hosts.d/*.conf')",
			fmt.Sprintf("file('%s')", path),
			fmt.Sprintf("server.server_name('%s')", server.Domains[0]),
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
		path = fmt.Sprintf("stream.d/%s.conf", server.Domains[0])
		server.Queries = []string{
			"http", "include('stream.d/*.conf')",
			fmt.Sprintf("file('%s')", path),
			fmt.Sprintf("server.server_name('%s')", server.Domains[0]),
		}
	}
	err := a.Files().NewWithContent(path, []byte(conf.String()))
	return server.Queries, err
}
