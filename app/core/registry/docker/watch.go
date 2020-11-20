package docker

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/ihaiker/aginx/v2/api"
	"github.com/ihaiker/aginx/v2/core/registry/addition"
	"github.com/ihaiker/aginx/v2/core/util/errors"
	"github.com/ihaiker/aginx/v2/plugins/registry"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"
)

type dockerWatcher struct {
	aginx  api.Aginx
	docker *client.Client

	ip     string //docker 地址
	events map[string] /*domain*/ map[string] /*id*/ registry.Domain

	//当前注册管理定义一个名字，默认为：""。
	//这个名字用户在外部配置name label时使用，防止在两个不同注册中心存在相同的名字，概率还是很高的
	name string

	eventsChan chan registry.LabelsEvent
	closeChan  chan struct{}
}

func newWatcher(closeChan chan struct{}, eventsChan chan registry.LabelsEvent, config url.URL, aginx api.Aginx) (*dockerWatcher, error) {
	docker, ip, err := fromClient(config)
	if err != nil {
		return nil, err
	}
	name := config.Query().Get("name")
	reg := &dockerWatcher{
		name: name, aginx: aginx, docker: docker, ip: ip,
		closeChan: closeChan, eventsChan: eventsChan,
		events: map[string]map[string]registry.Domain{},
	}
	return reg, nil
}

func (d *dockerWatcher) Start() error {
	event, err := d.fetchAll()
	if err != nil {
		return err
	}
	d.cacheEvent(event)
	//第一次发送全部事件
	if len(event) > 0 {
		d.eventsChan <- event
	}
	go d.watch()
	return nil
}

func (c *dockerWatcher) findLabel(finderLabel addition.LabelFinder,
	containerLabels map[string]string, containerName string) ([]label, error) {
	labels, err := findLabels(containerLabels)
	if err == nil && (labels == nil || len(labels) == 0) {
		externalName := containerName
		if c.name != "" {
			externalName = c.name + "." + containerName
		}
		return findLabels(finderLabel(externalName, containerLabels))
	}
	return labels, err
}

//获取有所有services
func (d *dockerWatcher) services() (registry.LabelsEvent, error) {
	logger.Debug("查询docker服务")
	labelsEvent := registry.LabelsEvent{}

	info, err := d.docker.Info(context.TODO())
	if err != nil {
		logger.WithError(err).Warn("获取info错误")
		return nil, errors.Wrap(err, "获取info错误")
	}
	if info.Swarm.NodeID == "" || !info.Swarm.ControlAvailable {
		//不是swarm模式，或者不是主控节点
		return labelsEvent, nil
	}

	services, err := d.docker.ServiceList(context.TODO(), types.ServiceListOptions{})
	if err != nil {
		return nil, err
	}
	for _, service := range services {
		event, err := d.findInService(service.ID, service.Spec.Name)
		if err != nil {
			return nil, err
		}
		labelsEvent = append(labelsEvent, event...)
	}
	return labelsEvent, nil
}

func (d *dockerWatcher) containers() (registry.LabelsEvent, error) {
	logger.Debug("查询所有容器")
	labelsEvent := registry.LabelsEvent{}
	containers, err := d.docker.ContainerList(context.TODO(), types.ContainerListOptions{
		All: true, Filters: filters.NewArgs(filters.Arg("status", "running")),
	})
	if err != nil {
		return nil, err
	}
	for _, container := range containers {
		event, err := d.findInContainer(container.ID, container.Names[0])
		if err != nil {
			return nil, err
		}
		labelsEvent = append(labelsEvent, event...)
	}
	return labelsEvent, nil
}

func (c *dockerWatcher) labelFinder() addition.LabelFinder {
	return addition.Load(c.aginx, "registry/docker-labels.conf")
}

