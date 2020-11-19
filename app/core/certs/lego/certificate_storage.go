package lego

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"github.com/go-acme/lego/v3/certificate"
	"github.com/go-acme/lego/v3/challenge"
	"github.com/go-acme/lego/v3/challenge/http01"
	"github.com/go-acme/lego/v3/lego"
	"github.com/ihaiker/aginx/v2/api"
	"github.com/ihaiker/aginx/v2/plugins/storage"
	"net"
	"path/filepath"
	"time"
)

//证书管理器
type certificateStorage struct {
	certDir string
	data    map[string]*Certificate
	aginx   api.Aginx
}

func (cfs *certificateStorage) Get(domain string) (cert *Certificate, has bool) {
	cert, has = cfs.data[domain]
	return
}

func (cfs *certificateStorage) NewWithProvider(account *Account, domain string, provider challenge.Provider) (*Certificate, error) {
	config := lego.NewConfig(account)
	config.Certificate.KeyType = account.KeyType
	config.Certificate.Timeout = time.Minute

	client, err := lego.NewClient(config)
	if err != nil {
		return nil, err
	}
	_ = client.Challenge.SetHTTP01Provider(provider)

	if res, err := client.Certificate.Obtain(certificate.ObtainRequest{
		Domains: []string{domain}, Bundle: true}); err != nil {
		return nil, err
	} else {
		cert := &Certificate{
			Email:         account.Email,
			Domain:        domain,
			CertURL:       res.CertURL,
			CertStableURL: res.CertStableURL,
		}
		cert.CertificatePath = filepath.Join(cfs.certDir, domain, "server.crt")
		cert.PrivateKeyPath = filepath.Join(cfs.certDir, domain, "server.key")

		{
			block, _ := pem.Decode([]byte(res.Certificate))
			certInfo, _ := x509.ParseCertificate(block.Bytes)
			cert.ExpireTime = certInfo.NotAfter
		}

		domainFile := filepath.Join(cfs.certDir, domain+".json")
		//存储域名说明文件
		bs, _ := json.MarshalIndent(cert, "", "\t")
		if err := cfs.aginx.Files().NewWithContent(domainFile, bs); err != nil {
			return nil, err
		}

		if err = cfs.aginx.Files().NewWithContent(cert.CertificatePath, res.Certificate); err != nil {
			return nil, err
		}
		if err = cfs.aginx.Files().NewWithContent(cert.PrivateKeyPath, res.PrivateKey); err != nil {
			return nil, err
		}
		cfs.data[domain] = cert
		return cert, nil
	}
}

func (cfs *certificateStorage) New(account *Account, domain, address string) (*Certificate, error) {
	if host, port, err := net.SplitHostPort(address); err != nil {
		return nil, err
	} else {
		provider := http01.NewProviderServer(host, port)
		return cfs.NewWithProvider(account, domain, provider)
	}
}

func (cfs *certificateStorage) List() (map[string]*Certificate, error) {
	return cfs.data, nil
}

func loadCertificates(dir string, aginx api.Aginx) (cs *certificateStorage, err error) {
	cs = &certificateStorage{
		data:  map[string]*Certificate{},
		aginx: aginx, certDir: dir,
	}

	var files []*storage.File
	if files, err = aginx.Files().Search(filepath.Join(dir, "*.json")); err != nil {
		return
	}
	for _, file := range files {
		if file.Name == "account.json" {
			continue
		}
		path := file.Name
		keyBytes := file.Content

		cert := new(Certificate)
		err = json.Unmarshal(keyBytes, cert)
		if err == nil {
			cs.data[cert.Domain] = cert
			logger.Debug("加载证书 ", path)
		} else {
			logger.WithError(err).Warn("加载证书 ", path)
		}
	}
	return
}
