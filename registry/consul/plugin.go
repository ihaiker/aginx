package consul

import "github.com/ihaiker/aginx/plugins"

var Plugin = &plugins.RegistryPlugin{
	Name:             "consul",
	LoadRegistry:     LoadRegistry,
	AddRegistryFlags: AddRegistryFlags,
	Support:          plugins.RegistrySupportAll,
}
