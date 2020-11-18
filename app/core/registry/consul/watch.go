package consul

import (
	"github.com/hashicorp/consul/api"
	aginxApi "github.com/ihaiker/aginx/v2/api"
	"github.com/ihaiker/aginx/v2/core/registry/addition"
	"github.com/ihaiker/aginx/v2/core/util"
	"github.com/ihaiker/aginx/v2/plugins/registry"
	"net"
	"net/url"
	"strconv"
	"time"
)

type consulWatcher struct {
	aginx  aginxApi.Aginx
	consul *api.Client

	lastIdx uint64                                                      //最后一次监听编号
	domains map[string] /*domain*/ map[string] /*id*/ *api.AgentService /*这个只是临时保存一下*/

	//当前注册管理定义一个名字，默认为：""。
	//这个名字用户在外部配置name label时使用，防止在两个不同注册中心存在相同的名字，概率还是很高的
	name string

	events chan registry.LabelsEvent
}

func newWatch(events chan registry.LabelsEvent, cfg url.URL, aginx aginxApi.Aginx) (*consulWatcher, error) {
	consulConfig, err := util.Consul(cfg)
	if err != nil {
		return nil, err
	}
	client, err := api.NewClient(consulConfig)
	if err != nil {
		return nil, err
	}
	name := cfg.Query().Get("name")
	reg := &consulWatcher{
		name: name, consul: client, events: events, aginx: aginx,
		domains: map[string]map[string]*api.AgentService{},
	}
	return reg, nil
}

//先从server meta中获取，然后从外部name labels 获取
func (c *consulWatcher) findLabel(labelFinder addition.LabelFinder, serverLabels map[string]string, name string) ([]*label, error) {
	labels, err := findLabel(serverLabels)
	if err == nil && (labels == nil || len(labels) == 0) {
		externalName := name
		if c.name != "" {
			externalName = c.name + "." + name
		}
		findLabels := labelFinder(externalName, serverLabels)
		labels, err = findLabel(findLabels)
	}
	return labels, err
}

func (c *consulWatcher) watch() error {
	services, meta, err := c.consul.Catalog().Services(&api.QueryOptions{
		WaitIndex:  c.lastIdx, /*WaitTime: time.Second * 10, 放到配置里了 */
		AllowStale: false,
	})
	if err != nil {
		logger.WithError(err).Warn("list domains")
		return err
	}
	if meta.LastIndex == c.lastIdx {
		return nil
	}
	c.lastIdx = meta.LastIndex

	labelFinder := addition.Load(c.aginx, "registry/consul-labels.conf")

	event := registry.LabelsEvent{}
	for serviceName, _ := range services {
		if catalogServiceEntries, _, err := c.consul.Health().Service(serviceName, "", true, nil); err != nil {
			logger.WithError(err).Warn("service health ", serviceName)
			continue
		} else {
			for _, serviceEntry := range catalogServiceEntries {
				if serviceEntry.Checks.AggregatedStatus() != api.HealthPassing {
					continue
				}
				//优先从meta获取如果没有从外部对应处获取
				labels, err := c.findLabel(labelFinder, serviceEntry.Service.Meta, serviceName)
				if err != nil {
					logger.WithError(err).Warn("search %s:%s labels", serviceName, serviceEntry.Service.ID)
					continue
				}
				for _, label := range labels {
					if d := c.add(label, c.lastIdx, serviceEntry); d != nil {
						event = append(event, *d)
					}
				}

			}
		}
	}

	//搜索下线服务
	for domain, idServers := range c.domains {
		for id, service := range idServers {
			if service.CreateIndex < c.lastIdx {
				event = append(event, registry.Domain{
					ID: id, Domain: domain, Alive: false,
					Address: []string{net.JoinHostPort(service.Address, strconv.Itoa(service.Port))},
					//下线服务不需要其他的内容
				})
				delete(idServers, id)
			}
		}
	}

	if len(event) > 0 {
		c.events <- event
	}
	return nil
}

//又返回说明是一个新注册的服务
func (c *consulWatcher) add(label *label, lastIndex uint64, server *api.ServiceEntry) *registry.Domain {
	server.Service.CreateIndex = lastIndex //这里修改这个的目的是保存更新状态，如果没有更新的说明挂掉了

	domain := &registry.Domain{
		ID:      server.Service.ID,
		Domain:  label.Domain,
		Address: []string{net.JoinHostPort(server.Service.Address, strconv.Itoa(server.Service.Port))},
		Weight:  server.Service.Weights.Passing, AutoSSL: label.AutoSSL,
		Alive: true, Template: label.Template, Provider: label.Provider,
	}

	idServer, has := c.domains[label.Domain]
	if !has {
		c.domains[label.Domain] = map[string]*api.AgentService{
			server.Service.ID: server.Service,
		}
		return domain
	}

	serverEntry, has := idServer[server.Service.ID]
	idServer[server.Service.ID] = server.Service
	if !has {
		return domain
	}
	//有修改
	if serverEntry.ModifyIndex < server.Service.ModifyIndex {
		return domain
	}
	return nil
}

func (c *consulWatcher) Start() error {
	go func() {
		for {
			if err := c.watch(); err != nil {
				time.Sleep(time.Second * 5)
			}
		}
	}()
	return nil
}
