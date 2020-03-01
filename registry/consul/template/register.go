package consulTemplate

import (
	consulApi "github.com/hashicorp/consul/api"
	"github.com/ihaiker/aginx/logs"
	"github.com/ihaiker/aginx/plugins"
	"github.com/ihaiker/aginx/util"
	"text/template"
	"time"
)

var logger = logs.New("register", "engine", "consul.template")

type ConsulTemplateRegister struct {
	consul  *consulApi.Client
	events  chan interface{}
	lastIdx uint64
}

func NewTemplateRegister(consul *consulApi.Client) *ConsulTemplateRegister {
	return &ConsulTemplateRegister{
		consul: consul,
		events: make(chan interface{}),
	}
}

func (self *ConsulTemplateRegister) allServices() {
	defer util.Catch(func(err error) {
		logger.Warn("search all service ", err)
	})

	services, meta, err := self.consul.Catalog().Services(&consulApi.QueryOptions{
		WaitIndex: self.lastIdx, WaitTime: time.Second * 3, RequireConsistent: true,
	})
	if err != nil {
		logger.Warn("list services ", err)
		return
	}

	event := &ConsulTemplateEvent{Services: map[string][]*consulApi.ServiceEntry{}, Consul: self.consul}
	for serviceName, _ := range services {
		serviceEntries, _, _ := self.consul.Health().Service(serviceName, "", true, nil)
		event.Services[serviceName] = serviceEntries
	}

	self.events <- event
	self.lastIdx = meta.LastIndex
}

func (self *ConsulTemplateRegister) Start() error {
	go func() {
		for {
			self.allServices()
		}
	}()
	return nil
}

func (self *ConsulTemplateRegister) Stop() error {
	close(self.events)
	return nil
}

func (self *ConsulTemplateRegister) Support() plugins.RegistrySupport {
	return plugins.RegistrySupportTemplate
}

func (self *ConsulTemplateRegister) Listener() <-chan interface{} {
	return self.events
}

func (self *ConsulTemplateRegister) TemplateFuncMap() template.FuncMap {
	return templateFuncs(self.consul)
}
