package docker

import "github.com/ihaiker/aginx/plugins"

var Plugin = &plugins.RegistryPlugin{
	Name:             "docker",
	LoadRegistry:     LoadRegistry,
	AddRegistryFlags: AddRegistryFlags,
	Support:          plugins.RegistrySupportAll,
}