func (d *dockerWatcher) findInService(serviceId, serviceName string) (registry.LabelsEvent, error) {
	service, _, err := d.docker.ServiceInspectWithRaw(context.TODO(), serviceId, types.ServiceInspectOptions{InsertDefaults: true})
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("获取service信息: %s", serviceName))
	}

	labels, err := d.findLabel(d.labelFinder(), service.Spec.TaskTemplate.ContainerSpec.Labels, service.Spec.Name)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("获取标签: %s", serviceName))
	}
	labelEvents := registry.LabelsEvent{}
	for _, label := range labels {
		event := registry.Domain{
			ID:     service.ID,
			Domain: label.Domain, Address: []string{},
			Weight: label.Weight, AutoSSL: label.AutoSSL,
			Alive: true, Template: label.Template,
			Provider: label.Provider,
		}
		if label.Port == 0 && len(service.Endpoint.Ports) != 1 {
			return nil, fmt.Errorf("未能找到正确的代理端口：service %s", serviceName)
		}

		//确定端口
		targetPort, publishPort := 0, 0
		if label.Port == 0 { //未指定但是公开一个端口
			targetPort = int(service.Endpoint.Ports[0].TargetPort)
			publishPort = int(service.Endpoint.Ports[0].PublishedPort)
		} else if len(service.Endpoint.Ports) == 0 { //已指定但是未公开端口
			targetPort = label.Port
		} else { //已指定但是公开多个端口
			for _, port := range service.Endpoint.Ports {
				if int(port.TargetPort) == label.Port || int(port.PublishedPort) == label.Port {
					targetPort = int(port.TargetPort)
					publishPort = int(port.PublishedPort)
					break
				}
			}
		}

		if label.Internal || publishPort == 0 || (label.Port == targetPort && targetPort != publishPort) {
			//确定网络地址
			vip := service.Endpoint.VirtualIPs[len(service.Endpoint.VirtualIPs)-1].Addr
			if label.Networks != "" {
				for _, virtualIP := range service.Endpoint.VirtualIPs {
					if strings.HasPrefix(virtualIP.Addr, label.Networks) {
						vip = virtualIP.Addr
						break
					} else if nw, err := d.docker.NetworkInspect(
						context.TODO(), virtualIP.NetworkID, types.NetworkInspectOptions{},
					); err == nil {
						if nw.Name == label.Networks {
							vip = virtualIP.Addr
						}
					}
				}
			}
			ip, _, _ := net.ParseCIDR(vip)
			vip = ip.String()
			event.Address = []string{fmt.Sprintf("%s:%d", vip, targetPort)}

		} else {
			nodes, err := d.docker.NodeList(context.TODO(), types.NodeListOptions{})
			if err != nil {
				return nil, errors.Wrap(err, "node list")
			}
			for _, node := range nodes {
				event.Address = append(event.Address, fmt.Sprintf("%s:%d", node.Status.Addr, publishPort))
			}
		}
		labelEvents = append(labelEvents, event)
	}
	return labelEvents, nil
}

func (d *dockerWatcher) findInContainer(containerId, containerName string) (registry.LabelsEvent, error) {
	container, err := d.docker.ContainerInspect(context.TODO(), containerId)
	if err != nil {
		logger.WithError(err).Warnf("获取容器信息错误: %s", containerName)
		return nil, err
	}

	labels, err := d.findLabel(d.labelFinder(), container.Config.Labels, containerName)
	if err != nil {
		logger.WithError(err).Warnf("search container(%s) labels", containerName)
		return nil, err
	}
	labelsEvent := registry.LabelsEvent{}

	for _, label := range labels {
		event := registry.Domain{
			ID: containerId, Alive: true, Template: label.Template,
			Domain: label.Domain, Address: []string{},
			Weight: label.Weight, AutoSSL: label.AutoSSL,
			Provider: label.Provider,
		}

		targetPort, publishPort := 0, 0
		if label.Port == 0 {
			if len(container.HostConfig.PortBindings) == 0 {
				logger.Debug("未指定端口并且未公开端口，查找ExposedPorts端口: ", container.Name, label.Source)
				if len(container.Config.ExposedPorts) != 1 {
					return nil, fmt.Errorf("端口不可确定：%s %s", container.Name, label.Source)
				} else {
					for port, _ := range container.Config.ExposedPorts {
						targetPort = port.Int()
					}
				}
			} else if len(container.HostConfig.PortBindings) == 1 { //从公开端口获取
				for port, bindings := range container.HostConfig.PortBindings {
					targetPort = port.Int()
					publishPort, _ = nat.ParsePort(bindings[0].HostPort)
				}
			} else { //开放多个端口不可确定
				return nil, fmt.Errorf("端口不可确定：%s %s", container.Name, label.Source)
			}
		} else if container.HostConfig.NetworkMode == "host" {
			//端口一致
			targetPort, publishPort = label.Port, label.Port
		} else {
			//从指定的端口中获取当前端口和公开端口
			for port, portBinds := range container.HostConfig.PortBindings {
				if label.Port == port.Int() { //没有指定或者和内部端口一样
					targetPort = port.Int()
					publishPort, _ = nat.ParsePort(portBinds[0].HostPort)
				} else {
					for _, bind := range portBinds {
						if strconv.Itoa(label.Port) == bind.HostPort {
							targetPort = port.Int()
							publishPort, _ = nat.ParsePort(bind.HostPort)
						}
					}
				}
			}
			//指定的端口并未开放，指定的端口一定是内部端口
			if targetPort == 0 {
				targetPort = label.Port
			}
		}

		//获取容器的地址
		vip := container.NetworkSettings.DefaultNetworkSettings.IPAddress
		if container.HostConfig.NetworkMode.IsHost() {
			vip = d.ip
		} else if container.HostConfig.NetworkMode.IsUserDefined() {
			vip = container.NetworkSettings.Networks[container.HostConfig.NetworkMode.NetworkName()].IPAddress
			if label.Networks != "" { //选择网络
				for _, settings := range container.NetworkSettings.Networks {
					if strings.HasPrefix(settings.IPAddress, label.Networks) {
						vip = settings.IPAddress
						break
					} else if nw, err := d.docker.NetworkInspect(context.TODO(),
						settings.NetworkID, types.NetworkInspectOptions{}); err == nil {
						if nw.Name == label.Networks {
							vip = settings.IPAddress
							break
						}
					}
				}
			}
		}

		//确定使用内部地址，或者找不到外部地址，或者指定的端口和内部端口一致（内外不一致）
		if label.Internal || publishPort == 0 || (label.Port == targetPort && targetPort != publishPort) {
			event.Address = []string{fmt.Sprintf("%s:%d", vip, targetPort)}
		} else { //使用node ip
			event.Address = []string{fmt.Sprintf("%s:%d", d.ip, publishPort)}
		}

		labelsEvent = append(labelsEvent, event)
	}
	return labelsEvent, nil
}

