package registry

import (
	"bytes"
	"github.com/ihaiker/aginx/api"
	"github.com/ihaiker/aginx/logs"
	"github.com/ihaiker/aginx/nginx"
	"github.com/ihaiker/aginx/plugins"
	"github.com/ihaiker/aginx/util"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"
)

var logger = logs.New("registry")

const default_template = `
upstream {{.Domain}} { {{range .Servers}}
	server {{.Address}} {{if ne .Weight 0}} weight={{.Weight}}{{end}};{{end}}
}
{{if .AutoSSL}}server {
	listen       80;
	server_name {{.Domain}};	
	return 301 https://$host$request_uri;
}{{end}}
server { {{if .AutoSSL}}
	listen 443 ssl;
	ssl_certificate     {{.SSL.Certificate}};        
	ssl_certificate_key {{.SSL.PrivateKey}};
	ssl_session_timeout 5m;
	ssl_ciphers ECDHE-RSA-AES128-GCM-SHA256:ECDHE:ECDH:AES:HIGH:!NULL:!aNULL:!MD5:!ADH:!RC4;
	ssl_protocols TLSv1 TLSv1.1 TLSv1.2;
	ssl_prefer_server_ciphers on; {{else}}
	listen 80; {{end}}

    server_name {{.Domain}};
    try_files $uri @tornado;

    location @tornado {
        proxy_set_header        X-Scheme        $scheme;
        proxy_set_header        Host            $host;
        proxy_set_header        X-Real-IP       $remote_addr;
        proxy_set_header        X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_pass http://{{.Domain}};
    }
}
`

var funcMap = template.FuncMap{}

type RegisterBridge struct {
	Register           plugins.Register
	LocalTemplateDir   string //本地模板
	StorageTemplateDir string //使用配置存储获取模板
	Aginx              api.Aginx
}

func (rb *RegisterBridge) listenChange() error {
	go func() {
		for {
			select {
			case event := <-rb.Register.Listener():
				for domain, _ := range event.Servers.Group() {
					servers := rb.Register.Get(domain)
					if len(servers) == 0 {
						relPath := "reg.d/" + domain + ".ngx.conf"
						err := rb.Aginx.File().Remove(relPath)
						logger.WithError(err).Info("Publishing service changes ", domain, " remove ", relPath)
					} else {
						err := rb.publishServer(domain, servers)
						logger.WithError(err).Info("Publishing service changes ", domain)
					}
				}
				_ = rb.Aginx.Reload()
			}
		}
	}()
	return nil
}

func (rb *RegisterBridge) createRegDInclude() (err error) {
	if _, err = rb.Aginx.Directive().Select("http", "include('reg.d/*.ngx.conf')"); err != nil {
		logger.Debug("create NGINX directive (include reg.d/*.ngx.conf)")
		err = rb.Aginx.Directive().Add(api.Queries("http"),
			nginx.NewDirective("include", "reg.d/*.ngx.conf"))
		if err != nil {
			logger.Debug("create NGINX directive  (include reg.d/*.ngx.conf) error ", err)
		}
	}
	return
}

func (rb *RegisterBridge) findTemplate(domain string) string {
	//本地查找模板
	if rb.LocalTemplateDir != "" {
		localTemplate := filepath.Join(rb.LocalTemplateDir, domain+".ngx.tpl")
		if content, err := ioutil.ReadFile(localTemplate); err == nil {
			return string(content)
		} else if !os.IsNotExist(err) {
			logger.Warn("read template file ", localTemplate, " error ", err)
		}
		localTemplate = filepath.Join(rb.LocalTemplateDir, "default.ngx.tpl")
		if content, err := ioutil.ReadFile(localTemplate); err == nil {
			return string(content)
		} else if !os.IsNotExist(err) {
			logger.Warn("read template file ", localTemplate, " error ", err)
		}
	}

	if rb.StorageTemplateDir != "" {
		domainTemplatePath := filepath.Join(rb.StorageTemplateDir, domain+".ngx.tpl")      //针对domain服务的模板
		userDefinedTemplatePath := filepath.Join(rb.StorageTemplateDir, "default.ngx.tpl") //用户定义的全局模板
		if templateFiles, err := rb.Aginx.File().Search(domainTemplatePath, userDefinedTemplatePath); err != nil {
			logger.WithField("dir", rb.StorageTemplateDir).Warn("read store template file error: ", err)
			return default_template
		} else if r1, has := templateFiles[domainTemplatePath]; has {
			return r1
		} else if r2, has := templateFiles[userDefinedTemplatePath]; has {
			return r2
		}
	}

	return default_template
}

func (rb *RegisterBridge) publishServer(domain string, servers plugins.Domains) error {
	autoSsl := servers[0].AutoSSL
	data := map[string]interface{}{
		"Domain":  domain,
		"AutoSSL": autoSsl,
		"Servers": servers,
	}
	if autoSsl {
		if certFile, err := rb.Aginx.SSL().New("", domain); err != nil {
			return err
		} else {
			data["SSL"] = certFile
		}
	}
	templateFile := rb.findTemplate(domain)
	out := bytes.NewBufferString("")
	if t, err := template.New("").Parse(templateFile); err != nil {
		return err
	} else if err := t.Funcs(funcMap).Execute(out, data); err != nil {
		return err
	} else {
		relPath := "reg.d/" + domain + ".ngx.conf"
		return rb.Aginx.File().NewWithContent(relPath, out.Bytes())
	}
}

func (rb *RegisterBridge) Start() error {
	logger.Info("start")

	if err := rb.createRegDInclude(); err != nil {
		return err
	}

	if err := util.StartService(rb.Register); err != nil {
		return err
	}

	services := rb.Register.Sync().Group()
	for domain, servers := range services {
		if err := rb.publishServer(domain, servers); err != nil {
			return err
		}
	}

	_ = rb.Aginx.Reload()
	return rb.listenChange()
}

func (rb *RegisterBridge) Stop() error {
	if rb.Register != nil {
		return util.StopService(rb.Register)
	}
	return nil
}
