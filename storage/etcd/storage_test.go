package etcd

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

func newClient(t *testing.T) *etcdV3Storage {
	url, _ := url2.Parse("etcd://127.0.0.1:2379/aginx")
	engine, _ := New(url, ignore.Empty())
	return engine
}

func TestPut(t *testing.T) {
	api := newClient(t)
	err := api.Store("nginx.conf", []byte("etcd configuration "+time.Now().Format(time.RFC3339)))

	if err != nil {
		t.Fatal(err)
	}
}

func TestGet(t *testing.T) {
	api := newClient(t)

	r, err := api.File("nginx.conf")
	if err != nil {
		t.Fatal(err)
	}
	bs, err := ioutil.ReadAll(r)
	t.Log(string(bs))
}

func TestList(t *testing.T) {
	api := newClient(t)

	if err := api.Start(); err != nil {
		t.Fatal(err)
	}
	_ = api.Stop()
}

func TestSearch(t *testing.T) {
	api := newClient(t)

	readers, err := api.Search("hosts.d/*.conf")
	if err != nil {
		t.Fatal(err)
	}
	for _, reader := range readers {
		t.Log(reader.Name)
	}
}

func TestRemove(t *testing.T) {
	api := newClient(t)

	for i := 0; i < 10; i++ {
		err := api.Store("test/nginx"+strconv.Itoa(i)+".conf", []byte("nginx configuration ."+strconv.Itoa(i)))
		assert.Nil(t, err)
	}

	t.Log(api.Remove("test/nginx0.conf"))
	t.Log(api.Remove("test"))
}
