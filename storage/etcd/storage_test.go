package etcd

import (
	"github.com/ihaiker/aginx/logs"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	url2 "net/url"
	"strconv"
	"sync"
	"testing"
	"time"
)

func init() {
	logs.SetLevel(logrus.DebugLevel)
}

func newClient(t *testing.T) *etcdV3Storage {
	url, _ := url2.Parse("etcd://127.0.0.1:2379/aginx")
	engine, _ := New(url)
	return engine
}

func TestPut(t *testing.T) {
	api := newClient(t)
	err := api.Put("nginx.conf", []byte("etcd configuration "+time.Now().Format(time.RFC3339)))

	if err != nil {
		t.Fatal(err)
	}
}

func TestGet(t *testing.T) {
	api := newClient(t)
	r, err := api.Get("nginx.conf")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r.Content)
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
		err := api.Put("test/nginx"+strconv.Itoa(i)+".conf", []byte("nginx configuration ."+strconv.Itoa(i)))
		assert.Nil(t, err)
	}

	t.Log(api.Remove("test/nginx0.conf"))
	t.Log(api.Remove("test"))
}

func TestAll(t *testing.T) {
	api := newClient(t)
	files, err := api.Search()
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		t.Log(file)
	}
}

func TestWatcher(t *testing.T) {
	api := newClient(t)

	gw := new(sync.WaitGroup)
	gw.Add(1)
	go func() {
		defer gw.Done()
		time.Sleep(time.Second)

		gw.Add(1)
		if err := api.Put("test", []byte("123")); err != nil {
			gw.Done()
		}

		time.Sleep(time.Second)
		gw.Add(1)
		if err := api.Put("test", []byte("345")); err != nil {
			gw.Done()
		}

		time.Sleep(time.Second)
		gw.Add(1)
		if err := api.Remove("test"); err != nil {
			gw.Done()
		}
	}()

	go func() {
		events := api.StartListener()
		for event := range events {
			t.Log(event.String())
			gw.Done()
		}
	}()

	gw.Wait()
}
