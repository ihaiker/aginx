package http

import (
	"github.com/ihaiker/aginx/v2/api"
	"github.com/ihaiker/aginx/v2/core/util/errors"
	"github.com/kataras/iris/v12"
)

type sslController struct {
	aginx api.Aginx
}

func (s *sslController) New(ctx iris.Context) *api.CertFile {
	domain := ctx.URLParam("domain")
	provider := ctx.URLParam("provider")
	cert, err := s.aginx.Certs().New(provider, domain)
	errors.Panic(err)
	return cert
}

func (s *sslController) Get(ctx iris.Context) *api.CertFile {
	domain := ctx.URLParam("domain")
	cert, err := s.aginx.Certs().Get(domain)
	errors.Panic(err)
	return cert
}

func (s *sslController) List(ctx iris.Context) []*api.CertFile {
	certFiles, err := s.aginx.Certs().List()
	errors.Panic(err)
	return certFiles
}
