package main

import (
	"fmt"
	"github.com/ihaiker/aginx/cmd"
	"github.com/ihaiker/aginx/logs"
	"github.com/ihaiker/aginx/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"math/rand"
	"os"
	"runtime"
	"time"
)

var (
	VERSION        = "v0.0.1"
	BUILD_TIME     = "2012-12-12 12:12:12"
	GITLOG_VERSION = "0000"
)

var rootCmd = &cobra.Command{
	Use:     "aginx",
	Long:    fmt.Sprintf(`api for nginx. Build: %s, Go: %s, Commit: %s`, BUILD_TIME, runtime.Version(), GITLOG_VERSION),
	Version: "" + VERSION + "",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
		defer util.CatchError(err)
		debug := viper.GetBool("debug")
		level := viper.GetString("level")
		util.PanicIfError(logs.SetLogger(debug, level))
		return
	},
}

func init() {
	cobra.OnInitialize(func() {
		viper.SetEnvPrefix("AGINX")
		viper.AutomaticEnv()
	})
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "debug mode")
	rootCmd.PersistentFlags().StringP("level", "l", "info", "log level")
	_ = viper.BindPFlags(rootCmd.PersistentFlags())

	rootCmd.AddCommand(cmd.ServerCmd, cmd.SyncCmd, cmd.RegistryCmd)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	rand.Seed(time.Now().Unix())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
