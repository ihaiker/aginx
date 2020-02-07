package lego

import (
	"encoding/json"
	"github.com/go-acme/lego/v3/certcrypto"
	"github.com/go-acme/lego/v3/certificate"
	"github.com/go-acme/lego/v3/challenge/http01"
	"github.com/go-acme/lego/v3/lego"
	"github.com/ihaiker/aginx/storage"
	"github.com/ihaiker/aginx/util"
	"github.com/sirupsen/logrus"
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

func (cfs *CertificateStorage) New(keyType certcrypto.KeyType, account *Account, domain, address string) (cert *Certificate, err error) {

	config := lego.NewConfig(account)
	config.Certificate.KeyType = keyType

	var client *lego.Client
	if client, err = lego.NewClient(config); err != nil {
		return
	}

	var iface, port string
	if iface, port, err = net.SplitHostPort(address); err != nil {
		return
	}
	if err = client.Challenge.SetHTTP01Provider(http01.NewProviderServer(iface, port)); err != nil {
		return
	}

	cert = new(Certificate)
	if cert.Resource, err = client.Certificate.Obtain(certificate.ObtainRequest{
		Domains: []string{domain}, Bundle: true}); err != nil {
		return
	} else if err = cert.LoadCertificate(); err != nil {
		return
	}
	cert.ExpireTime = time.Now().AddDate(0, 3, 0)

	cfs.data[domain] = cert
	if err = cfs.restore(domain); err != nil {
		delete(cfs.data, domain)
		return
	}
	return
}

func (cfs *CertificateStorage) restore(domain string) error {
	cert, _ := cfs.Get(domain)
	file := certificateDir + "/" + domain + ".json"
	bs, err := json.MarshalIndent(cert, "", "\t")
	if err != nil {
		return err
	}
	if err := cfs.engine.Store(file, bs); err != nil {
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
		logrus.WithField("file", reader.Name).WithField("module", "certificate").
			Debug("load certificate")

		keyBytes, err := ioutil.ReadAll(reader)
		if err != nil {
			return nil, err
		}

		cert := new(Certificate)
		err = json.Unmarshal(keyBytes, cert)
		if err == nil {
			certificateStorage.data[cert.Domain] = cert
		} else {
			logrus.WithField("file", path).WithField("module", "certificate").
				WithError(err).Error("load certificate file error")
		}
	}
	return
}
