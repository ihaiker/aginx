package consulTemplate

import (
	"fmt"
	consulApi "github.com/hashicorp/consul/api"
	"testing"
)

func TestNewLabelRegister(t *testing.T) {
	consul, err := consulApi.NewClient(consulApi.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}

	reg := NewTemplateRegister(consul, []string{".*"})
	if err = reg.Start(); err != nil {
		t.Fatal(err)
	}

	events := reg.Listener()
	for {
		select {
		case event, has := <-events:
			if has {
				val := event.(*ConsulTemplateEvent)
				for name, entries := range val.Services {
					fmt.Print(name)
					fmt.Print(" (")
					for _, entry := range entries {
						fmt.Print(" ", entry.Service.ID)
					}
					fmt.Println(" )")
				}
			}
		}
	}
}
