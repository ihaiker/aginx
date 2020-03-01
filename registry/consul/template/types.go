package consulTemplate

import (
	consulApi "github.com/hashicorp/consul/api"
)

type ConsulTemplateEvent struct {
	Services map[string][]*consulApi.ServiceEntry
	Keys     map[string]consulApi.KVPair
	Consul   *consulApi.Client
}
