package consulTemplate

import (
	consulApi "github.com/hashicorp/consul/api"
	"github.com/ihaiker/aginx/logs"
	"github.com/ihaiker/aginx/plugins"
	"github.com/ihaiker/aginx/util"
	"regexp"
	"text/template"
	"time"
)

var logger = logs.New("register", "engine", "consul.template")

type ConsulTemplateRegister struct {
	consul        *consulApi.Client
	events        chan interface{}
	lastIdx       uint64
	cacheServices map[string][]*consulApi.ServiceEntry
	filters       []string
}

func NewTemplateRegister(consul *consulApi.Client, filters []string) *ConsulTemplateRegister {
	return &ConsulTemplateRegister{
		consul:        consul,
		events:        make(chan interface{}),
		cacheServices: map[string][]*consulApi.ServiceEntry{},
		filters:       filters,
	}
}

func (self *ConsulTemplateRegister) diff(oldServices, serviceEntries []*consulApi.ServiceEntry) bool {
	for _, entry := range serviceEntries {
		notExists := true
		for _, service := range oldServices {
			if service.Service.ID == entry.Service.ID ||
				service.Service.Address == entry.Service.Address ||
				service.Service.Port == entry.Service.Port {
				notExists = false
			}
		}
		if notExists {
			return true
		}
	}

	for _, entry := range oldServices {
		notExists := true
		for _, service := range serviceEntries {
			if service.Service.ID == entry.Service.ID ||
				service.Service.Address == entry.Service.Address ||
				service.Service.Port == entry.Service.Port {
				notExists = false
			}
		}
		if notExists {
			return true
		}
	}
	return false
}

func (self *ConsulTemplateRegister) filter(name string) bool {
	for _, filter := range self.filters {
		if matched, _ := regexp.MatchString(filter, name); matched {
			return true
		}
	}
	return false
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
		time.Sleep(time.Second * 3)
		return
	}

	changed := false
	for serverName, _ := range self.cacheServices {
		if self.filter(serverName) {
			if _, has := services[serverName]; !has {
				changed = true
			}
		}
	}
	for serviceName, _ := range services {
		if self.filter(serviceName) {
			serviceEntries, _, _ := self.consul.Health().Service(serviceName, "", true, nil)
			if changed {
				//ignore
			} else if oldServices, has := self.cacheServices[serviceName]; !has {
				changed = true
			} else if len(oldServices) != len(serviceEntries) {
				changed = true
			} else {
				changed = self.diff(oldServices, serviceEntries)
			}
			self.cacheServices[serviceName] = serviceEntries
		}
	}

	if changed {
		kvPairs, _, _ := self.consul.KV().List("", nil)
		keys := map[string]*consulApi.KVPair{}
		for _, pair := range kvPairs {
			keys[pair.Key] = pair
		}
		self.events <- &ConsulTemplateEvent{
			Consul:   self.consul,
			Services: self.cacheServices, Keys: keys,
		}
	}

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
