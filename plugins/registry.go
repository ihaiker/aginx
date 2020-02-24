package plugins

type Domain struct {
	ID      string
	Domain  string
	Address string
	Weight  int
	AutoSSL bool
	Attrs   map[string]string
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

func (ds Domains) GetDomains() []string {
	domains := make([]string, 0)
	for domain, _ := range ds.Group() {
		domains = append(domains, domain)
	}
	return domains
}

type RegistryEventType int

const (
	Online RegistryEventType = iota
	Offline
)

type RegistryDomainEvent struct {
	EventType RegistryEventType
	Servers   Domains
}

type Register interface {
	Sync() Domains

	Get(domain string) Domains

	Listener() <-chan RegistryDomainEvent
}
