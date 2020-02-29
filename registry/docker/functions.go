package docker

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	dockerClient "github.com/docker/docker/client"
	"github.com/ihaiker/aginx/nginx"
	"strings"
	"text/template"
)

func nodeHasLabel(node swarm.Node, label string) bool {
	_, has := node.Spec.Labels[label]
	return has
}

func nodeIsWorker(node swarm.Node) bool {
	return node.Spec.Role == swarm.NodeRoleWorker
}

func nodeAvailability(node swarm.Node) swarm.NodeAvailability {
	return node.Spec.Availability
}

func serviceVirtualAddress(service swarm.Service, port int) []string {
	address := make([]string, 0)
	for _, virtualIP := range service.Endpoint.VirtualIPs {
		addr := virtualIP.Addr
		if idx := strings.Index(addr, "/"); idx != -1 {
			address = append(address, fmt.Sprintf("%s:%d", addr[0:idx], port))
		} else {
			address = append(address, fmt.Sprintf("%s:%d", addr, port))
		}
	}
	return address
}

func serviceInternalAddress(docker *dockerClient.Client, service swarm.Service, port int) []string {
	serviceName := service.Spec.Name
	tasks, _ := docker.TaskList(context.TODO(), types.TaskListOptions{
		Filters: filters.NewArgs(filters.Arg("desired-state", "running"), filters.Arg("service", serviceName))})

	addresses := make([]string, 0)
	for _, task := range tasks {
		for _, attachment := range task.NetworksAttachments {
			for _, address := range attachment.Addresses {
				idx := strings.Index(address, "/")
				addresses = append(addresses, fmt.Sprintf("%s:%d", address[0:idx], port))
			}
		}
	}
	return addresses
}

func servicePublishedPort(service swarm.Service, internalPort int) int {
	for _, port := range service.Endpoint.Ports {
		if port.TargetPort == uint32(internalPort) {
			return int(port.PublishedPort)
		}
	}
	return 0
}

func serviceHasLabel(service swarm.Service, label string) bool {
	_, has := service.Spec.Labels[label]
	return has
}

func containerAddress(container types.ContainerJSON, network string) string {
	if nw, has := container.NetworkSettings.Networks[network]; has {
		return nw.IPAddress
	} else {
		return ""
	}
}

func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"upstreamName": nginx.UpstreamName,

		"nodeIsWorker":     nodeIsWorker,
		"nodeHasLabel":     nodeHasLabel,
		"nodeAvailability": nodeAvailability,
		"nodeHasAnyLabels": func(node swarm.Node, labels ...string) bool {
			for _, label := range labels {
				if nodeHasLabel(node, label) {
					return true
				}
			}
			return false
		},

		"serviceName": func(service swarm.Service) string {
			return service.Spec.Name
		},
		"serviceVirtualAddress":  serviceVirtualAddress,
		"serviceInternalAddress": serviceInternalAddress,
		"servicePublishedPort":   servicePublishedPort,
		"serviceHasLabel":        serviceHasLabel,
		"containerAddress":       containerAddress,
	}
}
