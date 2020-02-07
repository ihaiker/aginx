package main

import (
	"fmt"
	"github.com/ihaiker/aginx/cmd/aginx"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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

func setLogger(cmd *cobra.Command) error {
	logrus.SetReportCaller(true)
	if debug, err := cmd.Root().PersistentFlags().GetBool("debug"); err != nil {
		return err
	} else if debug {
		logrus.SetLevel(logrus.DebugLevel)
	} else if level, err := cmd.Root().PersistentFlags().GetString("level"); err != nil {
		return err
	} else if logrusLevel, err := logrus.ParseLevel(level); err != nil {
		return err
	} else {
		logrus.SetLevel(logrusLevel)
	}
	return nil
}

var rootCmd = &cobra.Command{
	Use:     "aginx",
	Long:    fmt.Sprintf(`api for nginx. Build: %s, Go: %s, Commit: %s`, BUILD_TIME, runtime.Version(), GITLOG_VERSION),
	Version: "" + VERSION + "",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
		if err = setLogger(cmd); err != nil {
			return
		}
		return
	},
	RunE: aginx.Start,
}

func init() {
	cobra.OnInitialize(func() {})
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "debug mode")
	rootCmd.PersistentFlags().StringP("level", "l", "info", "log level")
	rootCmd.PersistentFlags().StringP("api", "a", ":8011", "restful api port")
	rootCmd.PersistentFlags().StringP("security", "s", "", "base auth for restful api, example: user:passwd.")
}

func main() {
	rand.Seed(time.Now().Unix())
	runtime.GOMAXPROCS(runtime.NumCPU())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
