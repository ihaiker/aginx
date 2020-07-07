package api

import (
	"github.com/ihaiker/aginx/lego"
	ngx "github.com/ihaiker/aginx/nginx/config"
)

func Queries(query ...string) []string {
	return query
}

type ApiError struct {
	Code    string `json:"error"`
	Message string `json:"message"`
}

func (err *ApiError) Error() string {
	return err.Message
}

type AginxFile interface {
	New(relativePath, localFileAbsPath string) error

	NewWithContent(relativePath string, content []byte) error

	Remove(relativePath string) error

	Search(relativePaths ...string) (map[string]string, error)

	Get(relativePath string) (string, error)
}

type AginxSSL interface {
	New(accountEmail, domain string) (*lego.StoreFile, error)
	ReNew(domain string) (*lego.StoreFile, error)
}

type AginxDirective interface {
	//查询配置
	Select(queries ...string) ([]*ngx.Directive, error)

	//添加配置
	Add(queries []string, addDirectives ...*ngx.Directive) error

	//删除
	Delete(queries ...string) error

	//更新配置
	Modify(queries []string, directive *ngx.Directive) error
}

type Aginx interface {
	Auth(name, password string)

	//获取全局配置
	Configuration() (*ngx.Configuration, error)

	//nginx -s reload
	Reload() error

	File() AginxFile

	Directive() AginxDirective

	SSL() AginxSSL
}
