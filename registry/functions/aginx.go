package functions

import (
	"github.com/ihaiker/aginx/api"
	"github.com/ihaiker/aginx/lego"
	"text/template"
)

func sslTemplateFuncMap(aginx api.Aginx) interface{} {
	return func(accountEmail, domain string) (*lego.StoreFile, error) {
		return aginx.SSL().New(accountEmail, domain)
	}
}

func aginxTemplateFuncMap(aginx api.Aginx) template.FuncMap {
	return template.FuncMap{
		"autossl": sslTemplateFuncMap(aginx),
	}
}
