package docker

import (
	"github.com/ihaiker/aginx/util"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

func AddFlags(cmd *cobra.Command) {
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

	cmd.PersistentFlags().StringP("ip", "", "", `IP for ports mapped to the host`)
}

func dockerEnv(cmd *cobra.Command, key string) {
	if value := util.GetString(cmd, key, ""); value != "" {
		envKey := strings.ToUpper(strings.ReplaceAll(key, "-", "_"))
		err := os.Setenv(envKey, value)
		logger.WithError(err).Debugf("find %s, set  %s=%s", key, envKey, value)
	}
}

func Cmd(cmd *cobra.Command) *DockerRegistor {
	if util.GetBool(cmd, "docker") {
		ip := util.GetString(cmd, "ip", "")
		dockerEnv(cmd, "docker-host")
		dockerEnv(cmd, "docker-tls-verify")
		dockerEnv(cmd, "docker-cert-path")
		reg, err := Register(ip)
		util.PanicIfError(err)
		return reg
	}
	return nil
}
