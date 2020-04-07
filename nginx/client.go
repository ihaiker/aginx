package nginx

import (
	"bytes"
	"fmt"
	"github.com/go-acme/lego/v3/certcrypto"
	"github.com/ihaiker/aginx/lego"
	"github.com/ihaiker/aginx/nginx/config"
	"github.com/ihaiker/aginx/nginx/query"
	"github.com/ihaiker/aginx/plugins"
	"github.com/ihaiker/aginx/util"
	"os"
	"strings"
)

func Queries(query ...string) []string {
	return query
}

type Client struct {
	doc     *config.Configuration
	Email   string
	Engine  plugins.StorageEngine
	Lego    *lego.Manager
	Process *Process
}

func NewClient(email string, engine plugins.StorageEngine, lego *lego.Manager, process *Process) (*Client, error) {
	doc, err := Readable(engine)
	if err != nil {
		return nil, err
	}
	return &Client{
		Email: email,
		doc:   doc, Engine: engine, Lego: lego,
		Process: process,
	}, nil
}

func MustClient(email string, engine plugins.StorageEngine, lego *lego.Manager, process *Process) *Client {
	client, err := NewClient(email, engine, lego, process)
	util.PanicMessage(err, "parse config error")
	return client
}

func (client Client) Configuration() *config.Configuration {
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

func (client *Client) Select(queries ...string) ([]*config.Directive, error) {
	if len(queries) == 0 {
		return client.doc.Body, nil
	} else {
		return Select(client.doc, queries...)
	}
}

func (client *Client) MustSelect(queries ...string) []*config.Directive {
	directives, err := client.Select(queries...)
	util.PanicIfError(err)
	return directives
}

func (client *Client) Add(queries []string, addDirectives ...*config.Directive) error {
	if directives, err := client.Select(queries...); err == util.ErrNotFound {
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
		return util.ErrRootCannotBeDeleted
	}
	finder := queries[0 : len(queries)-1]
	directives, err := client.Select(finder...)
	if err != nil {
		return err
	}

	deleteQuery := queries[len(queries)-1]
	expr, err := query.Lexer(deleteQuery)
	if err != nil {
		return err
	}

	err = util.ErrNotFound
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

func (client *Client) Modify(queries []string, directive *config.Directive) error {
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
		util.PanicIfError(client.Add(Queries("http"), config.NewDirective("include", include)))
	}
}

//设置domain对应的负载
func (client *Client) SimpleServer(domain string, ssl bool, address ...string) (err error) {
	defer util.Catch(func(e error) {
		logger.WithError(err).Debug("new simple server ", domain, " ", strings.Join(address, ","))
	})
	client.hostsd("hosts.d/*.conf")

	upstreamName := UpstreamName(domain)

	serverQueries, selectServers := client.selectServer("http", domain)
	for _, selectServer := range selectServers {
		if proxyPassDirectives, err := Select(selectServer, "location", "proxy_pass"); err == nil {
			proxyPassAddress := proxyPassDirectives[0].Args[0]
			if !strings.Contains(proxyPassAddress[7:], ":") {
				upstreamName = proxyPassAddress[len("http://"):]
			}
		}
	}
	if selectServers != nil {
		_ = client.Delete(serverQueries...)
	}

	upstreamQueries, selectUpstream := client.selectUpStream("http", upstreamName)
	if selectUpstream != nil {
		_ = client.Delete(upstreamQueries...)
	}

	upstream, server := SimpleServer(domain, address...)
	if ssl {
		sslFile := client.NewCertificate(client.Email, domain)
		listen := MustSelect(server, "listen")[0]
		listen.Args = []string{"443", "ssl"}

		server.AddBody("ssl_certificate", sslFile.Certificate)
		server.AddBody("ssl_certificate_key", sslFile.PrivateKey)
		server.AddBody("ssl_session_timeout", "5m")
		server.AddBody("ssl_ciphers", "ECDHE-RSA-AES128-GCM-SHA256:ECDHE:ECDH:AES:HIGH:!NULL:!aNULL:!MD5:!ADH:!RC4")
		server.AddBody("ssl_protocols", "TLSv1", "TLSv1.1", "TLSv1.2")
		server.AddBody("ssl_prefer_server_ciphers", "on")
	}
	rewrite := config.NewDirective("server")
	{
		rewrite.AddBody("listen", "80")
		rewrite.AddBody("server_name", domain)
		rewrite.AddBody("return", "301", "https://$host$request_uri")
	}

	files, err := client.Select("http", "include('hosts.d/*.conf')", fmt.Sprintf("file('hosts.d/%s.ngx.conf')", domain))
	if os.IsNotExist(err) {
		file := config.NewDirective("file", fmt.Sprintf("hosts.d/%s.ngx.conf", domain))
		file.Virtual = config.Include
		if ssl {
			file.AddBodyDirective(upstream, rewrite, server)
		} else {
			file.AddBodyDirective(upstream, server)
		}
		err = client.Add(Queries("http", "include('hosts.d/*.conf')"), file)
	} else {
		if ssl {
			files[0].Body = append(files[0].Body, upstream, rewrite, server)
		} else {
			files[0].Body = append(files[0].Body, upstream, server)
		}
	}
	return
}

func (client *Client) selectServer(first, domain string) (queries []string, selectDirectives []*config.Directive) {
	serverQuery := fmt.Sprintf("server.server_name('%s')", domain)
	queries = Queries(first, "include", "*", serverQuery)
	if directives, err := client.Select(queries...); err == nil {
		selectDirectives = directives
		return
	}
	queries = Queries(first, serverQuery)
	if directives, err := client.Select(queries...); err == nil {
		selectDirectives = directives
		return
	}
	return
}

func (client *Client) selectUpStream(first, name string) (queries []string, directive *config.Directive) {
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

func (self *Client) NewCertificate(email, domain string) *lego.StoreFile {
	if email == "" {
		email = self.Email
	}
	if cert, has := self.Lego.CertificateStorage.Get(domain); has {
		return cert.GetStoreFile()
	}

	var err error
	account, has := self.Lego.AccountStorage.Get(email)
	if !has {
		account, err = self.Lego.AccountStorage.New(email, certcrypto.EC384)
		util.PanicIfError(err)
	}

	provider := NewAginxProvider(self, self.Process)
	cert, err := self.Lego.CertificateStorage.NewWithProvider(account, domain, provider)
	util.PanicIfError(err)

	return cert.GetStoreFile()
}
