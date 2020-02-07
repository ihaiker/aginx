package lego

import (
	"github.com/go-acme/lego/v3/certificate"
	"io/ioutil"
	"os"
	"time"
)

type Certificate struct {
	ExpireTime time.Time `json:"expire"`

	*certificate.Resource

	IssuerCertificate string `json:"issuerCertificate"`
	Certificate       string `json:"certificate"`
	PrivateKey        string `json:"privateKey"`

	PEM string `json:"pem"`
}

func (cfs *Certificate) StoreFile(path string) (err error) {
	if err = os.MkdirAll(path, os.ModePerm); err != nil {
		return
	}
	if err = ioutil.WriteFile(path+"/server.crt", []byte(cfs.Certificate), 0666); err != nil {
		return
	}
	if err = ioutil.WriteFile(path+"/server.issuer.crt", []byte(cfs.IssuerCertificate), 0666); err != nil {
		return
	}
	if err = ioutil.WriteFile(path+"/server.key", []byte(cfs.PrivateKey), 0666); err != nil {
		return
	}
	if err = ioutil.WriteFile(path+"/server.pem", []byte(cfs.PEM), 0666); err != nil {
		return
	}
	return nil
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
