package consul

import (
	"github.com/ihaiker/aginx/v2/api"
	"github.com/ihaiker/aginx/v2/core/logs"
	"github.com/ihaiker/aginx/v2/core/registry/functions"
	"github.com/ihaiker/aginx/v2/plugins/registry"
	"net/url"
	"text/template"
)

var logger = logs.New("registry", "engine", "consul")

type consulRegistry struct {
	watchers []*consulWatcher
	events   chan registry.LabelsEvent
}

func LoadRegistry() *consulRegistry {
	return &consulRegistry{
		events:   make(chan registry.LabelsEvent),
		watchers: make([]*consulWatcher, 0),
	}
}
func (c *consulRegistry) Scheme() string {
	return "consul"
}

func (c *consulRegistry) Name() string {
	return "Consul 注册中心"
}

func (c *consulRegistry) Version() string {
	return "v2.0.0"
}

func (c *consulRegistry) Help() string {
	return "consul registry"
}

func (c *consulRegistry) Watch(config url.URL, aginx api.Aginx) error {
	watcher, err := newWatch(c.events, config, aginx)
	if err == nil {
		c.watchers = append(c.watchers, watcher)
	}
	return err
}

func (c *consulRegistry) Label() <-chan registry.LabelsEvent {
	return c.events
}

func (c *consulRegistry) TemplateFuncMap() template.FuncMap {
	return functions.TemplateFuncs()
}

func (c *consulRegistry) Start() error {
	for _, watcher := range c.watchers {
		if err := watcher.Start(); err != nil {
			return err
		}
	}
	return nil
}

func (c *consulRegistry) Stop() error {
	close(c.events)
	return nil
}
