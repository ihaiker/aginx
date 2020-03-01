package bridge

import (
	"bytes"
	"fmt"
	"github.com/ihaiker/aginx/api"
	"github.com/ihaiker/aginx/nginx"
	"github.com/ihaiker/aginx/plugins"
	"github.com/ihaiker/aginx/registry/functions"
	"github.com/ihaiker/aginx/util"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"
)

type TemplateRegisterBridge struct {
	plugins.Register
	Aginx                 api.Aginx
	Name                  string
	Template              string
	AppendTemplateFuncMap template.FuncMap
}

func (self *TemplateRegisterBridge) createRegisterInclude() (err error) {
	if _, err = self.Aginx.Directive().Select("http", "include('register.d/*.ngx.conf')"); err != nil {
		logger.Debug("create NGINX directive (include register.d/*.ngx.conf)")
		err = self.Aginx.Directive().Add(api.Queries("http"),
			nginx.NewDirective("include", "register.d/*.ngx.conf"))
		if err != nil {
			logger.Debug("create NGINX directive  (include register.d/*.ngx.conf error: ", err.Error())
		} else {
			_ = self.Aginx.Reload()
		}
	}
	return
}

func (self *TemplateRegisterBridge) findTemplate() string {
	localDomainTemplate := filepath.Join(nginx.MustConfigDir(), self.Template)
	if content, err := ioutil.ReadFile(localDomainTemplate); err == nil {
		return string(content)
	} else if !os.IsNotExist(err) {
		logger.Warn("read template file ", localDomainTemplate, " error ", err)
	}

	if templateFiles, err := self.Aginx.File().Search(self.Template); err != nil {
		logger.Warnf("read store template(%s) file error: %s", self.Template, err)
	} else if content, has := templateFiles[self.Template]; has {
		return content
	}
	return ""
}

func (self *TemplateRegisterBridge) publishEvent(data interface{}) error {
	templateContent := self.findTemplate()
	if templateContent == "" {
		return fmt.Errorf("Failed to find template file: %s ", self.Template)
	}
	funcs := functions.Merge(self.AppendTemplateFuncMap, self.TemplateFuncMap())
	out := bytes.NewBufferString(fmt.Sprintf("# generate by %s template register\n", self.Name))
	if t, err := template.New("").Funcs(funcs).Parse(templateContent); err != nil {
		return err
	} else if err := t.Execute(out, Data(self.Aginx, data)); err != nil {
		return err
	} else {
		relPath := fmt.Sprintf("register.d/%s.ngx.conf", self.Name)
		return self.Aginx.File().NewWithContent(relPath, util.CleanEmptyLine(out.Bytes()))
	}
}

func (self *TemplateRegisterBridge) listenChange() {
	for {
		select {
		case event, has := <-self.Register.Listener():
			if has {
				if err := self.publishEvent(event); err != nil {
					logger.Warn("publish error: ", err)
				} else {
					_ = self.Aginx.Reload()
				}
			}
		}
	}
}

func (self *TemplateRegisterBridge) Start() error {
	if err := self.createRegisterInclude(); err != nil {
		return err
	}
	go self.listenChange()
	return self.Register.Start()
}
