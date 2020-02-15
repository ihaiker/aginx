package util

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func GetString(cmd *cobra.Command, key, def string) string {
	value, err := cmd.PersistentFlags().GetString(key)
	PanicIfError(err)
	if value == "" {
		value = viper.GetString(key)
	}
	if value == "" {
		return def
	}
	return value
}
func GetBool(cmd *cobra.Command, key string) bool {
	value, err := cmd.PersistentFlags().GetBool(key)
	PanicIfError(err)
	if !value {
		value = viper.GetBool(key)
	}
	return value
}
