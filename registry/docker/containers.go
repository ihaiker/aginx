package docker

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/go-connections/nat"
	"github.com/ihaiker/aginx/plugins"
	"strconv"
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

func (self *DockerRegistor) findContainerPort(container types.ContainerJSON, port int) (*UsePort, error) {
	usePort := &UsePort{}
	//单价欧
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

func (self *DockerRegistor) findFromContainer(containerId string) (plugins.Domains, error) {
	if container, err := self.docker.ContainerInspect(context.TODO(), containerId); err != nil {
		return nil, err
	} else if labs := findLabels(container.Config.Labels, true); labs.Has() {
		domains := plugins.Domains{}
		for port, label := range labs {
			usePort, err := self.findContainerPort(container, port)
			if err != nil {
				return nil, err
			}
			domain := plugins.Domain{
				ID: containerId, Domain: label.Domain,
				Weight: label.Weight, AutoSSL: label.AutoSSL, Attrs: container.Config.Labels,
			}
			if usePort.PublishedPort != 0 && self.ip != "" {
				domain.Address = self.ip + ":" + strconv.Itoa(usePort.PublishedPort)
			}

			if label.Internal || domain.Address == "" {
				nm := container.HostConfig.NetworkMode
				if nm != "bridge" && nm != "default" && nm != "host" {
					domain.Address = fmt.Sprintf("%s:%d", container.NetworkSettings.Networks[string(nm)].IPAddress, usePort.InternalPort)
				} else {
					for _, network := range container.NetworkSettings.Networks {
						domain.Address = fmt.Sprintf("%s:%d", network.IPAddress, usePort.InternalPort)
					}
				}
			}
			domains = append(domains, domain)
		}
		return domains, nil
	}
	return nil, nil
}
