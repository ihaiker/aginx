package lego

import (
	"github.com/go-acme/lego/v3/certcrypto"
	fileStorage "github.com/ihaiker/aginx/storage/file"
	"github.com/kr/pretty"
	"github.com/sirupsen/logrus"
	"math/rand"
	"os"
	"testing"
	"time"
)

var cfs *CertificateStorage
var acs *AccountStorage

func init() {
	rand.Seed(time.Now().Unix())
	logrus.SetLevel(logrus.DebugLevel)

	pwd, _ := os.Getwd()
	engine := fileStorage.New(pwd + "/nginx.conf")

	cfs, _ = LoadCertificates(engine)
	acs, _ = LoadAccounts(engine)
}

func TestCertificateStorage_Get(t *testing.T) {
	cert, _ := cfs.Get("ni.renzhen.la")
	_, _ = pretty.Println(cert)
}

func TestNewDomain(t *testing.T) {
	account, has := acs.Get("who@renzhen.la")
	if !has {
		t.Fatal(os.ErrNotExist)
	}
	domain := "who.renzhen.la"
	keyType := certcrypto.EC384
	cert, err := cfs.New(keyType, account, domain, ":5002")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(cert)
}

func TestStoreFile(t *testing.T) {
	cwd, _ := os.Getwd()
	domain := "who.renzhen.la"
	cert, _ := cfs.Get(domain)
	if err := cert.StoreFile(cwd + "/" + domain); err != nil {
		t.Fatal(err)
	}
}
