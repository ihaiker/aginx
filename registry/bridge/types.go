package bridge

import "github.com/ihaiker/aginx/api"

type TemplateDate struct {
	Aginx api.Aginx
	Data  interface{}
}

func Data(aginx api.Aginx, data interface{}) *TemplateDate {
	return &TemplateDate{Aginx: aginx, Data: data}
}
