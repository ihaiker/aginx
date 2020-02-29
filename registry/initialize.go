package registry

import (
	"fmt"
	aginx "github.com/ihaiker/aginx/api"
	"github.com/ihaiker/aginx/logs"
	"github.com/ihaiker/aginx/plugins"
	"github.com/ihaiker/aginx/registry/bridge"
	"github.com/ihaiker/aginx/util"
	"github.com/spf13/cobra"
	"net"
	"strings"
)

var logger = logs.New("registry")

var registryPlugins = findPlugins()

func RegisterFlags(cmd *cobra.Command) {
	for name, registryPlugin := range registryPlugins {
		registryPlugin.AddRegistryFlags(cmd)
		if registryPlugin.Support.Support(plugins.RegistrySupportLabel) {
			cmd.PersistentFlags().StringP(fmt.Sprintf("%s-labels-template-dir", registryPlugin.Name), "", "",
				fmt.Sprintf(`Template file directory (must be a relative path),
It is used to search ${domain}.ngx.tpl or default.tpl generate NGINX configuration files.
default: templates/%s`, registryPlugin.Name))
		}
		if registryPlugin.Support.Support(plugins.RegistrySupportTemplate) {
			cmd.PersistentFlags().StringP(fmt.Sprintf("%s-template", name), "", "",
				fmt.Sprintf(`Template file, It is used to generate NGINX configuration files. 
default: templates/%s.tpl`, registryPlugin.Name))
		}
	}
}

func withAginx(cmd *cobra.Command) aginx.Aginx {
	address := util.GetString(cmd, "api", "127.0.0.1:8011")
	host, port, err := net.SplitHostPort(address)
	util.PanicIfError(err)

	if host == "" || host == "0.0.0.0" {
		host = "127.0.0.1"
	}
	api := aginx.New(fmt.Sprintf("http://%s:%s", host, port))
	if security := util.GetString(cmd, "security", ""); security != "" {
		userAndPwd := strings.SplitN(security, ":", 2)
		api.Auth(userAndPwd[0], userAndPwd[1])
	}
	return api
}

func FindRegistry(cmd *cobra.Command) *MultiRegister {
	registries := new(MultiRegister)

	api := withAginx(cmd)

	for name, registryPlugin := range registryPlugins {
		if register, err := registryPlugin.LoadRegistry(cmd); err != nil {
			logger.Errorf("load %s registry error: %s", name, err.Error())
		} else if register != nil {
			logger.Info("start using registry ", name)
			if register.Support().Support(plugins.RegistrySupportLabel) {

				templateDir := util.GetString(cmd,
					fmt.Sprintf("%s-labels-template-dir", registryPlugin.Name),
					fmt.Sprintf("templates/%s", registryPlugin.Name))
				if strings.HasPrefix(templateDir, "/") {
					templateDir = templateDir[1:]
				}
				registerBridge := &bridge.LabelRegisterBridge{
					Aginx:         api,
					Register:      register,
					Name:          registryPlugin.Name,
					TemplateDir:   templateDir,
					TemplateFuncs: registryPlugin.TemplateFuns(),
				}
				registries.Add(registerBridge)

			} else if register.Support().Support(plugins.RegistrySupportTemplate) {

				templateFile := util.GetString(cmd, fmt.Sprintf("%s-template", name), fmt.Sprintf("templates/%s.tpl", registryPlugin.Name))
				if strings.HasPrefix(templateFile, "/") {
					templateFile = templateFile[1:]
				}
				registerBridge := &bridge.TemplateRegisterBridge{
					Aginx:         api,
					Register:      register,
					Name:          registryPlugin.Name,
					Template:      templateFile,
					TemplateFuncs: registryPlugin.TemplateFuns(),
				}
				registries.Add(registerBridge)
			}
		}
	}

	if registries.Size() == 0 {
		return nil
	}
	return registries
}
