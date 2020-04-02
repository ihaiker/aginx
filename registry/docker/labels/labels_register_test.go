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
		"aginx.domain.8080": "www.b.com,networks=10.24.0",
	}, true)
	t.Log(lables.String())
}

func TestDocker(t *testing.T) {
	docker, err := LabelsRegister("10.24.0.1")
	PanicIfError(err)

	servers := docker.allDomains()

	for s, d := range servers.Group() {
		fmt.Println("Domain ", s)
		for _, server := range d {
			fmt.Println("\t", server.Domain, server.Address, ", Weight:", server.Weight, ", ssl:", server.AutoSSL)
		}
	}
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
