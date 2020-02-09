package util

import (
	"testing"
)

func TestEventbus(t *testing.T) {

	EBus.Subscribe("test", func(i int) {
		t.Log("test 1 ", i)
	})

	EBus.Subscribe("test", func(i int) error {
		t.Log("test 2 ", i)
		return nil
	})

	EBus.Publish("test", 111)
	EBus.Publish("test", 222)

	EBus.WaitAsync()
}
