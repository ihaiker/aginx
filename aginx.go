package main

import (
	"fmt"
	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/ihaiker/aginx/cmd"
	"github.com/sirupsen/logrus"
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

func setLogger(cmd *cobra.Command) error {
	//logrus.SetReportCaller(true)
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
	RunE: cmd.ServerCmd.RunE,
}

func init() {
	cobra.OnInitialize(func() {
		viper.SetEnvPrefix("AGINX")
		viper.AutomaticEnv()
	})
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "debug mode")
	rootCmd.PersistentFlags().StringP("level", "l", "info", "log level")
	cmd.AddServerFlags(rootCmd)
	_ = viper.BindPFlags(rootCmd.PersistentFlags())
	rootCmd.AddCommand(cmd.ServerCmd, cmd.ClusterCmd)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	rand.Seed(time.Now().Unix())
	logrus.SetFormatter(&nested.Formatter{
		TimestampFormat: "2006-01-02 15:04:05.000", FieldsOrder: []string{"engine"},
	})
	logrus.SetOutput(os.Stdout)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
