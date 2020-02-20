package registry

import (
	"encoding/json"
	"github.com/ihaiker/aginx/util"
)

type Domain struct {
	ID      string
	Domain  string
	Address string
	Weight  int
	AutoSSL bool
	Attrs   map[string]string
}

func (ds Domain) String() string {
	bs, _ := json.Marshal(ds)
	return string(bs)
}

type Domains []Domain

func (ds Domains) Group() map[string] /*domain*/ Domains {
	groups := map[string]Domains{}
	for _, d := range ds {
		domain := d.Domain
		if _, has := groups[domain]; has {
			groups[domain] = append(groups[domain], d)
		} else {
			groups[domain] = []Domain{d}
		}
	}
	return groups
}

func (ds Domains) String() string {
	bs, _ := json.Marshal(ds)
	return string(bs)
}

type EventType int

const (
	Online EventType = iota
	Offline
)

type DomainEvent struct {
	EventType EventType
	Servers   Domains
}

type Registor interface {
	util.Service

	Sync() Domains

	Get(domain string) Domains

	Listener() <-chan DomainEvent
}
