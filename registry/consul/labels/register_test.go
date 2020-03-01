package consulLabels

import (
	consulApi "github.com/hashicorp/consul/api"
	"github.com/kr/pretty"
	"testing"
)

func TestNewLabelRegister(t *testing.T) {
	consul, err := consulApi.NewClient(consulApi.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}

	reg := NewLabelRegister(consul)
	if err = reg.Start(); err != nil {
		t.Fatal(err)
	}

	events := reg.Listener()
	for {
		select {
		case event, has := <-events:
			if has {
				pretty.Println(event)
			}
		}
	}
}
