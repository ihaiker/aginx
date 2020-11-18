package functions

import (
	"text/template"
)

func TemplateFuncs() template.FuncMap {
	return Merge(
		envTemplateFuncs(),
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
