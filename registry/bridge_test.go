package registry_test

import (
	"github.com/ihaiker/aginx/api"
	"github.com/ihaiker/aginx/registry"
	"github.com/ihaiker/aginx/registry/docker"
	"testing"
)

func TestBridge(t *testing.T) {
	reg, _ := docker.Registor("10.24.0.1")
	aginx := api.New(":8011")

	bridge := registry.RegistorBridge{
		Registrator:      reg,
		LocalTemplateDir: "template",
		Aginx:            aginx,
	}
	if err := bridge.Start(); err != nil {
		t.Fatal(err)
	}
}
