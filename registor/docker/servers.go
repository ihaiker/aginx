package docker

import (
	"errors"
	dockerapi "github.com/fsouza/go-dockerclient"
	"github.com/ihaiker/aginx/registor"
	"os"
	"strconv"
	"strings"
)

func serverPort(port int, container *dockerapi.Container) (dockerapi.Port, error) {
	if port != 0 {
		return dockerapi.Port(strconv.Itoa(port) + "/tcp"), nil
	}

	for p, _ := range container.HostConfig.PortBindings {
		if strings.HasSuffix(string(p), "/tcp") {
			if port != 0 {
				return "", errors.New("Port not explicitly specified")
			}
			port, _ = strconv.Atoi(strings.Replace(string(p), "/tcp", "", 1))
		}
	}
	if port == 0 {
		for p, _ := range container.Config.ExposedPorts {
			if strings.HasSuffix(string(p), "/tcp") {
				if port != 0 {
					return "", errors.New("Port not explicitly specified")
				}
				port, _ = strconv.Atoi(strings.Replace(string(p), "/tcp", "", 1))
			}
		}
	}

	return dockerapi.Port(strconv.Itoa(port) + "/tcp"), nil
}

func getServer(ip string, docker *dockerapi.Client, containerId string) (registor.Servers, error) {
	if container, err := docker.InspectContainerWithOptions(dockerapi.InspectContainerOptions{ID: containerId}); err != nil {
		return nil, err
	} else if labs := findLabels(container.Config.Labels); labs.Has() {
		domains := registor.Servers{}
		for _, label := range labs {
			if port, err := serverPort(label.Port, container); err != nil {
				return nil, err
			} else {

				if !label.Internal { //use bind port
					if bindings, has := container.HostConfig.PortBindings[port]; has {
						server := &DockerServer{
							IDAtr:         containerId,
							DomainAtr:     label.Domain,
							AddressAtr:    bindings[0].HostIP + ":" + bindings[0].HostPort,
							WeightAtr:     label.Weight,
							ContainerName: container.Name,
						}
						if ip != "" {
							server.AddressAtr = ip + ":" + bindings[0].HostPort
						}
						domains = append(domains, server)
						continue
					}
				}

				if _, has := container.NetworkSettings.Ports[port]; has {
					hostIp := container.NetworkSettings.IPAddress
					nm := container.HostConfig.NetworkMode
					if nm != "bridge" && nm != "default" && nm != "host" {
						hostIp = container.NetworkSettings.Networks[nm].IPAddress
					} else {
						for _, network := range container.NetworkSettings.Networks {
							hostIp = network.IPAddress
						}
					}
					domains = append(domains, &DockerServer{
						IDAtr:         containerId,
						DomainAtr:     label.Domain,
						AddressAtr:    hostIp + ":" + port.Port(),
						WeightAtr:     label.Weight,
						ContainerName: container.Name,
					})
					continue
				}
			}
		}
		return domains, nil
	}
	return nil, os.ErrNotExist
}
