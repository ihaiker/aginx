package cmd

import (
	"github.com/spf13/cobra"
	"os"
)

var completionCmd = &cobra.Command{
	Use: "completion", Short: "输出命令帮助",
	Args: cobra.ExactValidArgs(1), ValidArgs: []string{"bash", "zsh", "powershell", "fish"},
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			_ = root.GenBashCompletion(os.Stdout)
		case "zsh":
			_ = root.GenZshCompletion(os.Stdout)
		case "powershell":
			_ = root.GenPowerShellCompletion(os.Stdout)
		case "fish":
			_ = root.GenFishCompletion(os.Stdout, true)
		}
		_ = root.GenBashCompletion(os.Stdout)
	},
}
