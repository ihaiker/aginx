package bridge

import (
	"bytes"
	"fmt"
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

var logger = logs.New("registry", "module", "bridge")

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

type LabelRegisterBridge struct {
	Aginx api.Aginx
	plugins.Register
	Name          string
	TemplateDir   string
	TemplateFuncs template.FuncMap
}

func (rb *LabelRegisterBridge) listenChange() error {
	go func() {
		for {
			select {
			case event, has := <-rb.Register.Listener():
				if has {
					domains := event.(plugins.LabelsRegistryEvent)
					for domain, servers := range domains {
						if len(servers) == 0 {
							relPath := fmt.Sprintf("%s.d/%s.ngx.conf", rb.Name, domain)
							logger.Info("Publishing service changes ", domain, " remove ", relPath)
							if err := rb.Aginx.File().Remove(relPath); err != nil {
								logger.WithError(err).Warn("Publishing service changes ", domain, " remove ", relPath)
							}
						} else {
							logger.Info("Publishing service changes ", domain)
							if err := rb.publishServer(domain, servers); err != nil {
								logger.Warn("Publishing service changes ", domain, " error ", err)
							}
						}
					}
					_ = rb.Aginx.Reload()
				}
			}
		}
	}()
	return nil
}

func (rb *LabelRegisterBridge) createInclude() (err error) {
	if _, err = rb.Aginx.Directive().Select("http", fmt.Sprintf("include('%s.d/*.ngx.conf')", rb.Name)); err != nil {
		logger.Debugf("create NGINX directive (include %s.d/*.ngx.conf)", rb.Name)
		err = rb.Aginx.Directive().Add(api.Queries("http"),
			nginx.NewDirective("include", fmt.Sprintf("%s.d/*.ngx.conf", rb.Name)))
		if err != nil {
			logger.Debugf("create NGINX directive  (include %s.d/*.ngx.conf) error: %s ", rb.Name, err.Error())
		} else {
			_ = rb.Aginx.Reload()
		}
	}
	return
}

func (rb *LabelRegisterBridge) findTemplate(domain string) string {
	{
		configDir := nginx.MustConfigDir()
		localDomainTemplate := filepath.Join(configDir, rb.TemplateDir, domain+".ngx.tpl")
		if content, err := ioutil.ReadFile(localDomainTemplate); err == nil {
			return string(content)
		} else if !os.IsNotExist(err) {
			logger.Warn("read template file ", localDomainTemplate, " error ", err)
		}

		localDomainTemplate = filepath.Join(configDir, rb.TemplateDir, "default.ngx.tpl")
		if content, err := ioutil.ReadFile(localDomainTemplate); err == nil {
			return string(content)
		} else if !os.IsNotExist(err) {
			logger.Warn("read template file ", localDomainTemplate, " error ", err)
		}
	}

	{
		domainTemplatePath := filepath.Join(rb.TemplateDir, domain+".ngx.tpl")      //针对domain服务的模板
		userDefinedTemplatePath := filepath.Join(rb.TemplateDir, "default.ngx.tpl") //用户定义的全局模板
		if templateFiles, err := rb.Aginx.File().Search(domainTemplatePath, userDefinedTemplatePath); err != nil {
			logger.Warnf("read store template(%s) file error: %s", rb.TemplateDir, err)
			return default_template
		} else if r1, has := templateFiles[domainTemplatePath]; has {
			return r1
		} else if r2, has := templateFiles[userDefinedTemplatePath]; has {
			return r2
		}
	}

	return default_template
}

func (rb *LabelRegisterBridge) publishServer(domain string, servers plugins.Domains) error {
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
	} else if err := t.Funcs(rb.TemplateFuncs).Execute(out, data); err != nil {
		return err
	} else {
		relPath := fmt.Sprintf("%s.d/%s.ngx.conf", rb.Name, domain)
		return rb.Aginx.File().NewWithContent(relPath, util.CleanEmptyLine(out.Bytes()))
	}
}

func (rb *LabelRegisterBridge) Start() error {
	if err := rb.createInclude(); err != nil {
		return err
	}

	if err := rb.Register.Start(); err != nil {
		return err
	}

	return rb.listenChange()
}

func (rb *LabelRegisterBridge) Stop() error {
	if rb.Register != nil {
		return rb.Register.Stop()
	}
	return nil
}
