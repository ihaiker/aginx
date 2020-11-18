package api

import (
	"github.com/ihaiker/aginx/v2/core/nginx/config"
	"github.com/ihaiker/aginx/v2/plugins/storage"
	"time"
)

const (
	BatchAdd    DirectiveBatchType = "add"
	BatchDelete DirectiveBatchType = "delete"
	BatchModify DirectiveBatchType = "motidy"
)

type (
	Aginx interface {
		//获取全局配置
		Configuration() (*config.Configuration, error)

		Files() File

		Directive() Directive

		Certs() Certs

		GetServers(filter *Filter) ([]*Server, error)
		SetServer(server *Server) (queries []string, err error)

		GetUpstream(filter *Filter) ([]*Upstream, error)
		SetUpstream(upstream *Upstream) (queries []string, err error)

		Info() (map[string]map[string]string, error) /*获取插件信息*/
	}

	//配置文件操作
	File interface {
		//把本地文件上传到指定位置
		New(relativePath, localFilePath string) error
		//上传文本内容
		NewWithContent(relativePath string, content []byte) error

		Remove(relativePath string) error

		//搜索文件，relativePaths可以为 *.conf , conf.d/*等
		Search(relativePaths ...string) ([]*storage.File, error)

		//获取具体文件
		Get(relativePath string) (*storage.File, error)
	}

	CertFile struct {
		Provider    string    `json:"provider"`    //提供商
		Domain      string    `json:"domain"`      //证书域名
		ExpireTime  time.Time `json:"expireTime"`  //过期时间
		Certificate string    `json:"certificate"` //证书公钥
		PrivateKey  string    `json:"privateKey"`  //证书私钥
	}

	Certs interface {
		//申请一个SSL证书，可以调用多次，多次调用每次都会申请
		New(provider, domain string) (*CertFile, error)
		//获取SSL证书
		Get(domain string) (*CertFile, error)

		List() ([]*CertFile, error)
	}

	DirectiveBatchType string
	DirectiveBatch     struct {
		Type       DirectiveBatchType
		Queries    []string
		Directives []*config.Directive
	}
	DirectiveBatches []*DirectiveBatch

	Directive interface {
		//查询配置
		Select(queries ...string) ([]*config.Directive, error)

		//添加配置
		Add(queries []string, addDirectives ...*config.Directive) error

		//删除
		Delete(queries ...string) error

		//更新配置
		Modify(queries []string, directive *config.Directive) error

		//批量操作
		Batch(batch []*DirectiveBatch) error
	}
)

func (bs *DirectiveBatches) Add(queries []string, conf ...*config.Directive) {
	*bs = append(*bs, &DirectiveBatch{
		Type:       BatchAdd,
		Queries:    queries,
		Directives: conf,
	})
}

func (bs *DirectiveBatches) Delete(queries ...string) {
	*bs = append(*bs, &DirectiveBatch{
		Type:    BatchDelete,
		Queries: queries,
	})
}

func (bs *DirectiveBatches) Modify(queries []string, conf *config.Directive) {
	*bs = append(*bs, &DirectiveBatch{
		Type:       BatchModify,
		Queries:    queries,
		Directives: []*config.Directive{conf},
	})
}
