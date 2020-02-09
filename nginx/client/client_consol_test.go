package client

import (
	"github.com/ihaiker/aginx/storage/consul"
	"github.com/ihaiker/aginx/storage/file"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConsul(t *testing.T) {
	store, err := file.System()
	assert.Nil(t, err)

	api, err := NewClient(store)
	assert.Nil(t, err)

	consulStorage, err := consul.New("127.0.0.1:8500/aginx")
	assert.Nil(t, err)

	err = consulStorage.StoreConfiguration(api.Configuration())
	assert.Nil(t, err)
}
