package cmd

import (
	"fmt"
	"github.com/ihaiker/aginx/v2/core/certs"
	"github.com/ihaiker/aginx/v2/core/registry"
	"github.com/ihaiker/aginx/v2/core/storage"
	"github.com/spf13/cobra"
)

var pluginCmd = &cobra.Command{
	Use: "plugin", Short: "显示插件配置的帮助信息",
	Example: `
		aginx plugin registry/reg       显示所有registry插件
		aginx plugin storage            显示所有storage插件
		aginx plugin storage consul     显示所有consul storage插件的配置帮助信息
		aginx plugin certificate        显示所有 certificate插件
		aginx plugin certs              显示所有 certificate插件
	`,
	Args: cobra.RangeArgs(0, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		switch args[0] {
		default:
			return fmt.Errorf("系统不包含插件不存在：%s", args[0])
		case "registry", "reg":
			if len(args) == 1 {
				for _, p := range registry.Plugins {
					fmt.Println(p.Scheme())
				}
			} else {
				plugin, has := registry.Plugins[args[1]]
				if !has {
					return fmt.Errorf("not found: %s", args[1])
				}
				fmt.Println(plugin.Help())
			}
		case "storage":
			if len(args) == 1 {
				for _, p := range storage.Plugins {
					fmt.Println(p.Scheme())
				}
			} else {
				plugin, has := storage.Plugins[args[1]]
				if !has {
					return fmt.Errorf("not found: %s", args[1])
				}
				fmt.Println(plugin.Help())
			}
		case "certificate", "certs":
			if len(args) == 1 {
				for _, p := range certs.Plugins {
					fmt.Println(p.Scheme())
				}
			} else {
				plugin, has := certs.Plugins[args[1]]
				if !has {
					return fmt.Errorf("not found: %s", args[1])
				}
				fmt.Println(plugin.Help())
			}
		}
		return nil
	},
}

func init() {
	pluginCmd.SilenceUsage = true
	pluginCmd.SilenceErrors = true
}
