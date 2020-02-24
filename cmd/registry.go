package cmd

import (
	"fmt"
	aginx "github.com/ihaiker/aginx/api"
	"github.com/ihaiker/aginx/registry"
	"github.com/ihaiker/aginx/registry/docker"
	"github.com/ihaiker/aginx/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"net"
	"os"
	"strings"
)

func AddRegistryFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP("api", "i", "", "restful api address. (default :8011)")
	cmd.PersistentFlags().StringP("security", "s", "", "base auth for restful api, example: user:passwd")

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

	cmd.PersistentFlags().StringP("template", "", "", `Local template directory, It is used to generate NGINX configuration files.`)
	cmd.PersistentFlags().StringP("storage-template", "", "", `AGINX 'storage' template directory, It is used to generate NGINX configuration files.`)

	cmd.PersistentFlags().StringP("ip", "", "", `IP for ports mapped to the host`)

	//TODO reload 操作必须是API执行，或者给定参数，自从重载参数
	cmd.PersistentFlags().BoolP("auto-reload", "", false, "Configuration file changes NGINX auto reload. no need to call API reload function.")
}

var RegistryCmd = &cobra.Command{
	Use: "registry", Short: "the aginx registry server", Example: "aginx registry -D",
	RunE: func(cmd *cobra.Command, args []string) error {
		defer util.Catch(func(err error) {
			cmd.PrintErr(err)
		})
		bridge, err := findBridge(cmd)
		util.PanicIfError(err)
		util.AssertTrue(bridge == nil, "flag --docker require one.")
		return util.NewDaemon().Add(bridge).Start()
	},
}

func init() {
	AddRegistryFlag(RegistryCmd)
	_ = viper.BindPFlags(RegistryCmd.PersistentFlags())
}

func dockerEnv(cmd *cobra.Command, key string) {
	if value := util.GetString(cmd, key, ""); value != "" {
		envKey := strings.ToUpper(strings.ReplaceAll(key, "-", "_"))
		err := os.Setenv(envKey, value)
		logger.WithError(err).Debugf("find %s, set  %s=%s", key, envKey, value)
	}
}

func findBridge(cmd *cobra.Command) (*registry.RegistorBridge, error) {
	ip := util.GetString(cmd, "ip", "")
	address := util.GetString(cmd, "api", ":8011")

	if util.GetBool(cmd, "docker") {
		dockerEnv(cmd, "docker-host")
		dockerEnv(cmd, "docker-tls-verify")
		dockerEnv(cmd, "docker-cert-path")

		if registor, err := docker.Registor(ip); err != nil {
			return nil, err
		} else {
			host, port, err := net.SplitHostPort(address)
			if err != nil {
				return nil, err
			}
			if host == "" || host == "0.0.0.0" {
				host = "127.0.0.1"
			}
			api := aginx.New(fmt.Sprintf("http://%s:%s", host, port))
			security := util.GetString(cmd, "security", "")
			if security != "" {
				userAndPwd := strings.SplitN(security, ":", 2)
				api.Auth(userAndPwd[0], userAndPwd[1])
			}
			return &registry.RegistorBridge{
				Registrator:        registor,
				LocalTemplateDir:   util.GetString(cmd, "template", ""),
				StorageTemplateDir: util.GetString(cmd, "storage-template", ""),
				Aginx:              api,
			}, nil
		}
	}
	return nil, nil
}
