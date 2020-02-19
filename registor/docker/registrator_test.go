package docker

import (
	"fmt"
	"github.com/ihaiker/aginx/logs"
	"github.com/ihaiker/aginx/registor"
	. "github.com/ihaiker/aginx/util"
	"github.com/sirupsen/logrus"
	"testing"
)

func init() {
	logs.SetLevel(logrus.DebugLevel)
}

func TestDocker(t *testing.T) {
	docker, err := Registrator("10.24.0.1")
	PanicIfError(err)

	domains, err := docker.Sync()
	PanicIfError(err)

	for s, d := range domains.Group() {
		fmt.Println("DomainAtr: ", s)
		for _, domain := range d {
			fmt.Println("\t", domain.(*DockerServer).ContainerName, domain.Domain(), domain.Address(), ", WeightAtr:", domain.Weight())
		}
	}
	if err := docker.Start(); err != nil {
		fmt.Println(err)
	}
	for event := range docker.Listener() {
		fmt.Println("============================================")
		fmt.Println("event: ", event.EventType == registor.Online)
		for domain, servers := range event.Servers.Group() {
			fmt.Println("DomainAtr: ", domain)
			for _, server := range servers {
				fmt.Println("\t", server.(*DockerServer).String())
			}
		}
	}
}
