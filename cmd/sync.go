package cmd

import (
	"errors"
	"fmt"
	"github.com/ihaiker/aginx/nginx"
	"github.com/ihaiker/aginx/storage"
	fileStorage "github.com/ihaiker/aginx/storage/file"
	. "github.com/ihaiker/aginx/util"
	"github.com/spf13/cobra"
)

var SyncCmd = &cobra.Command{
	Use: "sync", Short: "Sync configuration files from nginx to cluster storage",
	Long: "Sync configuration files to storage", Example: "aginx sync consul://127.0.0.1:8500/aginx",
	Args: cobra.ExactValidArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		defer Catch(func(err error) {
			fmt.Println(err)
		})
		cluster := storage.FindStorage(args[0])
		if cluster == nil || !cluster.IsCluster() {
			return errors.New("the flag cluster not found")
		}
		engine := fileStorage.MustSystem()

		//format
		client := nginx.MustClient("", engine, nil, nil)
		PanicIfError(client.Store())

		return storage.Sync(engine, cluster)
	},
}
