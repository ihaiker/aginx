package consul

import (
	"errors"
	consulApi "github.com/hashicorp/consul/api"
	"github.com/ihaiker/aginx/plugins"
	consulLabels "github.com/ihaiker/aginx/registry/consul/labels"
	consulTemplate "github.com/ihaiker/aginx/registry/consul/template"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"regexp"
	"strings"
)

func AddRegistryFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolP("consul", "C", false, "Automatically obtain consul registered services and publish them to NGINX.")

	cmd.PersistentFlags().StringP("consul-http-addr", "", "", `Address is the address of the Consul server.
You can use system environment variables AGINX_CONSUL_HTTP_ADDR or CONSUL_HTTP_ADDR`)
	cmd.PersistentFlags().StringP("consul-http-token", "", "", `Token is used to provide a per-request ACL token. which overrides the agent's default token.
You can use system environment variables AGINX_CONSUL_HTTP_TOKEN or CONSUL_HTTP_TOKEN`)
	cmd.PersistentFlags().StringP("consul-http-token-file", "", "", `defines an environment variable name which sets the HTTP token file.
You can use system environment variables AGINX_CONSUL_HTTP_TOKEN_FILE or CONSUL_HTTP_TOKEN_FILE`)
	cmd.PersistentFlags().StringP("consul-http-auth", "", "", `defines an environment variable name which sets the HTTP authentication header.
You can use system environment variables AGINX_CONSUL_HTTP_AUTH or CONSUL_HTTP_AUTH`)
	cmd.PersistentFlags().StringP("consul-http-ssl", "", "", `defines an environment variable name which sets whether or not to use HTTPS.
You can use system environment variables AGINX_CONSUL_HTTP_SSL or CONSUL_HTTP_SSL`)
	cmd.PersistentFlags().StringP("consul-cacert", "", "", `defines an environment variable name which sets the CA file to use for talking to Consul over TLS.
You can use system environment variables AGINX_CONSUL_CACERT or CONSUL_CACERT`)
	cmd.PersistentFlags().StringP("consul-capath", "", "", `defines an environment variable name which sets the path to a directory of CA certs to use for talking to Consul over TLS.
You can use system environment variables AGINX_CONSUL_CAPATH or CONSUL_CAPATH`)
	cmd.PersistentFlags().StringP("consul-client-cert", "", "", `defines an environment variable name which sets the client cert file to use for talking to Consul over TLS.
You can use system environment variables AGINX_CONSUL_CLIENT_CERT or CONSUL_CLIENT_CERT`)
	cmd.PersistentFlags().StringP("consul-client-key", "", "", `defines an environment variable name which sets the client key file to use for talking to Consul over TLS.
You can use system environment variables AGINX_CONSUL_CLIENT_KEY or CONSUL_CLIENT_KEY`)
	cmd.PersistentFlags().StringP("consul-tls-server-name", "", "", `defines an environment variable name which sets the server name to use as the SNI host when connecting via TLS
You can use system environment variables AGINX_CONSUL_TLS_SERVER_NAME or CONSUL_TLS_SERVER_NAME`)
	cmd.PersistentFlags().StringP("consul-http-ssl-verify", "", "", `defines an environment variable name which sets whether or not to disable certificate checking.
You can use system environment variables AGINX_CONSUL_HTTP_SSL_VERIFY or CONSUL_HTTP_SSL_VERIFY`)

	cmd.PersistentFlags().StringP("consul-datacenter", "", "dc1", `Datacenter to use. If not provided, the default agent datacenter is used. 
You can use system environment variables AGINX_CONSUL_DATACENTER`)

	cmd.PersistentFlags().StringArrayP("consul-filter", "", []string{".*"}, `Filter services that need attention, see regexp
You can use system environment variables AGINX_CONSUL_FILTER`)

	cmd.PersistentFlags().BoolP("consul-template-mode", "", false, `Use template mode, You can use system variables AGINX_CONSUL_TEMPLATE_MODE`)
}

func consulEnv(envs ...string) {
	for _, env := range envs {
		value := viper.GetString(env)
		if value != "" {
			envKey := strings.ReplaceAll(env, "-", "_")
			_ = os.Setenv(envKey, value)
		}
	}
}

func LoadRegistry(cmd *cobra.Command) (plugins.Register, error) {
	if !viper.GetBool("consul") {
		return nil, nil
	}

	consulEnv("consul-http-addr", "consul-http-token", "consul-http-token-file",
		"consul-http-auth", "consul-http-ssl", "consul-cacert", "consul-capath", "consul-client-cert", "consul-client-key",
		"consul-tls-server-name", "consul-http-ssl-verify")

	filters := viper.GetStringSlice("consul-filter")
	for _, filter := range filters {
		if _, err := regexp.Compile(filter); err != nil {
			return nil, errors.New("--consul-filter error : " + err.Error())
		}
	}

	config := consulApi.DefaultConfig()
	config.Datacenter = viper.GetString("consul-datacenter")

	if client, err := consulApi.NewClient(config); err != nil {
		return nil, err
	} else if viper.GetBool("consul-template-mode") {
		return consulTemplate.NewTemplateRegister(client, filters), nil
	} else {
		return consulLabels.NewLabelRegister(client), nil
	}
}

var Plugin = &plugins.RegistryPlugin{
	Name:             "consul",
	LoadRegistry:     LoadRegistry,
	AddRegistryFlags: AddRegistryFlags,
	Support:          plugins.RegistrySupportAll,
}
