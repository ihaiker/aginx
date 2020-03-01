package dockerTemplates

import (
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	dockerClient "github.com/docker/docker/client"
)

type DockerTemplateRegisterEvents struct {
	Containers []types.ContainerJSON
	Services   []swarm.Service
	PublishIP  string
	Nodes      []swarm.Node
	Docker     *dockerClient.Client
}
