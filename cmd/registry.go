package cmd

import (
	"github.com/ihaiker/aginx/registry"
	"github.com/ihaiker/aginx/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func AddRegistryFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP("api", "i", "", "restful api address. (default 127.0.0.1:8011)")
	cmd.PersistentFlags().StringP("security", "s", "", "base auth for restful api, example: user:passwd")
	registry.RegisterFlags(cmd)
}

var RegistryCmd = &cobra.Command{
	Use: "registry", Short: "the AGINX registry server", Example: "aginx registry --docker",
	RunE: func(cmd *cobra.Command, args []string) error {
		defer util.Catch(func(err error) {
			cmd.PrintErr(err)
		})
		bridge := registry.FindRegistry(cmd)
		util.AssertTrue(bridge == nil, "Did not find any registry")
		return util.NewDaemon().Add(bridge).Start()
	},
}

func init() {
	AddRegistryFlag(RegistryCmd)
	_ = viper.BindPFlags(RegistryCmd.PersistentFlags())
}
