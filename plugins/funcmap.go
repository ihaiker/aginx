package plugins

import "text/template"

type LoadFuncMap func(register Register) template.FuncMap

const (
	PLUGIN_FUNCMAP = "LoadFuncMap"
)
