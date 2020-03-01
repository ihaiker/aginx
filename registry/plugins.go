package registry

import (
	"github.com/ihaiker/aginx/plugins"
	"github.com/ihaiker/aginx/registry/consul"
	"github.com/ihaiker/aginx/registry/docker"
	"github.com/ihaiker/aginx/util"
	"reflect"
)

func userPlugins(registryPlugins map[string]*plugins.RegistryPlugin) {
	userPlugins := util.FindPlugins("registry")
	for name, userPlugin := range userPlugins {
		if method, err := userPlugin.Lookup(plugins.PLUGIN_REGISTRY); err != nil {
			logger.Warnf("plugin %s error: %s", name, err)
		} else if loadRegistry, match := method.(plugins.LoadRegistry); !match {
			register := loadRegistry()
			registryPlugins[register.Name] = register
		} else {
			logger.Warnf("plugin %s error: %s not match %s",
				name, plugins.PLUGIN_REGISTRY, reflect.TypeOf(new(plugins.LoadRegistry)).String())
		}
	}
}

func findPlugins() map[string]*plugins.RegistryPlugin {
	registryPlugins := map[string]*plugins.RegistryPlugin{
		"docker.v1.19.3": docker.Plugin,
		"consul.1.7.1":   consul.Plugin,
	}
	userPlugins(registryPlugins)
	return registryPlugins
}