//第一次启动刷新所有
func (d *dockerWatcher) fetchAll() (registry.LabelsEvent, error) {
	labelsEvent, err := d.services()
	if err != nil {
		labelsEvent = registry.LabelsEvent{}
		logger.WithError(err).Warn("查询服务错误")
	}

	containerEvents, err := d.containers()
	if err != nil {
		logger.WithError(err).Warn("查询容器错误")
	}

	labelsEvent = append(labelsEvent, containerEvents...)
	return labelsEvent, err
}

func (d *dockerWatcher) containerEvent(message events.Message) {
	containerName := message.Actor.Attributes["name"]
	event := registry.LabelsEvent{}

	switch message.Status {
	case "start":
		{
			logger.Debugf("container start: name=%s, id=%s ", containerName, message.ID)
			if labelEvent, err := d.findInContainer(message.ID, containerName); err != nil {
				logger.WithError(err).Warnf("find in container %s", containerName)
			} else {
				d.cacheEvent(labelEvent)
				event = append(event, labelEvent...)
			}
		}
	case "die":
		{
			logger.Debugf("container die: name=%s, id=%s", containerName, message.Actor.ID)
			event = d.removeEvent(message.ID)
		}
	}

	if len(event) > 0 {
		d.eventsChan <- event
	}
}

func (d *dockerWatcher) serviceEvent(message events.Message) {
	serviceName := message.Actor.Attributes["name"]
	event := registry.LabelsEvent{}

	switch message.Action {
	case "update":
		{
			if labelEvent, err := d.findInService(message.ID, serviceName); err != nil {
				logger.WithError(err).Warnf("find in service %s", serviceName)
			} else {
				d.cacheEvent(labelEvent)
				event = append(event, labelEvent...)
			}
		}
	case "remove":
		{
			logger.Debugf("remove service name=%s, id=%s", serviceName, message.Actor.ID)
			event = d.removeEvent(message.ID)
		}
	}

	if len(event) > 0 {
		d.eventsChan <- event
	}
}

//监听docker事件
func (d *dockerWatcher) watch() {
	eventChan, errChan := d.docker.Events(context.TODO(), types.EventsOptions{})
	for {
		select {
		case <-d.closeChan:
			return
		case err, has := <-errChan:
			if has && err != io.EOF {
				logger.WithError(err).Warn("docker event")
				eventChan, errChan = d.docker.Events(context.TODO(), types.EventsOptions{})
			}
		case event, has := <-eventChan:
			if !has {
				continue
			}
			if event.Type == events.ServiceEventType {
				d.serviceEvent(event)
			} else if event.Type == events.ContainerEventType {
				d.containerEvent(event)
			}
		}
	}
}

func (d *dockerWatcher) cacheEvent(events registry.LabelsEvent) {
	//保存已经注册内容
	for _, domain := range events {
		if idServices, has := d.events[domain.Domain]; has {
			idServices[domain.ID] = domain
			continue
		}
		d.events[domain.Domain] = map[string]registry.Domain{
			domain.ID: domain,
		}
	}
}

func (d *dockerWatcher) removeEvent(id string) registry.LabelsEvent {
	event := registry.LabelsEvent{}
	deleteDomains := make([]string, 0)
	for domain, idServer := range d.events {
		if e, has := idServer[id]; has {
			deleteDomains = append(deleteDomains, domain)
			e.Alive = false
			event = append(event, e)
		}
	}
	for _, domain := range deleteDomains {
		delete(d.events[domain], id)
		if len(d.events[domain]) == 0 {
			delete(d.events, domain)
		}
	}
	return event
}
