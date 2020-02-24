package zookeeper

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

func newClient(t *testing.T) *zkStorage {
	url, _ := url2.Parse("zk://127.0.0.1:2181/aginx")
	engine, _ := New(url)
	return engine
}

func TestStore(t *testing.T) {
	api := newClient(t)

	err := api.Put("nginx.conf", []byte("zookeeper configuration "+time.Now().Format(time.RFC3339)))
	assert.Nil(t, err)

	err = api.Put("lego/nginx.conf", []byte("zookeeper configuration "+time.Now().Format(time.RFC3339)))
	assert.Nil(t, err)
}

func TestSearch(t *testing.T) {
	api := newClient(t)

	files, err := api.Search("*")
	assert.Nil(t, err)

	for _, file := range files {
		t.Log(file.Name)
	}
}

func TestStart(t *testing.T) {
	api := newClient(t)
	err := api.Start()
	assert.Nil(t, err)
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

func TestList(t *testing.T) {
	api := newClient(t)
	if files, err := api.List(); err != nil {
		t.Fatal(err)
	} else {
		for _, file := range files {
			t.Log(file)
		}
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
