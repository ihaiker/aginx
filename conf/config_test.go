package conf_test

import (
	. "github.com/ihaiker/aginx/cmd"
	"github.com/ihaiker/aginx/conf"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"testing"
)

func TestConfig(t *testing.T) {
	cmd := &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {
			email := viper.GetString("email")
			t.Log(email)
		},
	}
	AddServerFlags(cmd)
	viper.SetEnvPrefix("AGINX")
	viper.AutomaticEnv()
	_ = viper.BindPFlags(cmd.PersistentFlags())

	_ = cmd.Execute()

	if err := conf.ReadConfig("aginx.conf", cmd); err != nil {
		t.Fatal(err)
	}
	_ = cmd.Execute()

	_ = os.Setenv("AGINX_EMAIL", "env@env.com")
	_ = cmd.Execute()

	cmd.SetArgs([]string{"--email", "test@test.com"})
	_ = cmd.Execute()
}
