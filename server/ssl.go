package server

import (
	"github.com/go-acme/lego/v3/certcrypto"
	"github.com/ihaiker/aginx/lego"
	"github.com/ihaiker/aginx/nginx/client"
	"github.com/ihaiker/aginx/storage"
	"github.com/ihaiker/aginx/util"
	"github.com/kataras/iris/v12"
	"sync"
)

type sslController struct {
	vister  *Supervister
	engine  storage.Engine
	manager *lego.Manager
	lock    sync.Locker
}

func (self *sslController) newCertificate(api *client.Client, email, domain string) *lego.StoreFile {

	if cert, has := self.manager.CertificateStorage.Get(domain); has {
		return cert.GetStoreFile()
	}

	var err error
	account, has := self.manager.AccountStorage.Get(email)
	if !has {
		account, err = self.manager.AccountStorage.New(email, certcrypto.EC384)
		util.PanicIfError(err)
	}

	provider := NewAginxProvider(api, self.vister)
	cert, err := self.manager.CertificateStorage.NewWithProvider(account, domain, provider)
	util.PanicIfError(err)

	file, err := cert.StoreFile(self.engine)
	util.PanicIfError(err)

	return file
}

func (self *sslController) New(ctx iris.Context, api *client.Client, domain string) *lego.StoreFile {
	self.lock.Lock()
	defer self.lock.Unlock()

	email := ctx.URLParamDefault("email", "aginx@examples.com")
	return self.newCertificate(api, email, domain)
}

func (self *sslController) Expire(domain string) {
	self.lock.Lock()
	defer self.lock.Unlock()

	api, _ := client.NewClient(self.engine)
	_ = self.Renew(api, domain)
}

func (self *sslController) Renew(api *client.Client, domain string) *lego.StoreFile {
	self.lock.Lock()
	defer self.lock.Unlock()

	cert, has := self.manager.CertificateStorage.Get(domain)
	if !has {
		util.PanicIfError(client.ErrNotFound)
	}
	return self.newCertificate(api, cert.Email, cert.Domain)
}
