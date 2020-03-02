package util

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"strings"
)

func GetStringArray(cmd *cobra.Command, key string) []string {
	services := viper.GetStringSlice(key)
	if len(services) == 1 &&
		strings.HasPrefix(services[0], "[") && strings.HasSuffix(services[0], "]") {
		services, _ = cmd.PersistentFlags().GetStringArray(key)
	} else {
		if ss, _ := cmd.PersistentFlags().GetStringArray(key); len(ss) != 0 {
			services = ss
		}
	}
	return services
}
