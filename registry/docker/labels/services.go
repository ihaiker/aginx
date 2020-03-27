package dockerLabels

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	dockerClient "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/ihaiker/aginx/plugins"
	"github.com/ihaiker/aginx/util"
	"os"
	"strings"
	"time"
)

func (self *DockerLabelsRegister) firstPort(portSet nat.PortSet) int {
	for p, _ := range portSet {
		return p.Int()
	}
	return 0 //ignore
}

func (self *DockerLabelsRegister) nodeContainerImageInspect(nodeId, containerId string) (nat.PortSet, error) {
	node, _, err := self.docker.NodeInspectWithRaw(context.TODO(), nodeId)
	if err != nil {
		return nil, err
	}

	host := os.Getenv("DOCKER_HOST")
	port := host[strings.LastIndex(host, ":")+1:]
	dockerNode := fmt.Sprintf("tcp://%s:%s", node.Status.Addr, port)
	docker, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv, dockerClient.WithHost(dockerNode))
	containerInspect, err := docker.ContainerInspect(context.TODO(), containerId)
	if err != nil {
		return nil, err
	}
	imageId := containerInspect.Image
	imageInspect, _, err := docker.ImageInspectWithRaw(context.TODO(), imageId)
	if err != nil {
		return nil, err
	}
	return imageInspect.Config.ExposedPorts, nil
}

func (self *DockerLabelsRegister) imagePort(service swarm.Service) nat.PortSet {
	tasks, err := self.docker.TaskList(context.TODO(), types.TaskListOptions{Filters: filters.NewArgs(filters.Arg("service", service.Spec.Name))})
	if err != nil {
		return nat.PortSet{}
	}
	for _, task := range tasks {
		if task.Status.State == swarm.TaskStateRunning || task.Status.State == swarm.TaskStateShutdown {
			if inspect, err := self.nodeContainerImageInspect(task.NodeID, task.Status.ContainerStatus.ContainerID); err == nil {
				return inspect
			}
		}
	}
	return nat.PortSet{}
}

func (self *DockerLabelsRegister) getVirtualAddress(service swarm.Service, port uint32) (address string) {
	defer util.Catch(func(err error) {
		logger.Warn("get virtual address error: ", err)
		address = self.ip
	})

	host := ""
	//此处，可能获取不到端口因为可能服务还未起来，端口并未打开，所以ping在其后
	for _, vip := range service.Endpoint.VirtualIPs {
		ip := vip.Addr[0:strings.Index(vip.Addr, "/")]
		if util.SockTo(ip, port, time.Second) ||
			util.Ping(ip, 3, time.Millisecond*100, time.Second) {
			host = ip
			break
		}
	}

	if host == "" {
		host = util.GetRecommendIp()
	}
	address = fmt.Sprintf("%s:%d", host, port)
	return
}

func (self *DockerLabelsRegister) getServiceTaskAddress(service swarm.Service, port uint32) map[int]string {
	serviceName := service.Spec.Name
	tasks, _ := self.docker.TaskList(context.TODO(), types.TaskListOptions{
		Filters: filters.NewArgs(filters.Arg("desired-state", "running"), filters.Arg("service", serviceName))})

	addresses := map[int]string{}
	for _, task := range tasks {
		for _, attachment := range task.NetworksAttachments {
			for _, address := range attachment.Addresses {
				idx := strings.Index(address, "/")
				addresses[task.Slot] = fmt.Sprintf("%s:%d", address[0:idx], port)
			}
		}
	}
	return addresses
}

func (self *DockerLabelsRegister) makeDomain(service swarm.Service, lab Label, address string) plugins.Domain {
	domainLabel := map[string]string{}
	for k, v := range service.Spec.TaskTemplate.ContainerSpec.Labels {
		domainLabel[k] = v
	}
	for k, v := range service.Spec.Labels {
		domainLabel[k] = v
	}
	domain := plugins.Domain{
		ID: service.Spec.Name, Domain: lab.Domain,
		Weight: lab.Weight, AutoSSL: lab.AutoSSL,
		Attrs: domainLabel, Address: address,
	}
	return domain
}

func (self *DockerLabelsRegister) findFromServiceById(serverId string) (plugins.Domains, error) {
	if service, _, err := self.docker.ServiceInspectWithRaw(context.TODO(), serverId, types.ServiceInspectOptions{}); err != nil {
		return nil, err
	} else {
		return self.findFromService(service)
	}
}

func (self *DockerLabelsRegister) findFromService(service swarm.Service) (plugins.Domains, error) {
	serviceName := service.Spec.Name
	domains := plugins.Domains{}
	labs := FindLabels(service.Spec.TaskTemplate.ContainerSpec.Labels, false)

	if labs.Has() {
		for port, label := range labs {
			usePort := swarm.PortConfig{TargetPort: uint32(port)}

			if port == 0 { //未指定端口
				if len(service.Endpoint.Ports) == 0 { //未有公开端口
					imagePorts := self.imagePort(service)
					if len(imagePorts) == 1 {
						usePort = swarm.PortConfig{TargetPort: uint32(self.firstPort(imagePorts))}
					} else {
						logger.Warnf("server %s image has multi port, but not explicitly specified. ignore it. ", serviceName)
						return nil, ErrExplicitlyPort
					}
				} else if len(service.Endpoint.Ports) == 1 {
					usePort = service.Endpoint.Ports[0]
				} else {
					logger.Warnf("server %s image has multi port, but not explicitly specified. ignore it. ", serviceName)
					return nil, ErrExplicitlyPort
				}
			} else { //明确指定端口
				for _, pcfg := range service.Endpoint.Ports {
					if int(pcfg.PublishedPort) == port || int(pcfg.TargetPort) == port {
						usePort = pcfg
					}
				}
			}

			if label.Virtual {
				address := self.getVirtualAddress(service, usePort.TargetPort)
				domains = append(domains, self.makeDomain(service, label, address))
			} else if label.Internal {
				addresses := self.getServiceTaskAddress(service, usePort.TargetPort)
				for slot, address := range addresses {
					domain := self.makeDomain(service, label, address)
					domain.ID = fmt.Sprintf("%s:%d", serviceName, slot)
					domains = append(domains, domain)
				}
			} else if usePort.PublishedPort != uint32(0) {
				if label.Nodes {
					_, nodes, _ := self.getNodes(Normal(Or(isWorker, Not(isMulti))))
					for i, node := range nodes {
						address := fmt.Sprintf("%s:%d", node, usePort.PublishedPort)
						domain := self.makeDomain(service, label, address)
						domain.ID = fmt.Sprintf("%s:%d", serviceName, i)
						domains = append(domains, domain)
					}
				} else if self.ip != "" {
					address := fmt.Sprintf("%s:%d", self.ip, usePort.PublishedPort)
					domains = append(domains, self.makeDomain(service, label, address))
				} else {
					address := self.getVirtualAddress(service, usePort.TargetPort)
					domains = append(domains, self.makeDomain(service, label, address))
				}
			} else {
				address := self.getVirtualAddress(service, usePort.TargetPort)
				domains = append(domains, self.makeDomain(service, label, address))
			}
		}
	}
	return domains, nil
}
