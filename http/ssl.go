package http

import (
	"github.com/go-acme/lego/v3/certcrypto"
	"github.com/ihaiker/aginx/lego"
	"github.com/ihaiker/aginx/nginx"
	"github.com/ihaiker/aginx/plugins"
	"github.com/ihaiker/aginx/util"
	"github.com/kataras/iris/v12"
	"sync"
)

type sslController struct {
	email   string
	process *nginx.Process
	engine  plugins.StorageEngine
	manager *lego.Manager
	lock    sync.Locker
}

func (self *sslController) newCertificate(api *nginx.Client, email, domain string) *lego.StoreFile {

	if cert, has := self.manager.CertificateStorage.Get(domain); has {
		return cert.GetStoreFile()
	}

	var err error
	account, has := self.manager.AccountStorage.Get(email)
	if !has {
		account, err = self.manager.AccountStorage.New(email, certcrypto.EC384)
		util.PanicIfError(err)
	}

	provider := NewAginxProvider(api, self.process)
	cert, err := self.manager.CertificateStorage.NewWithProvider(account, domain, provider)
	util.PanicIfError(err)

	return cert.GetStoreFile()
}

func (self *sslController) New(ctx iris.Context, api *nginx.Client, domain string) *lego.StoreFile {
	self.lock.Lock()
	defer self.lock.Unlock()

	email := ctx.URLParamDefault("email", self.email)
	return self.newCertificate(api, email, domain)
}

func (self *sslController) Expire(domain string) {
	self.lock.Lock()
	defer self.lock.Unlock()

	api, _ := nginx.NewClient(self.engine)
	_ = self.Renew(api, domain)
}

func (self *sslController) Renew(api *nginx.Client, domain string) *lego.StoreFile {
	self.lock.Lock()
	defer self.lock.Unlock()

	cert, has := self.manager.CertificateStorage.Get(domain)
	if !has {
		util.PanicIfError(nginx.ErrNotFound)
	}
	return self.newCertificate(api, cert.Email, cert.Domain)
}
