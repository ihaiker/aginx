package cmd

import (
	"errors"
	"fmt"
	"github.com/ihaiker/aginx/server/ignore"
	"github.com/ihaiker/aginx/storage/file"
	. "github.com/ihaiker/aginx/util"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var ClusterCmd = &cobra.Command{
	Use: "cluster", Short: "Sync configuration files from nginx to cluster storage",
	Long: "Sync configuration files to cluster storage",
	Args: cobra.ExactValidArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		defer Catch(func(err error) {
			fmt.Println(err)
		})

		engine := clusterConfiguration(args[0], ignore.Empty())
		if engine == nil {
			return errors.New("the flag cluster not found")
		}

		_, conf, err := file.GetInfo()
		PanicIfError(err)

		_ = engine.Remove("")

		root := filepath.Dir(conf)
		return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return err
			}
			if bs, err := ioutil.ReadFile(path); err != nil {
				return err
			} else {
				file := strings.Replace(path, root+"/", "", 1)
				logrus.WithField("module", "cluster").Info("sync file ", file)
				return engine.Store(file, bs)
			}
		})
	},
}
