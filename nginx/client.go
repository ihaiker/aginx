package nginx

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/ihaiker/aginx/plugins"
	"github.com/ihaiker/aginx/util"
	"os"
	"strings"
)

var (
	ErrNotFound            = os.ErrNotExist
	ErrRootCannotBeDeleted = errors.New("root cannot be deleted")
)

func Queries(query ...string) []string {
	return query
}

type Client struct {
	doc    *Configuration
	Engine plugins.StorageEngine
}

func NewClient(engine plugins.StorageEngine) (*Client, error) {
	doc, err := Readable(engine)
	if err != nil {
		return nil, err
	}
	return &Client{doc: doc, Engine: engine}, nil
}

func MustClient(engine plugins.StorageEngine) *Client {
	client, err := NewClient(engine)
	util.PanicIfError(err)
	return client
}

func (client Client) Configuration() *Configuration {
	return client.doc
}

func (client Client) Store() error {
	return Write(client.doc,
		func(file string, content []byte) bool {
			if cfgFile, err := client.Engine.Get(file); err == nil {
				return !bytes.Equal(cfgFile.Content, content)
			}
			return true //匹配not_found
		},
		func(file string, content []byte) error {
			return client.Engine.Put(file, content)
		},
	)
}

func (client *Client) Select(queries ...string) ([]*Directive, error) {
	if len(queries) == 0 {
		return client.doc.Body, nil
	} else {
		return client.doc.Select(queries...)
	}
}

func (client *Client) Add(queries []string, addDirectives ...*Directive) error {
	if directives, err := client.Select(queries...); err == ErrNotFound {
		return err
	} else {
		for _, directive := range directives {
			directive.Body = append(directive.Body, addDirectives...)
		}
		return nil
	}
}

func (client *Client) Delete(queries ...string) error {
	if len(queries) == 0 {
		return ErrRootCannotBeDeleted
	}
	finder := queries[0 : len(queries)-1]
	directives, err := client.Select(finder...)
	if err != nil {
		return err
	}

	deleteQuery := queries[len(queries)-1]
	expr, err := Parser(deleteQuery)
	if err != nil {
		return err
	}

	err = ErrNotFound
	for _, directive := range directives {

		deleteDirectiveIdx := make([]int, 0)
		for i, body := range directive.Body {
			if expr.Match(body) {
				deleteDirectiveIdx = append(deleteDirectiveIdx, i)
			}
		}
		if len(deleteDirectiveIdx) > 0 {
			err = nil
		}

		for i := len(deleteDirectiveIdx) - 1; i >= 0; i-- {
			idx := deleteDirectiveIdx[i]
			directive.Body = append(directive.Body[:idx], directive.Body[idx+1:]...)
		}
	}
	return err
}

func (client *Client) Modify(queries []string, directive *Directive) error {
	selectDirectives, err := client.Select(queries...)
	if err != nil {
		return err
	}
	for _, selectDirective := range selectDirectives {
		selectDirective.Name = directive.Name
		selectDirective.Args = directive.Args
		selectDirective.Body = directive.Body
	}
	return nil
}

func (client *Client) hostsd(include string) {
	directives, err := client.Select("http", "include")
	exists := os.IsNotExist(err)
	for _, directive := range directives {
		if include == directive.Args[0] {
			exists = true
		}
	}
	if !exists {
		util.PanicIfError(client.Add(Queries("http"), NewDirective("include", include)))
	}
}

//设置domain对应的负载
func (client *Client) SimpleServer(domain string, address ...string) (err error) {
	defer util.Catch(func(e error) {
		logger.WithError(err).Debug("new simple server ", domain, strings.Join(address, ","))
	})

	client.hostsd("hosts.d/*.conf")

	upstreamName := UpstreamName(domain)

	serverQueries, selectServer := client.selectServer("http", domain)
	if selectServer != nil {
		if proxyPassDirectives, err := selectServer.Select("location", "proxy_pass"); err == nil {
			proxyPassAddress := proxyPassDirectives[0].Args[0]
			if !strings.Contains(proxyPassAddress[7:], ":") {
				upstreamName = proxyPassAddress[len("http://"):]
			}
		}
	}
	upstreamQueries, selectUpstream := client.selectUpStream("http", upstreamName)

	if selectServer != nil {
		_ = client.Delete(serverQueries...)
	}
	if selectUpstream != nil {
		_ = client.Delete(upstreamQueries...)
	}

	upstream, server := SimpleServer(domain, address...)
	files, err := client.Select("http", "include('hosts.d/*.conf')", fmt.Sprintf("file('hosts.d/%s.ngx.conf')", domain))

	if os.IsNotExist(err) {
		file := NewDirective("file", fmt.Sprintf("hosts.d/%s.ngx.conf", domain))
		file.Virtual = Include
		file.AddBodyDirective(upstream, server)
		err = client.Add(Queries("http", "include('hosts.d/*.conf')"), file)
	} else {
		files[0].Body = append(files[0].Body, upstream, server)
	}
	return
}

func (client *Client) selectServer(first, domain string) (queries []string, directive *Directive) {
	serverQuery := fmt.Sprintf("server.[server_name('%s') & listen('80')]", domain)
	queries = Queries(first, "include", "*", serverQuery)
	if directives, err := client.Select(queries...); err == nil {
		directive = directives[0]
		return
	}
	queries = Queries(first, serverQuery)
	if directives, err := client.Select(queries...); err == nil {
		directive = directives[0]
		return
	}
	return
}

func (client *Client) selectUpStream(first, name string) (queries []string, directive *Directive) {
	serverQuery := fmt.Sprintf("upstream('%s')", name)
	queries = Queries(first, "include", "*", serverQuery)
	if directives, err := client.Select(queries...); err == nil {
		directive = directives[0]
		return
	}
	queries = Queries(first, serverQuery)
	if directives, err := client.Select(queries...); err == nil {
		directive = directives[0]
		return
	}
	return
}
