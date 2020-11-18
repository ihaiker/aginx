package cmd

import (
	"github.com/ihaiker/aginx/v2/core/nginx"
	"github.com/ihaiker/aginx/v2/core/storage"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use: "sync", Short: "同步存储配置",
	Long: `在不同的存储上同步nginx配置
	将本机的nginx配置同步到指定存储上: 本机将会使用nginx可执行程序自动寻找，如果不存在可执行程序将无法同步
		aginx sync <存储> 
	在连个不同的存储之间同步:						
		aginx sync <存储1> <存储2>

	存储配置规则：
		consul://127.0.0.1:8500/aginx[?token=authtoken]   consul k/v配置.
		zk://127.0.0.1:2182/aginx[?scheme=&auth=]         zookeeper 配置.
		etcd://127.0.0.1:2379/aginx[?user=&password]      
		file://etc/nginx/nginx.conf                       本机配置       
		其他插件存储配置方式
	`,
	Example: `
		aginx sync consul://127.0.0.1:8500/aginx
		aginx sync consul://127.0.0.1:8500/aginx etcd://127.0.0.1:2379/aginx
		aginx sync file://etc/nginx/nginx.conf etcd://127.0.0.1:2379/aginx
	`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			if _, _, conf, err := nginx.Nginx(); err != nil {
				return err
			} else {
				args = []string{"file:/" + conf, args[0]}
			}
		}

		fromStorage, err := storage.Get(args[0])
		if err != nil {
			return err
		}
		toStorage, err := storage.Get(args[1])
		if err != nil {
			return err
		}
		return storage.Sync(fromStorage, toStorage)
	},
}

func init() {
	syncCmd.SilenceUsage = true
	syncCmd.SilenceErrors = true
}
