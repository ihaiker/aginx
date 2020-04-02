package dockerLabels

import (
	"fmt"
	"github.com/ihaiker/aginx/util"
	"testing"
)

func TestNewDockerWrapper(t *testing.T) {
	docker, err := NewDockerWrapper(util.GetRecommendIp()[0], true)
	if err != nil {
		t.Fatal(err)
	}
	defer docker.Stop()
	for event := range docker.Events {
		fmt.Println(event)
	}
}
