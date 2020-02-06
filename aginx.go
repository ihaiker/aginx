package main

import (
	"github.com/spf13/cobra"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

var (
	VERSION        string = "v0.0.1"
	BUILD_TIME     string = "2012-12-12 12:12:12"
	GITLOG_VERSION string = "0000"
)

var rootCmd = &cobra.Command{
	Use:     filepath.Base(os.Args[0]),
	Long:    "api for nginx. \tBuild: " + BUILD_TIME + ", Go: " + runtime.Version() + ", GitLog: " + GITLOG_VERSION,
	Version: VERSION + "",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	cobra.OnInitialize(func() {})
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Debug模式")
	rootCmd.PersistentFlags().StringP("level", "l", "info", "日志级别")
	rootCmd.PersistentFlags().StringP("conf", "f", "", "配置文件")
	//rootCmd.AddCommand(initd.Cmd)
}

func main() {
	//defer logs.CloseAll()
	rand.Seed(time.Now().Unix())
	runtime.GOMAXPROCS(runtime.NumCPU())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
