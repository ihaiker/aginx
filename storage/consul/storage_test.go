package consul

import (
	"github.com/ihaiker/aginx/logs"
	"github.com/ihaiker/aginx/server/ignore"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	url2 "net/url"
	"strconv"
	"testing"
	"time"
)

func init() {
	logs.SetLevel(logrus.DebugLevel)
}

func newClient(t *testing.T) *consulStorage {
	url, _ := url2.Parse("consul://127.0.0.1:8500/aginx")
	engine, _ := New(url, ignore.Empty())
	return engine
}

func TestEngine(t *testing.T) {
	api := newClient(t)

	err := api.Put("nginx.conf", []byte("nginx configuration 2."))
	assert.Nil(t, err)

	reader, err := api.Get("nginx.conf")
	assert.Nil(t, err, "get file")

	bs, err := ioutil.ReadAll(reader)
	assert.Nil(t, err, "reader error")

	t.Log(string(bs))
}

func TestKeys(t *testing.T) {
	api := newClient(t)

	readers, err := api.Search("*")
	assert.Nil(t, err)
	t.Log(readers)
}

func TestAccounts(t *testing.T) {
	api := newClient(t)

	readers, err := api.Search("lego/accounts/*")
	assert.Nil(t, err)
	t.Log(readers)
}

func TestStart(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	api := newClient(t)

	err := api.Start()
	assert.Nil(t, err)
	defer func() { _ = api.Stop() }()

	time.Sleep(time.Hour * 7)
}

func TestRemove(t *testing.T) {
	api := newClient(t)

	for i := 0; i < 10; i++ {
		err := api.Put("test/nginx"+strconv.Itoa(i)+".conf", []byte("nginx configuration ."+strconv.Itoa(i)))
		assert.Nil(t, err)
	}

	t.Log(api.Remove("test/nginx0.conf"))

	t.Log(api.Remove("test"))
}
