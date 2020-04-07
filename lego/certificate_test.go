package lego

import (
	fileStorage "github.com/ihaiker/aginx/storage/file"
	"github.com/ihaiker/aginx/util"
	"github.com/kr/pretty"
	"math/rand"
	"os"
	"testing"
	"time"
)

var cfs *CertificateStorage
var acs *AccountStorage

func init() {
	rand.Seed(time.Now().Unix())
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
		t.Fatal(util.ErrNotFound)
	}
	domain := "who.renzhen.la"
	cert, err := cfs.New(account, domain, ":5002")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(cert)
}
