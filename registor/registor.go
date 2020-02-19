package registor

import (
	"encoding/json"
	"github.com/ihaiker/aginx/util"
)

type Server interface {
	ID() string
	Domain() string
	Address() string
	Weight() int
}

type Servers []Server

func (ds Servers) Group() map[string] /*domain*/ Servers {
	groups := map[string]Servers{}
	for _, d := range ds {
		domain := d.Domain()
		if _, has := groups[domain]; has {
			groups[domain] = append(groups[domain], d)
		} else {
			groups[domain] = []Server{d}
		}
	}
	return groups
}

func (ds Servers) String() string {
	bs, _ := json.Marshal(ds)
	return string(bs)
}

type EventType int

const (
	Online EventType = iota
	Offline
)

type ServerEvent struct {
	EventType EventType
	Servers   Servers
}

type Registrator interface {
	util.Service
	Sync() Servers
	Get(domain string) Servers
	Listener() <-chan ServerEvent
}
