package registry

import (
	"errors"
	"fmt"
	aginx "github.com/ihaiker/aginx/api"
	"github.com/ihaiker/aginx/logs"
	"github.com/ihaiker/aginx/plugins"
	"github.com/ihaiker/aginx/registry/bridge"
	"github.com/ihaiker/aginx/registry/functions"
	"github.com/ihaiker/aginx/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"net"
	"plugin"
	"strings"
)

var logger = logs.New("registry")

var registryPlugins = findPlugins()

func RegisterFlags(cmd *cobra.Command) {
	for _, registryPlugin := range registryPlugins {
		name := registryPlugin.Name
		registryPlugin.AddRegistryFlags(cmd)
		if registryPlugin.Support.Support(plugins.RegistrySupportLabel) {
			cmd.PersistentFlags().StringP(fmt.Sprintf("%s-labels-template-dir", name), "", fmt.Sprintf("templates/%s", name),
				"Template file directory (must be a relative path),\n"+
					"It is used to search ${domain}.ngx.tpl or default.tpl generate NGINX configuration files.")
		}
		if registryPlugin.Support.Support(plugins.RegistrySupportTemplate) {
			cmd.PersistentFlags().StringP(fmt.Sprintf("%s-template", name), "", fmt.Sprintf("templates/%s.tpl", name),
				"Template file, It is used to generate NGINX configuration files.")
		}
		cmd.PersistentFlags().StringP(fmt.Sprintf("%s-template-funcmap", name), "", "", "")
	}
}

func withAginx() aginx.Aginx {
	address := viper.GetString("api")
	host, port, err := net.SplitHostPort(address)
	util.PanicIfError(err)

	if host == "" || host == "0.0.0.0" {
		host = "127.0.0.1"
	}
	api := aginx.New(fmt.Sprintf("http://%s:%s", host, port))
	if security := viper.GetString("security"); security != "" {
		userAndPwd := strings.SplitN(security, ":", 2)
		api.Auth(userAndPwd[0], userAndPwd[1])
	}
	return api
}

func FindRegistry(cmd *cobra.Command) *MultiRegister {
	registries := new(MultiRegister)
	api := withAginx()

	for name, registryPlugin := range registryPlugins {
		if register, err := registryPlugin.LoadRegistry(cmd); err != nil {
			util.PanicIfError(err)
		} else if register != nil {
			logger.Info("start using registry ", name)

			systemFuncs := functions.TemplateFuncs(api)
			if fms := viper.GetString(fmt.Sprintf("%s-template-funcmap", registryPlugin.Name)); fms != "" {
				p, err := plugin.Open(fms)
				util.PanicIfError(err)
				loadFuncMap, err := p.Lookup(plugins.PLUGIN_FUNCMAP)
				util.PanicIfError(err)
				if funcMapMethod, match := loadFuncMap.(plugins.LoadFuncMap); match {
					userDefinedFuncs := funcMapMethod(register)
					systemFuncs = functions.Merge(systemFuncs, userDefinedFuncs)
				} else {
					util.PanicIfError(errors.New("load funcmap error"))
				}
			}

			if register.Support().Support(plugins.RegistrySupportLabel) {
				templateDir := viper.GetString(fmt.Sprintf("%s-labels-template-dir", registryPlugin.Name))
				if strings.HasPrefix(templateDir, "/") {
					templateDir = templateDir[1:]
				}
				registries.Add(&bridge.LabelRegisterBridge{
					Aginx:                 api,
					Register:              register,
					Name:                  registryPlugin.Name,
					TemplateDir:           templateDir,
					AppendTemplateFuncMap: systemFuncs,
				})

			} else if register.Support().Support(plugins.RegistrySupportTemplate) {
				templateFile := viper.GetString(fmt.Sprintf("%s-template", name))
				if strings.HasPrefix(templateFile, "/") {
					templateFile = templateFile[1:]
				}
				registries.Add(&bridge.TemplateRegisterBridge{
					Aginx:                 api,
					Register:              register,
					Name:                  registryPlugin.Name,
					Template:              templateFile,
					AppendTemplateFuncMap: systemFuncs,
				})
			}
		}
	}

	if registries.Size() == 0 {
		return nil
	}
	return registries
}
