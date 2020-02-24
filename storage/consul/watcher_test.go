package consul

import (
	"sync"
	"testing"
	"time"
)

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
