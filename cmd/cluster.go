package cmd

import (
	"errors"
	"fmt"
	"github.com/ihaiker/aginx/logs"
	"github.com/ihaiker/aginx/server/ignore"
	"github.com/ihaiker/aginx/storage"
	"github.com/ihaiker/aginx/storage/file"
	. "github.com/ihaiker/aginx/util"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path/filepath"
)

func syncupClusterConfiguration(root, appendRelativeDir string, engine storage.Engine) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		if filepath.Ext(path) == ".so" || filepath.Ext(path) == ".dll" {
			return nil
		}

		if info.Mode()&os.ModeSymlink == os.ModeSymlink {
			linkPath, _ := filepath.EvalSymlinks(path)
			if linkInfo, err := os.Stat(linkPath); err != nil {
				return err
			} else if linkInfo.IsDir() {
				relative, _ := filepath.Rel(root, path)
				return syncupClusterConfiguration(linkPath, relative, engine)
			}
		}

		if bs, err := ioutil.ReadFile(path); err != nil {
			return err
		} else {
			file, _ := filepath.Rel(root, path)
			if appendRelativeDir != "" {
				file = filepath.Join(appendRelativeDir, file)
			}
			logs.New("cluster").Info("sync file ", file)
			return engine.Store(file, bs)
		}
	})
}

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

		return syncupClusterConfiguration(root, "", engine)
	},
}
