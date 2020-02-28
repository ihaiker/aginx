package consul

import "github.com/spf13/cobra"

func AddRegistryFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolP("consul", "C", false, "Automatically obtain consul registered services and publish them to NGINX.")
}
