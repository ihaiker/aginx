package dockerLabels

import (
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/go-connections/nat"
	"github.com/ihaiker/aginx/plugins"
	"strconv"
	"strings"
)

type UsePort struct {
	PublishedPort int
	InternalPort  int
}

func (up UsePort) PublishedNatPort() nat.Port {
	return nat.Port(strconv.Itoa(up.PublishedPort) + "/tcp")
}

func (up UsePort) InternalNatPort() nat.Port {
	return nat.Port(strconv.Itoa(up.InternalPort) + "/tcp")
}

func (up *UsePort) IsSet(portmap nat.PortMap, port int) bool {
	for p, bindings := range portmap {
		if p.Proto() == "tcp" {
			if p.Int() == port {
				up.InternalPort = port
				up.PublishedPort, _ = strconv.Atoi(bindings[0].HostPort)
				return true
			}

			for _, binding := range bindings {
				if binding.HostPort == strconv.Itoa(port) {
					up.InternalPort = p.Int()
					up.PublishedPort = port
					return true
				}
			}
		}
	}
	return false
}

func (self *DockerLabelsRegister) findContainerPort(container types.ContainerJSON, port int) (*UsePort, error) {
	usePort := &UsePort{}

	if port == 0 {
		//公开断就查询，第一个
		for p, binding := range container.HostConfig.PortBindings {
			if p.Proto() == "tcp" {
				if usePort.InternalPort != 0 {
					return nil, ErrExplicitlyPort
				}
				usePort.InternalPort = p.Int()
				usePort.PublishedPort, _ = strconv.Atoi(binding[0].HostPort)
			}
		}
		//expose 定义的端口查询，第一个
		if usePort.InternalPort == 0 {
			for p, _ := range container.Config.ExposedPorts {
				if p.Proto() == "tcp" {
					if usePort.InternalPort != 0 {
						return nil, ErrExplicitlyPort
					}
					usePort.InternalPort = p.Int()
				}
			}
		}
		//未找到对应的端口
		if usePort.InternalPort == 0 {
			return nil, ErrExplicitlyPort
		}
	} else {
		//公开断就查询
		usePort.IsSet(container.HostConfig.PortBindings, port)

		if usePort.InternalPort == 0 {
			usePort.InternalPort = port
		}
	}
	return usePort, nil
}

func (self *DockerLabelsRegister) findFromContainer(containerId string) (plugins.Domains, error) {
	if container, info, err := self.docker.ContainerInspect(containerId); err != nil {
		return nil, err
	} else if container.Config.NetworkDisabled || container.HostConfig.NetworkMode.IsNone() {
		return nil, nil
	} else if labs := FindLabels(container.Config.Labels, true); labs.Has() {
		domains := plugins.Domains{}

		for _, label := range labs {
			usePort, err := self.findContainerPort(container, label.Port)
			if err != nil {
				return nil, err
			}
			domain := plugins.Domain{
				ID: containerId, Domain: label.Domain,
				Weight: label.Weight, AutoSSL: label.AutoSSL, Attrs: container.Config.Labels,
			}

			if label.Internal && !container.HostConfig.NetworkMode.IsHost() {
				if label.Networks != "" {
					for name, network := range container.NetworkSettings.Networks {
						if name == label.Networks || strings.HasPrefix(network.IPAddress, label.Networks) {
							domain.Address = fmt.Sprintf("%s:%d", network.IPAddress, usePort.InternalPort)
							break
						}
					}
				}
				if domain.Address == "" {
					for _, network := range container.NetworkSettings.Networks {
						domain.Address = fmt.Sprintf("%s:%d", network.IPAddress, usePort.InternalPort)
						break
					}
				}
			}
			if domain.Address == "" && container.HostConfig.NetworkMode.IsHost() {
				domain.Address = fmt.Sprintf("%s:%d", info.ip, usePort.InternalPort)
			}

			if domain.Address == "" && usePort.PublishedPort != 0 && info.ip != "" {
				domain.Address = fmt.Sprintf("%s:%d", info.ip, usePort.PublishedPort)
			}

			if domain.Address == "" {
				host := container.NetworkSettings.Networks[string(container.HostConfig.NetworkMode)].IPAddress
				domain.Address = fmt.Sprintf("%s:%d", host, usePort.InternalPort)
			}

			domains = append(domains, domain)
		}
		return domains, nil
	}
	return nil, nil
}
