package lego

import (
	"github.com/go-acme/lego/v3/certcrypto"
	fileStorage "github.com/ihaiker/aginx/storage/file"
	"math/rand"
	"os"
	"testing"
	"time"
)

var accountStorage *AccountStorage

func init() {
	rand.Seed(time.Now().Unix())

	pwd, _ := os.Getwd()
	engine := fileStorage.New(pwd + "/nginx.conf")
	accountStorage, _ = LoadAccounts(engine)
}

func TestPrivateKey(t *testing.T) {
	a, e := accountStorage.New("who@renzhen.la", certcrypto.EC384)
	t.Log(a, e)
	p := a.GetPrivateKey()
	t.Log(p)
}

func TestLoadPrivateKey(t *testing.T) {
	as, _ := accountStorage.Get("ni@renzhen.la")
	t.Log(as)
	_ = as.GetPrivateKey()
}
