package http

import (
	"github.com/ihaiker/aginx/lego"
	"github.com/ihaiker/aginx/nginx"
	"github.com/ihaiker/aginx/util"
	"github.com/kataras/iris/v12"
)

type sslController struct {
	email string
}

func (self *sslController) New(ctx iris.Context, api *nginx.Client, domain string) *lego.StoreFile {
	email := ctx.URLParamDefault("email", self.email)
	return api.NewCertificate(email, domain)
}

func (self *sslController) Renew(api *nginx.Client, domain string) *lego.StoreFile {
	cert, has := api.Lego.CertificateStorage.Get(domain)
	if !has {
		util.PanicIfError(util.ErrNotFound)
	}
	return api.NewCertificate(cert.Email, cert.Domain)
}
