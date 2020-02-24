package lego

import (
	"encoding/json"
	"github.com/go-acme/lego/v3/certificate"
	"github.com/go-acme/lego/v3/challenge"
	"github.com/go-acme/lego/v3/challenge/http01"
	"github.com/go-acme/lego/v3/lego"
	"github.com/ihaiker/aginx/storage"
	"github.com/ihaiker/aginx/util"
	"io/ioutil"
	"net"
	"time"
)

const certificateDir = "lego/certificates"

type CertificateStorage struct {
	data   map[string]*Certificate
	engine storage.Engine
}

func (cfs *CertificateStorage) Get(domain string) (cert *Certificate, has bool) {
	cert, has = cfs.data[domain]
	return
}

func (cfs *CertificateStorage) NewWithProvider(account *Account, domain string, provider challenge.Provider) (cert *Certificate, err error) {
	config := lego.NewConfig(account)
	config.Certificate.KeyType = account.KeyType
	config.Certificate.Timeout = time.Minute

	var client *lego.Client
	if client, err = lego.NewClient(config); err != nil {
		return
	}

	if err = client.Challenge.SetHTTP01Provider(provider); err != nil {
		return
	}

	cert = new(Certificate)
	if cert.Resource, err = client.Certificate.Obtain(certificate.ObtainRequest{
		Domains: []string{domain}, Bundle: true}); err != nil {
		return
	} else if err = cert.LoadCertificate(); err != nil {
		return
	}

	cert.Email = account.Email
	cert.ExpireTime = time.Now().AddDate(0, 3, 0)

	if _, err = cert.StoreFile(cfs.engine); err != nil {
		return nil, err
	}

	cfs.data[domain] = cert
	if err = cfs.restore(domain); err != nil {
		delete(cfs.data, domain)
		return
	}
	return
}

func (cfs *CertificateStorage) New(account *Account, domain, address string) (*Certificate, error) {
	if iface, port, err := net.SplitHostPort(address); err != nil {
		return nil, err
	} else {
		provider := http01.NewProviderServer(iface, port)
		return cfs.NewWithProvider(account, domain, provider)
	}
}

func (cfs *CertificateStorage) restore(domain string) error {
	cert, _ := cfs.Get(domain)
	file := certificateDir + "/" + domain + ".json"
	bs, err := json.MarshalIndent(cert, "", "\t")
	if err != nil {
		return err
	}
	if err := cfs.engine.Put(file, bs); err != nil {
		return err
	}
	return nil
}

func LoadCertificates(engine storage.Engine) (certificateStorage *CertificateStorage, err error) {
	certificateStorage = &CertificateStorage{
		data: map[string]*Certificate{}, engine: engine,
	}
	var readers []*util.NameReader
	if readers, err = engine.Search(certificateDir + "/*.json"); err != nil {
		return
	}

	for _, reader := range readers {
		path := reader.Name
		keyBytes, err := ioutil.ReadAll(reader)
		if err != nil {
			return nil, err
		}

		cert := new(Certificate)
		err = json.Unmarshal(keyBytes, cert)
		if err == nil {
			certificateStorage.data[cert.Domain] = cert
		}
		logrus.WithError(err).Debug("load certificate ", path)
	}
	return
}
