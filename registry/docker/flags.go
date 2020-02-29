package docker

import (
	"errors"
	"github.com/ihaiker/aginx/plugins"
	dockerLabels "github.com/ihaiker/aginx/registry/docker/labels"
	dockerTemplates "github.com/ihaiker/aginx/registry/docker/templates"
	"github.com/ihaiker/aginx/util"
	"github.com/spf13/cobra"
	"os"
	"regexp"
	"strings"
)

func AddRegistryFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolP("docker", "D", false, `Automatically configure docker containers or services to NGINX。
see how to configure  docker-client :https://github.com/docker/engine/tree/master/client, 
and AGINX flags --docker-host, --docker-tls-verify, --docker-cert-path, --docker-api-version`)

	cmd.PersistentFlags().StringP("docker-host", "", "", "Set the url to the docker server。You can use system variables DOCKER_HOST or AGINX_DOCKER_HOST")
	cmd.PersistentFlags().StringP("docker-tls-verify", "", "", "to enable or disable TLS verification, off by default.\n"+
		"You can use system variables DOCKER_TLS_VERIFY or AGINX_DOCKER_TLS_VERIFY")
	cmd.PersistentFlags().StringP("docker-cert-path", "", "", "Load the TLS certificates from. \n"+
		"You can use system variables DOCKER_CERT_PATH or AGINX_DOCKER_CERT_PATH")
	cmd.PersistentFlags().StringP("docker-api-version", "", "", "Set the version of the API to reach, leave empty for latest (1.40). \n"+
		"You can use system variables DOCKER_API_VERSION or AGINX_DOCKER_API_VERSION")

	cmd.PersistentFlags().BoolP("docker-template-mode", "", false, "Use template mode, You can use system variables AGINX_DOCKER_TEMPLATE_MODE")

	cmd.PersistentFlags().StringArrayP("docker-filter-service", "", []string{".*"}, "Filter services that need attention, see regexp")
	cmd.PersistentFlags().StringArrayP("docker-filter-container", "", []string{".*"}, "Filtering containers that need attention, see regexp")

	cmd.PersistentFlags().StringP("ip", "", "", `IP for ports mapped to the host`)
}

func dockerEnv(cmd *cobra.Command, keys ...string) {
	for _, key := range keys {
		if value := util.GetString(cmd, key, ""); value != "" {
			envKey := strings.ToUpper(strings.ReplaceAll(key, "-", "_"))
			_ = os.Setenv(envKey, value)
		}
	}
}

func LoadRegistry(cmd *cobra.Command) (plugins.Register, error) {
	if util.GetBool(cmd, "docker") == false {
		return nil, nil
	}
	ip := util.GetString(cmd, "ip", "")
	dockerEnv(cmd, "docker-host", "docker-tls-verify", "docker-cert-path")

	if util.GetBool(cmd, "docker-template-mode") {
		filterServices := util.GetStringArray(cmd, "docker-filter-service", []string{".*"})
		for _, filterService := range filterServices {
			if _, err := regexp.Compile(filterService); err != nil {
				return nil, errors.New("--docker-filter-service error : " + err.Error())
			}
		}
		filterContainers := util.GetStringArray(cmd, "docker-filter-container", []string{".*"})
		for _, filterContainer := range filterContainers {
			if _, err := regexp.Compile(filterContainer); err != nil {
				return nil, errors.New("--docker-filter-container error : " + err.Error())
			}
		}
		return dockerTemplates.TemplateRegister(ip, filterServices, filterContainers)
	}

	return dockerLabels.LabelsRegister(ip)
}

var Plugin = &plugins.RegistryPlugin{
	Name:             "docker",
	LoadRegistry:     LoadRegistry,
	AddRegistryFlags: AddRegistryFlags,
	Support:          plugins.RegistrySupportAll,
	TemplateFuns:     templateFuncs,
}
