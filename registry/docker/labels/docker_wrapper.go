package dockerLabels

import (
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/swarm"
	dockerClient "github.com/docker/docker/client"
	"github.com/ihaiker/aginx/util"
	"io"
	"os"
	"strings"
)

type (
	nodeInfo struct {
		id              string
		name            string
		ip              string
		manager, leader bool
		closeC          chan struct{}
		*dockerClient.Client
	}

	dockerWrapper struct {
		Events chan events.Message
		nodes  map[string]*nodeInfo
	}
)

func (node *nodeInfo) destroy() {
	util.Try(func() { close(node.closeC) })
	_ = node.Close()
}

func dockerHost(managerHost, addr string) string {
	idx := strings.LastIndex(managerHost, ":")
	return fmt.Sprintf("tcp://%s:%s", addr, managerHost[idx+1:])
}

func (self *dockerWrapper) swarmNodes(client *dockerClient.Client) error {
	host := os.Getenv("DOCKER_HOST")
	if nodes, err := client.NodeList(todo, types.NodeListOptions{}); err != nil {
		return util.Wrap(err, "node list")
	} else {
		for _, node := range nodes {
			if node.Status.State == swarm.NodeStateReady {

				if nodeInfo, has := self.nodes[node.ID]; has {
					nodeInfo.manager = node.ManagerStatus != nil
					nodeInfo.leader = node.ManagerStatus != nil && node.ManagerStatus.Leader
					continue
				}

				docker, err := dockerClient.NewClientWithOpts(
					dockerClient.FromEnv,
					dockerClient.WithHost(dockerHost(host, node.Status.Addr)),
				)
				if err != nil {
					return util.Wrap(err, fmt.Sprintf("init node client %s", node.Status.Addr))
				}
				ni := &nodeInfo{
					id: node.ID, name: node.Description.Hostname, ip: node.Status.Addr,
					manager: node.ManagerStatus != nil,
					leader:  node.ManagerStatus != nil && node.ManagerStatus.Leader,
					Client:  docker, closeC: make(chan struct{}),
				}
				logger.Infof("docker node [%s,%s] is %s ",
					node.Description.Hostname, node.Status.Addr, node.Status.State)
				go self.event(ni)
				self.nodes[node.ID] = ni

			} else {
				logger.Infof("docker node [%s,%s] is %s ignore",
					node.Description.Hostname, node.Status.Addr, node.Status.State)
			}
		}
		return nil
	}
}

func NewDockerWrapper(ip string, swarm bool) (*dockerWrapper, error) {
	if swarm {
		host := os.Getenv("DOCKER_HOST")
		if !strings.HasPrefix(host, "tcp://") {
			return nil, errors.New("must use --docker-host=tcp://managerIp:port")
		}
	}

	client, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv)
	if err != nil {
		return nil, err
	}
	info, err := client.Info(todo)
	if err != nil {
		return nil, err
	}
	w := &dockerWrapper{
		nodes:  make(map[string]*nodeInfo, 0),
		Events: make(chan events.Message, 20),
	}
	if swarm {
		if info.Swarm.NodeID == "" || !info.Swarm.ControlAvailable {
			return nil, errors.New("docker not swarm mode or node not manger")
		}
		err = w.swarmNodes(client)
	} else {
		w.nodes[info.ID] = &nodeInfo{
			id: info.ID, name: info.Name, ip: ip,
			Client: client, closeC: make(chan struct{}),
		}
	}
	return w, err
}

func (self *dockerWrapper) event(node *nodeInfo) {
	eventChannel, errChannel := node.Events(todo, types.EventsOptions{})
	for {
		select {
		case <-node.closeC:
			return
		case err, has := <-errChannel:
			if has && err != io.EOF {
				logger.Warn("docker event error ", err)
				eventChannel, errChannel = node.Events(todo, types.EventsOptions{})
			}
		case event, has := <-eventChannel:
			if !has {
				continue
			}
			switch event.Type {
			case events.ServiceEventType:
				if !node.leader {
					continue
				}
			case events.NodeEventType:
				if event.Action == "update" {
					id := event.Actor.ID
					state := event.Actor.Attributes["state.new"]
					if state == "down" {
						if node, has := self.nodes[id]; has {
							delete(self.nodes, id)
							node.destroy()
						}
					} else {
						if err := self.swarmNodes(node.Client); err != nil {
							logger.Warn("list swarm node error ", err)
						}
					}
				}
				continue
			}

			if event.Type == events.ServiceEventType || event.Type == events.ContainerEventType {
				self.Events <- event
			}
		}
	}
}

func (self *dockerWrapper) Stop() {
	for _, node := range self.nodes {
		node.destroy()
	}
}

func (self *dockerWrapper) ServiceList(options types.ServiceListOptions) (ss []swarm.Service, err error) {
	for _, node := range self.nodes {
		if node.manager && node.leader {
			if ss, err = node.ServiceList(todo, options); err == nil {
				return
			}
		}
	}
	return
}

func (self *dockerWrapper) ContainerList(options types.ContainerListOptions) ([]types.Container, error) {
	containers := make([]types.Container, 0)
	for _, node := range self.nodes {
		if cs, err := node.ContainerList(todo, options); err == nil {
			containers = append(containers, cs...)
		}
	}
	return containers, nil
}

func (self *dockerWrapper) ContainerInspect(containerID string) (container types.ContainerJSON, info *nodeInfo, err error) {
	for _, node := range self.nodes {
		if container, err = node.ContainerInspect(todo, containerID); err == nil {
			info = node
			return
		}
	}
	return
}

func (self *dockerWrapper) ServiceInspectWithRaw(serviceId string, opts types.ServiceInspectOptions) (service swarm.Service, bs []byte, err error) {
	for _, node := range self.nodes {
		if node.manager {
			if service, bs, err = node.ServiceInspectWithRaw(todo, serviceId, opts); err == nil {
				return
			}
		}
	}
	err = os.ErrNotExist
	return
}

func (self *dockerWrapper) Nodes() []string {
	nodeIps := make([]string, 0)
	for _, info := range self.nodes {
		nodeIps = append(nodeIps, info.ip)
	}
	return nodeIps
}

func (self *dockerWrapper) ImageInspectWithRaw(imageID string) (types.ImageInspect, []byte, error) {
	for _, node := range self.nodes {
		if image, bs, err := node.ImageInspectWithRaw(todo, imageID); err == nil {
			return image, bs, nil
		}
	}
	return types.ImageInspect{}, nil, os.ErrNotExist
}

func (self *dockerWrapper) TaskList(options types.TaskListOptions) ([]swarm.Task, error) {
	for _, node := range self.nodes {
		if node.manager {
			if tasks, err := node.TaskList(todo, options); err == nil {
				return tasks, err
			}
		}
	}
	return nil, nil
}
