package plugins

import (
	"github.com/ihaiker/aginx/util"
	"github.com/spf13/cobra"
	"text/template"
)

const (
	PLUGIN_REGISTRY = "LoadRegistry"
)

type RegistrySupport int

func (r1 RegistrySupport) Support(r2 RegistrySupport) bool { return r1&r2 == r2 }

const (
	RegistrySupportLabel    RegistrySupport = 0b01
	RegistrySupportTemplate                 = 0b10
	RegistrySupportAll                      = RegistrySupportLabel | RegistrySupportTemplate
)

type (
	RegistryPlugin struct {
		LoadRegistry     func(cmd *cobra.Command) (Register, error)
		AddRegistryFlags func(cmd *cobra.Command)
		TemplateFuns     func() template.FuncMap
		Support          RegistrySupport
		Name             string
	}

	LoadRegistry func() *RegistryPlugin

	Register interface {
		util.Service
		Support() RegistrySupport
		Listener() <-chan interface{}
	}
)

type (
	Domain struct {
		ID      string
		Domain  string
		Address string
		Weight  int
		AutoSSL bool
		Attrs   map[string]string
	}
	Domains []Domain

	LabelsRegistryEvent map[string]Domains
)

func (ds Domains) Group() LabelsRegistryEvent {
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

func (ds Domains) GetDomains() []string {
	domains := make([]string, 0)
	for domain, _ := range ds.Group() {
		domains = append(domains, domain)
	}
	return domains
}
