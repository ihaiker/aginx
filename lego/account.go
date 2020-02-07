package lego

import (
	"bytes"
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"github.com/go-acme/lego/v3/certcrypto"
	"github.com/go-acme/lego/v3/registration"
)

type Account struct {
	KeyType      certcrypto.KeyType
	Email        string
	Registration *registration.Resource
	Key          string
	privateKey   crypto.PrivateKey
}

func (u *Account) SetKey(privateKey crypto.PrivateKey) (err error) {
	u.privateKey = privateKey

	pemKey := certcrypto.PEMBlock(privateKey)
	out := bytes.NewBuffer([]byte{})
	if err = pem.Encode(out, pemKey); err != nil {
		return err
	}

	u.Key = out.String()
	return
}

func (u *Account) GetKey() (crypto.PrivateKey, error) {
	if u.privateKey != nil {
		return u.privateKey, nil
	}
	keyBlock, _ := pem.Decode([]byte(u.Key))
	switch keyBlock.Type {
	case "RSA PRIVATE KEY":
		return x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	case "EC PRIVATE KEY":
		return x509.ParseECPrivateKey(keyBlock.Bytes)
	}
	return nil, errors.New("unknown private key type")
}

func (a *Account) GetEmail() string {
	return a.Email
}
func (a *Account) GetRegistration() *registration.Resource {
	return a.Registration
}
func (a *Account) GetPrivateKey() crypto.PrivateKey {
	a.privateKey, _ = a.GetKey()
	return a.privateKey
}
