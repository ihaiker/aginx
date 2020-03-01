package functions

import (
	"github.com/ihaiker/aginx/api"
	"text/template"
)

func TemplateFuncs(aginx api.Aginx) template.FuncMap {
	return Merge(
		envTemplateFuncs(),
		aginxTemplateFuncMap(aginx),
		stringsFuncMap(),
	)
}

func Merge(fns ...template.FuncMap) template.FuncMap {
	merged := template.FuncMap{}
	for _, fnMap := range fns {
		for name, fn := range fnMap {
			merged[name] = fn
		}
	}
	return merged
}
