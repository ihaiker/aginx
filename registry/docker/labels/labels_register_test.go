package dockerLabels

import (
	"fmt"
	"github.com/ihaiker/aginx/logs"
	"github.com/ihaiker/aginx/plugins"
	. "github.com/ihaiker/aginx/util"
	"github.com/sirupsen/logrus"
	"testing"
)

func init() {
	logs.SetLevel(logrus.DebugLevel)
}

func TestLabel(t *testing.T) {
	lables := FindLabels(map[string]string{
		"aginx.domain":      "a.com;b.com",
		"aginx.domain.8080": "www.b.com,networks=ext-portainer-network",
	}, true)
	t.Log(lables.String())
}

func TestDocker(t *testing.T) {
	docker, err := LabelsRegister("172.16.100.10", false)
	PanicIfError(err)

	if err := docker.Start(); err != nil {
		fmt.Println(err)
	}

	for event := range docker.Listener() {
		fmt.Println("============================================")
		servers := event.(plugins.LabelsRegistryEvent)
		for domain, servers := range servers {
			fmt.Println("Domain: ", domain)
			for _, server := range servers {
				fmt.Println("\t", server.Domain, server.Address, ", Weight:", server.Weight, ", ssl:", server.AutoSSL)
			}
		}
	}
}
