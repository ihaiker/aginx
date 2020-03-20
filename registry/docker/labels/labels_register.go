package dockerLabels

import (
	"context"
	"errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	dockerClient "github.com/docker/docker/client"
	"github.com/ihaiker/aginx/logs"
	"github.com/ihaiker/aginx/plugins"
	"github.com/ihaiker/aginx/util"
	"strings"
	"text/template"
)

var logger = logs.New("register", "engine", "docker.labels")
var ErrExplicitlyPort = errors.New("Port not explicitly specified")

type DockerLabelsRegister struct {
	docker *dockerClient.Client

	events chan interface{}

	closeC chan struct{}

	servers map[string] /*domain*/ map[string] /*container id or service name[1..replaced]*/ plugins.Domain

	ip string
}

func (self *DockerLabelsRegister) TemplateFuncMap() template.FuncMap {
	return template.FuncMap{}
}

func (self *DockerLabelsRegister) Support() plugins.RegistrySupport {
	return plugins.RegistrySupportLabel
}

func LabelsRegister(ip string) (*DockerLabelsRegister, error) {
	docker, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv)
	if err != nil {
		return nil, err
	}
	return &DockerLabelsRegister{
		docker: docker, events: make(chan interface{}, 10), ip: ip,
		closeC: make(chan struct{}), servers: map[string]map[string]plugins.Domain{},
	}, nil
}

func (self *DockerLabelsRegister) listService() (domains []plugins.Domain) {
	defer util.Catch(func(err error) {
		logger.Warn("list swarm worker error ", err)
	})

	domains = make([]plugins.Domain, 0)
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
					}
				}
			}
		}
	}
	return
}

func (self *DockerLabelsRegister) allDomains() plugins.Domains {
	logger.Debug("Search all containers and services")
	domains := self.listService()

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
			}
		}
	}

	return domains
}

func (self *DockerLabelsRegister) Get(domain string) plugins.Domains {
	services := plugins.Domains{}
	if ss, has := self.servers[domain]; has {
		for _, server := range ss {
			services = append(services, server)
		}
	}
	return services
}

func (self *DockerLabelsRegister) appendDomains(domains plugins.Domains) {
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

func (self *DockerLabelsRegister) serviceEvent(event events.Message) {
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
				self.events <- domains.Group()
			}
		}
	case "remove":
		{
			labelsEvents := plugins.LabelsRegistryEvent(map[string]plugins.Domains{})
			for domain, servicesMap := range self.servers {
				for id, _ := range servicesMap {
					if id == serviceName || strings.HasPrefix(id, serviceName+":") {
						logger.Info("remove domain ", domain)
						labelsEvents[domain] = plugins.Domains{}
						delete(self.servers, domain)
						break
					}
				}
			}
			self.events <- labelsEvents
		}
	}
}

func (self *DockerLabelsRegister) containerEvent(event events.Message) {
	//containerName := event.Actor.Attributes["name"]
	if event.Status == "start" {
		if domains, err := self.findFromContainer(event.ID); err == nil && len(domains) > 0 {
			self.appendDomains(domains)

			labelsEvents := plugins.LabelsRegistryEvent(map[string]plugins.Domains{})
			for domain, _ := range domains.Group() {
				labelsEvents[domain] = self.Get(domain)
			}
			self.events <- labelsEvents
		}
	} else if event.Status == "die" {
		if labs := FindLabels(event.Actor.Attributes, true); labs.Has() {
			labelsEvents := plugins.LabelsRegistryEvent(map[string]plugins.Domains{})
			for _, label := range labs {
				domain := label.Domain
				if serverMap, has := self.servers[domain]; has {
					if _, has := serverMap[event.ID]; has {
						delete(self.servers[domain], event.ID)
					}
				}
				labelsEvents[domain] = self.Get(domain)
				if len(self.servers[domain]) == 0 {
					delete(self.servers, domain)
				}
			}
			self.events <- labelsEvents
		}
	}
}

func (self *DockerLabelsRegister) Start() error {
	logger.Info("start DOCKER registry")

	self.events <- self.allDomains().Group()

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

func (self *DockerLabelsRegister) Stop() error {
	close(self.closeC)
	return nil
}

func (self *DockerLabelsRegister) Listener() <-chan interface{} {
	return self.events
}
