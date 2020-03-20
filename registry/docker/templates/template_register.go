package dockerTemplates

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	dockerClient "github.com/docker/docker/client"
	"github.com/ihaiker/aginx/logs"
	"github.com/ihaiker/aginx/plugins"
	"github.com/ihaiker/aginx/util"
	"regexp"
	"strings"
	"text/template"
)

var logger = logs.New("register", "engine", "docker.template")

type DockerTemplateRegister struct {
	events                           chan interface{}
	docker                           *dockerClient.Client
	closeC                           chan struct{}
	ip                               string
	filterServices, filterContainers []string
}

func TemplateRegister(publishIp string, filterServices, filterContainers []string) (register *DockerTemplateRegister, err error) {
	register = &DockerTemplateRegister{
		events:           make(chan interface{}, 10),
		closeC:           make(chan struct{}),
		ip:               publishIp,
		filterServices:   filterServices,
		filterContainers: filterContainers,
	}
	if register.docker, err = dockerClient.NewClientWithOpts(dockerClient.FromEnv); err == nil {

	}
	return
}

func (self *DockerTemplateRegister) listServices() []swarm.Service {
	defer util.Catch(func(err error) {
		logger.Warn("list swarm worker error ", err)
	})
	//service
	if info, err := self.docker.Info(context.TODO()); err != nil {
		logger.Warn("docker info error ", err)
	} else if info.Swarm.NodeID != "" {
		if info.Swarm.ControlAvailable {
			if services, err := self.docker.ServiceList(context.TODO(), types.ServiceListOptions{}); err != nil {
				logger.Warn("docker list services error: ", err)
			} else {
				return services
			}
		}
	}
	return make([]swarm.Service, 0)
}

func (self *DockerTemplateRegister) allInfo() {
	data := new(DockerTemplateRegisterEvents)
	data.PublishIP = self.ip
	data.Docker = self.docker
	data.Services = self.listServices()

	//container
	if containers, err := self.docker.ContainerList(context.TODO(), types.ContainerListOptions{
		All: true, Filters: filters.NewArgs(filters.Arg("status", "running")),
	}); err != nil {
		logger.Warn("list container error:", err)
	} else {
		for _, containerSummary := range containers {
			if containerInspect, err := self.docker.ContainerInspect(context.TODO(), containerSummary.ID); err == nil {
				data.Containers = append(data.Containers, containerInspect)
			} else {
				logger.Warnf("inspect container %s error %s", strings.Join(containerSummary.Names, ","), err)
			}
		}
	}

	//nodes
	if nodes, err := self.docker.NodeList(context.TODO(), types.NodeListOptions{}); err == nil {
		data.Nodes = nodes
	}

	self.events <- data
}

func (self *DockerTemplateRegister) dockerEvent() {
	eventChannel, errChannel := self.docker.Events(context.TODO(), types.EventsOptions{})
	for {
		select {
		case <-self.closeC:
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
				if event.Action == "update" || event.Action == "remove" {
					serviceName := event.Actor.Attributes["name"]
					if self.filter(self.filterContainers, serviceName) {
						self.allInfo()
					}
				}
			} else if event.Type == "container" {
				containerName := event.Actor.Attributes["name"]
				if event.Action == "start" || event.Action == "die" {
					if self.filter(self.filterContainers, containerName) {
						self.allInfo()
					}
				}
			}
		}
	}
}

func (self *DockerTemplateRegister) filter(patterns []string, name string) bool {
	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, name); matched {
			return true
		}
	}
	return false
}

func (self *DockerTemplateRegister) Start() error {
	self.allInfo()
	go self.dockerEvent()
	return nil
}

func (self *DockerTemplateRegister) Stop() error {
	close(self.closeC)
	return nil
}

func (self *DockerTemplateRegister) Support() plugins.RegistrySupport {
	return plugins.RegistrySupportTemplate
}

func (self *DockerTemplateRegister) Listener() <-chan interface{} {
	return self.events
}

func (self *DockerTemplateRegister) TemplateFuncMap() template.FuncMap {
	return templateFuncs(self.docker)
}
