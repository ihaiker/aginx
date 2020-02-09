package consul

import (
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
	"time"
)

func TestEngine(t *testing.T) {
	api, err := New("127.0.0.1:8500", "aginx", "")
	assert.Nil(t, err)

	err = api.Store("nginx.conf", []byte("nginx configuration 2."))
	assert.Nil(t, err)

	reader, err := api.File("nginx.conf")
	assert.Nil(t, err, "get file")

	bs, err := ioutil.ReadAll(reader)
	assert.Nil(t, err, "reader error")

	t.Log(string(bs))
}

func TestKeys(t *testing.T) {
	api, err := New("127.0.0.1:8500", "aginx", "")
	assert.Nil(t, err)

	readers, err := api.Search("*")
	assert.Nil(t, err)
	t.Log(readers)
}

func TestAccounts(t *testing.T) {
	api, err := New("127.0.0.1:8500", "aginx", "")
	assert.Nil(t, err)

	readers, err := api.Search("lego/accounts/*")
	assert.Nil(t, err)
	t.Log(readers)
}

func TestStart(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)

	api, err := New("127.0.0.1:8500", "aginx", "")
	assert.Nil(t, err)

	err = api.Start()
	assert.Nil(t, err)
	defer func() { _ = api.Stop() }()

	time.Sleep(time.Second * 7)
}
