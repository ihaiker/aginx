package consulLabels

import (
	"fmt"
	consulApi "github.com/hashicorp/consul/api"
	"github.com/ihaiker/aginx/logs"
	"github.com/ihaiker/aginx/plugins"
	"github.com/ihaiker/aginx/util"
	"text/template"
	"time"
)

var logger = logs.New("register", "engine", "consul.labels")

type ConsulLabelRegister struct {
	consul   *consulApi.Client
	events   chan interface{}
	lastIdx  uint64
	services plugins.Domains
}

func (self *ConsulLabelRegister) TemplateFuncMap() template.FuncMap {
	return template.FuncMap{}
}

func NewLabelRegister(consul *consulApi.Client) *ConsulLabelRegister {
	return &ConsulLabelRegister{
		consul: consul, events: make(chan interface{}, 10),
	}
}

func (self *ConsulLabelRegister) allServices() {
	defer util.Catch(func(err error) {
		logger.Warn("search all service ", err)
	})

	services, meta, err := self.consul.Catalog().Services(&consulApi.QueryOptions{
		WaitIndex: self.lastIdx, WaitTime: time.Second * 3,
	})
	if err != nil {
		logger.Warn("list services ", err)
		return
	}

	searchServices := plugins.Domains{}
	for serviceName, _ := range services {
		if catalogServiceEntries, _, err := self.consul.Health().Service(serviceName, "", true, nil); err != nil {
			logger.Warn("error ", err)
			continue
		} else {
			for _, serviceEntry := range catalogServiceEntries {
				if labels := FindLabel(serviceEntry.Service.Meta); labels != nil && len(labels) > 0 {
					for _, label := range labels {
						if serviceEntry.Checks.AggregatedStatus() == consulApi.HealthPassing {
							weight := label.Weight
							if weight == 0 {
								weight = serviceEntry.Service.Weights.Passing
							}
							searchServices = append(searchServices, plugins.Domain{
								ID: serviceEntry.Service.ID, Domain: label.Domain,
								Weight: weight, AutoSSL: label.AutoSSL, Attrs: serviceEntry.Service.Meta,
								Address: fmt.Sprintf("%s:%d", serviceEntry.Service.Address, serviceEntry.Service.Port),
							})
						}
					}
				}
			}
		}
	}

	addDomains := make(map[string]bool)
	for _, service := range searchServices {
		if !self.find(self.services, service) {
			addDomains[service.Domain] = true
		}
	}
	removeDomains := make(map[string]bool)
	for _, service := range self.services {
		if !self.find(searchServices, service) {
			removeDomains[service.Domain] = true
		}
	}

	if len(addDomains) > 0 || len(removeDomains) > 0 {
		event := plugins.LabelsRegistryEvent{}
		groups := searchServices.Group()
		for domain, _ := range addDomains {
			event[domain] = groups[domain]
		}
		for domain, _ := range removeDomains {
			event[domain] = groups[domain]
		}
		if len(event) > 0 {
			self.events <- event
		}
	}

	self.lastIdx = meta.LastIndex
	self.services = searchServices
}

func (self *ConsulLabelRegister) find(domains plugins.Domains, search plugins.Domain) bool {
	for _, domain := range domains {
		if domain.ID == search.ID {
			return true
		}
	}
	return false
}

func (self *ConsulLabelRegister) Start() error {
	go func() {
		for {
			self.allServices()
		}
	}()
	return nil
}

func (self *ConsulLabelRegister) Stop() error {
	close(self.events)
	return nil
}

func (self *ConsulLabelRegister) Support() plugins.RegistrySupport {
	return plugins.RegistrySupportLabel
}

func (self *ConsulLabelRegister) Listener() <-chan interface{} {
	return self.events
}
