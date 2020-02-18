package docker

import (
	dockerApi "github.com/fsouza/go-dockerclient"
	"github.com/ihaiker/aginx/registor"
	"os"
)

type DockerRegistrator struct {
	docker *dockerApi.Client

	events chan registor.ServerEvent

	closeC chan struct{}

	servers map[string] /*domain*/ map[string] /*container id*/ registor.Server

	ip string
}

func Registrator(ip string) (*DockerRegistrator, error) {
	docker, err := dockerApi.NewClientFromEnv()
	if err != nil {
		return nil, err
	}
	return &DockerRegistrator{
		docker: docker, events: make(chan registor.ServerEvent), ip: ip,
		closeC: make(chan struct{}), servers: map[string]map[string]registor.Server{},
	}, nil
}

func (self *DockerRegistrator) Sync() (registor.Servers, error) {

	containers, err := self.docker.ListContainers(dockerApi.ListContainersOptions{
		All: true, Filters: map[string][]string{"status": {"running"}},
	})
	if err != nil {
		return nil, err
	}

	domains := make([]registor.Server, 0)
	for _, container := range containers {
		if ds, err := getServer(self.ip, self.docker, container.ID); err == nil {
			self.appendServers(ds)
			domains = append(domains, ds...)
		} else if err == os.ErrNotExist {
			logger.Debug("ignore id:", container.ID[:12], ", name:", container.Names)
		} else {
			logger.Warn("ignore id:", container.ID[:12], ", name:", container.Names, ", label error:", err)
		}
	}
	return domains, nil
}

func (self *DockerRegistrator) appendServers(domains registor.Servers) {
	for domain, servers := range domains.Group() {
		for _, server := range servers {
			if _, has := self.servers[domain]; has {
				self.servers[domain][server.ID()] = server
			} else {
				self.servers[domain] = map[string]registor.Server{server.ID(): server}
			}
		}
	}
}

func (self *DockerRegistrator) Start() error {
	events := make(chan *dockerApi.APIEvents)
	if err := self.docker.AddEventListener(events); err != nil {
		return err
	}
	go func() {
		for {
			select {
			case <-self.closeC:
				return
			case event := <-events:
				if event.Type == "container" {
					if event.Status == "start" {
						if domains, err := getServer(self.ip, self.docker, event.ID); err == nil {
							self.appendServers(domains)
							self.events <- registor.ServerEvent{
								EventType: registor.Online, Servers: domains,
							}
						} else {
							logger.WithError(err).Warn("get inspect ignore ", event.ID)
						}
					} else if event.Status == "die" {
						if labs := findLabels(event.Actor.Attributes); labs.Has() {
							for _, label := range labs {
								domain := label.Domain
								if serverMap, has := self.servers[domain]; has {
									if server, has := serverMap[event.ID]; has {
										delete(self.servers[domain], event.ID)
										self.events <- registor.ServerEvent{
											EventType: registor.Offline, Servers: registor.Servers{server},
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
			}
		}
	}()
	return nil
}

func (self *DockerRegistrator) Stop() error {
	close(self.closeC)
	return nil
}

func (self *DockerRegistrator) Listener() <-chan registor.ServerEvent {
	return self.events
}
