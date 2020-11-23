package lego

import (
	"fmt"
	"github.com/go-acme/lego/v3/certcrypto"
	"github.com/ihaiker/aginx/v2/api"
	"github.com/ihaiker/aginx/v2/plugins/certificate"
	"net/url"
	"path/filepath"
	"strings"
)

type legoCert struct {
	account *Account
	certs   *certificateStorage
	aginx   api.Aginx
}

func LoadCertificates() *legoCert {
	return &legoCert{}
}

func (l *legoCert) Scheme() string {
	return "lego"
}

func (l *legoCert) Name() string {
	return "Let's Encrypt API"
}

func (l *legoCert) Version() string {
	return "v2.0.0"
}

func (l *legoCert) Help() string {
	return `lego证书提供商
lego 使用ACMEApi申请免费证书。
配置方式： lego://<email>/<storage path>
例如(默认启用)： lego://aginx@renzhen.la/certs/lego  
email：          aginx@renzhen.la
storage path:   certs/lego
`
}

func (l *legoCert) dir(config url.URL) string {
	dir := config.Path
	if dir == "" || dir == "/" {
		dir = "certs/lego"
	}
	if strings.HasPrefix(dir, "/") {
		dir = dir[1:]
	}
	return dir
}

func (l *legoCert) email(config url.URL) string {
	if config.User == nil {
		return ""
	}
	return config.User.Username() + "@" + config.Host
}
func (l *legoCert) keyType(config url.URL) certcrypto.KeyType {
	keyType := certcrypto.KeyType(config.Query().Get("keyType"))
	if keyType == "" {
		keyType = certcrypto.EC384
	}
	return keyType
}

func (l *legoCert) Initialize(config url.URL, aginx api.Aginx) (err error) {
	email := l.email(config)
	baseDir := l.dir(config)
	keyType := l.keyType(config)

	var accountStorage *accountStorage
	if accountStorage, err = loadAccounts(filepath.Join(baseDir, "accounts"), aginx); err != nil {
		return
	} else if l.account, err = accountStorage.New(email, keyType); err != nil {
		return
	}

	if l.certs, err = loadCertificates(filepath.Join(baseDir, "certificates"), aginx); err != nil {
		return
	}

	l.aginx = aginx
	return
}

func (l *legoCert) New(domain string) (certificate.Files, error) {
	return l.certs.NewWithProvider(l.account, domain, newProvider(l.aginx))
}

func (l *legoCert) Get(domain string) (certificate.Files, error) {
	if cert, has := l.certs.Get(domain); !has {
		return nil, fmt.Errorf("not found: %s", domain)
	} else {
		return cert, nil
	}
}

func (l *legoCert) List() (map[string]certificate.Files, error) {
	files, err := l.certs.List()
	if err != nil {
		return nil, err
	}
	m := map[string]certificate.Files{}
	for s, c := range files {
		m[s] = c
	}
	return m, nil
}
