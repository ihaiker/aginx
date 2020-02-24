package lego

import (
	"github.com/go-acme/lego/v3/certificate"
	"github.com/ihaiker/aginx/plugins"
	"time"
)

type Certificate struct {
	ExpireTime time.Time `json:"expire"`
	Email      string    `json:"email"`

	*certificate.Resource

	IssuerCertificate string `json:"issuerCertificate"`
	Certificate       string `json:"certificate"`
	PrivateKey        string `json:"privateKey"`

	PEM string `json:"pem"`
}

type StoreFile struct {
	Certificate       string `json:"certificate"`
	IssuerCertificate string `json:"issuerCertificate"`
	PEM               string `json:"pem"`
	PrivateKey        string `json:"privateKey"`
}

func (cfs *Certificate) GetStoreFile() *StoreFile {
	storePath := certificateDir + "/" + cfs.Domain
	return &StoreFile{
		Certificate:       storePath + "/server.crt",
		IssuerCertificate: storePath + "/server.issuer.crt",
		PEM:               storePath + "/server.key",
		PrivateKey:        storePath + "/server.pem",
	}
}

func (cfs *Certificate) StoreFile(engine plugins.StorageEngine) (file *StoreFile, err error) {
	file = cfs.GetStoreFile()
	if err = engine.Put(file.Certificate, []byte(cfs.Certificate)); err != nil {
		return
	}
	if err = engine.Put(file.IssuerCertificate, []byte(cfs.IssuerCertificate)); err != nil {
		return
	}
	if err = engine.Put(file.PrivateKey, []byte(cfs.PrivateKey)); err != nil {
		return
	}
	if err = engine.Put(file.PEM, []byte(cfs.PEM)); err != nil {
		return
	}
	return
}

func (cfs *Certificate) LoadCertificate() error {

	// .crt
	if cfs.Resource.Certificate != nil {
		cfs.Certificate = string(cfs.Resource.Certificate)
	}

	//.issuer.crt
	if cfs.Resource.IssuerCertificate != nil {
		cfs.IssuerCertificate = string(cfs.Resource.IssuerCertificate)
	}

	//.key
	if cfs.Resource.PrivateKey != nil {
		cfs.PrivateKey = string(cfs.Resource.PrivateKey)

		//.pem
		cfs.PEM = cfs.Certificate + cfs.PrivateKey
	}
	return nil
}
