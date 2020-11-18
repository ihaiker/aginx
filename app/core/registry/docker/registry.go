package docker

import (
	"github.com/ihaiker/aginx/v2/api"
	"github.com/ihaiker/aginx/v2/core/logs"
	"github.com/ihaiker/aginx/v2/core/registry/functions"
	"github.com/ihaiker/aginx/v2/plugins/registry"
	"net/url"
	"text/template"
)

var logger = logs.New("register", "engine", "docker")

type dockerRegistry struct {
	eventsChan chan registry.LabelsEvent
	closeChan  chan struct{}
	watchers   []*dockerWatcher
}

func LoadRegistry() *dockerRegistry {
	return &dockerRegistry{
		eventsChan: make(chan registry.LabelsEvent),
		closeChan:  make(chan struct{}),
		watchers:   make([]*dockerWatcher, 0),
	}
}

func (d *dockerRegistry) Scheme() string {
	return "docker"
}

func (d *dockerRegistry) Name() string {
	return "docker 服务发现"
}

func (d *dockerRegistry) Version() string {
	return "v2.0.0"
}

func (d *dockerRegistry) Help() string {
	return "docker registry"
}

func (d *dockerRegistry) Watch(config url.URL, aginx api.Aginx) error {
	watcher, err := newWatcher(d.closeChan, d.eventsChan, config, aginx)
	if err == nil {
		d.watchers = append(d.watchers, watcher)
	}
	return err
}

func (d *dockerRegistry) Start() error {
	for _, watcher := range d.watchers {
		if err := watcher.Start(); err != nil {
			return err
		}
	}
	return nil
}

func (d *dockerRegistry) Stop() error {
	close(d.closeChan)
	return nil
}

func (d *dockerRegistry) Label() <-chan registry.LabelsEvent {
	return d.eventsChan
}

func (c *dockerRegistry) TemplateFuncMap() template.FuncMap {
	return functions.TemplateFuncs()
}
