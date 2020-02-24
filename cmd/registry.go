package cmd

import (
	"github.com/ihaiker/aginx/registry"
	"github.com/ihaiker/aginx/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func AddRegistryFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP("api", "i", "", "restful api address. (default :8011)")
	cmd.PersistentFlags().StringP("security", "s", "", "base auth for restful api, example: user:passwd")

	cmd.PersistentFlags().StringP("template", "", "", `Local template directory, It is used to generate NGINX configuration files.`)
	cmd.PersistentFlags().StringP("storage-template", "", "", `AGINX '--storage' template directory, It is used to generate NGINX configuration files.`)
	registry.RegisterFlags(cmd)
}

var RegistryCmd = &cobra.Command{
	Use: "registry", Short: "the aginx registry server", Example: "aginx registry -D",
	RunE: func(cmd *cobra.Command, args []string) error {
		defer util.Catch(func(err error) {
			cmd.PrintErr(err)
		})
		bridge := registry.FindRegistry(cmd)
		util.AssertTrue(bridge == nil, "flag --docker, --consul, --etcd, --zk need at least one.")
		return util.NewDaemon().Add(bridge).Start()
	},
}

func init() {
	AddRegistryFlag(RegistryCmd)
	_ = viper.BindPFlags(RegistryCmd.PersistentFlags())
}
