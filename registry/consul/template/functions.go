package consulTemplate

import (
	consulApi "github.com/hashicorp/consul/api"
	"text/template"
)

func templateFuncs(consul *consulApi.Client) template.FuncMap {
	return template.FuncMap{}
}
