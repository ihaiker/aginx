package docker

import (
	"context"
	"errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	dockerClient "github.com/docker/docker/client"
	"github.com/ihaiker/aginx/plugins"
	"strings"
)

var ErrExplicitlyPort = errors.New("Port not explicitly specified")

type DockerRegistor struct {
	docker *dockerClient.Client

	events chan plugins.RegistryDomainEvent

	closeC chan struct{}

	servers map[string] /*domain*/ map[string] /*container id or service name[1..replaced]*/ plugins.Domain

	ip string
}

func Register(ip string) (*DockerRegistor, error) {
	docker, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv)
	if err != nil {
		return nil, err
	}
	return &DockerRegistor{
		docker: docker, events: make(chan plugins.RegistryDomainEvent), ip: ip,
		closeC: make(chan struct{}), servers: map[string]map[string]plugins.Domain{},
	}, nil
}

func (self *DockerRegistor) Sync() plugins.Domains {
	logger.Debug("Search all containers and services")
	domains := make([]plugins.Domain, 0)

	if info, err := self.docker.Info(context.TODO()); err != nil {
		logger.Warn("docker info error ", err)

	} else if info.Swarm.NodeID != "" {
		if !info.Swarm.ControlAvailable {
			logger.Debug("docker is swarm worker, ignore list services")
		} else {
			logger.Debug("docker swarm manager, list services")
			if services, err := self.docker.ServiceList(context.TODO(), types.ServiceListOptions{}); err != nil {
				logger.Warn("docker list services error: ", err)
			} else {
				for _, service := range services {
					if ds, err := self.findFromService(service); err == nil && len(ds) > 0 {
						logger.Info("found service ", service.Spec.Name, " domains: ", strings.Join(ds.GetDomains(), ","))
						self.appendDomains(ds)
						domains = append(domains, ds...)
					} else {
						logger.WithError(err).Warn("ignore service ", service.Spec.Name)
					}
				}
			}
		}
	}

	if containers, err := self.docker.ContainerList(context.TODO(), types.ContainerListOptions{
		All: true, Filters: filters.NewArgs(filters.Arg("status", "running")),
	}); err != nil {
		logger.Warn("list container error:", err)
	} else {
		for _, container := range containers {
			if ds, err := self.findFromContainer(container.ID); err == nil && len(ds) > 0 {
				logger.Info("found container ", strings.Join(container.Names, ","), " domains: ", strings.Join(ds.GetDomains(), ","))
				self.appendDomains(ds)
				domains = append(domains, ds...)
			} else {
				logger.WithError(err).Warn("ignore container ", container.Names)
			}
		}
	}
	return domains
}

func (self *DockerRegistor) Get(domain string) plugins.Domains {
	services := plugins.Domains{}
	if ss, has := self.servers[domain]; has {
		for _, server := range ss {
			services = append(services, server)
		}
	}
	return services
}

func (self *DockerRegistor) appendDomains(domains plugins.Domains) {
	for domain, servers := range domains.Group() {
		for _, server := range servers {
			if _, has := self.servers[domain]; has {
				self.servers[domain][server.ID] = server
			} else {
				self.servers[domain] = map[string]plugins.Domain{server.ID: server}
			}
		}
	}
}

func (self *DockerRegistor) serviceEvent(event events.Message) {
	serviceName := event.Actor.Attributes["name"]
	switch event.Action {
	case /*"create",*/ "update":
		{
			if domains, err := self.findFromServiceById(serviceName); err == nil && len(domains) > 0 {
				//clear domains
				for domain, _ := range domains.Group() {
					delete(self.servers, domain)
				}
				self.appendDomains(domains)
				self.events <- plugins.RegistryDomainEvent{
					EventType: plugins.Online, Servers: domains,
				}
			} else {
				logger.Debug("ignore service ", serviceName)
			}
		}
	case "remove":
		{
			for domain, servicesMap := range self.servers {
				for id, _ := range servicesMap {
					if id == serviceName || strings.HasPrefix(id, serviceName+":") {
						logger.Info("remove domain ", domain)
						services := plugins.Domains{}
						for _, s := range servicesMap {
							services = append(services, s)
						}
						delete(self.servers, domain)
						self.events <- plugins.RegistryDomainEvent{
							EventType: plugins.Offline, Servers: services,
						}
						break
					}
				}
			}
		}
	}
}

func (self *DockerRegistor) containerEvent(event events.Message) {
	containerName := event.Actor.Attributes["name"]
	if event.Status == "start" {
		if domains, err := self.findFromContainer(event.ID); err == nil && len(domains) > 0 {
			self.appendDomains(domains)
			self.events <- plugins.RegistryDomainEvent{
				EventType: plugins.Online, Servers: domains,
			}
		} else {
			logger.WithError(err).Warn("ignore container: ", containerName)
		}
	} else if event.Status == "die" {
		if labs := findLabels(event.Actor.Attributes, true); labs.Has() {
			for _, label := range labs {
				domain := label.Domain
				if serverMap, has := self.servers[domain]; has {
					if server, has := serverMap[event.ID]; has {
						delete(self.servers[domain], event.ID)
						self.events <- plugins.RegistryDomainEvent{
							EventType: plugins.Offline, Servers: plugins.Domains{server},
						}
					}
				}
				if len(self.servers[domain]) == 0 {
					delete(self.servers, domain)
				}
			}
		}
	}
}

func (self *DockerRegistor) Start() error {
	logger.Info("start DOCKER registry")
	eventChannel, errChannel := self.docker.Events(context.TODO(), types.EventsOptions{})
	go func() {
		for {
			select {
			case <-self.closeC:
				close(self.events)
				return
			case err, has := <-errChannel:
				if has {
					logger.Warn("DOCKER event error ", err)
				}
			case event, has := <-eventChannel:
				if !has {
					continue
				}
				if event.Type == "service" {
					self.serviceEvent(event)
				} else if event.Type == "container" {
					self.containerEvent(event)
				}
			}
		}
	}()
	return nil
}

func (self *DockerRegistor) Stop() error {
	close(self.closeC)
	return nil
}

func (self *DockerRegistor) Listener() <-chan plugins.RegistryDomainEvent {
	return self.events
}
