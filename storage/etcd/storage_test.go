package etcd

import (
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"testing"
	"time"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

func TestPut(t *testing.T) {
	api, err := New("127.0.0.1:2379", "aginx", "", "")
	if err != nil {
		t.Fatal(err)
	}
	err = api.Store("nginx.conf", []byte("etcd configuration "+time.Now().Format(time.RFC3339)))

	if err != nil {
		t.Fatal(err)
	}
}

func TestGet(t *testing.T) {
	api, err := New("127.0.0.1:2379", "aginx", "", "")
	if err != nil {
		t.Fatal(err)
	}
	r, err := api.File("nginx.conf")
	if err != nil {
		t.Fatal(err)
	}
	bs, err := ioutil.ReadAll(r)
	t.Log(string(bs))
}

func TestList(t *testing.T) {
	api, err := New("127.0.0.1:2379", "aginx", "", "")
	if err != nil {
		t.Fatal(err)
	}

	if err = api.Start(); err != nil {
		t.Fatal(err)
	}
	_ = api.Stop()
}

func TestSearch(t *testing.T) {
	api, err := New("127.0.0.1:2379", "aginx", "", "")
	if err != nil {
		t.Fatal(err)
	}
	readers, err := api.Search("hosts.d/*.conf")
	if err != nil {
		t.Fatal(err)
	}
	for _, reader := range readers {
		t.Log(reader.Name)
	}
}
