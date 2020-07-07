package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/ihaiker/aginx/api"
	"github.com/ihaiker/aginx/nginx"
	"github.com/ihaiker/aginx/plugins"
	"github.com/ihaiker/aginx/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"strings"
)

var aginx api.Aginx

func preRun(cmd *cobra.Command, args []string) {
	address := "http://" + viper.GetString("api")
	security := viper.GetString("security")
	aginx = api.New(address)
	if security != "" {
		userAndPwd := strings.SplitN(security, ":", 2)
		aginx.Auth(userAndPwd[0], userAndPwd[1])
	}
}

var reloadCmd = &cobra.Command{
	Use: "reload", Short: "reload nginx",
	Args: cobra.NoArgs, PreRun: preRun, Example: "aginx client reload",
	RunE: func(cmd *cobra.Command, args []string) error {
		return aginx.Reload()
	},
}

var selectCmd = &cobra.Command{
	Use: "select", Short: "select configuration", PreRun: preRun,
	Example: "aginx client select http \"include('conf.d/*.conf')\" '*' server",
	RunE: func(cmd *cobra.Command, args []string) error {
		directives, err := aginx.Directive().Select(args...)
		for _, directive := range directives {
			fmt.Println(directive.Pretty(0))
		}
		if os.IsNotExist(err) {
			fmt.Println("## not found !!")
			return nil
		}
		return err
	},
}
var addCmd = &cobra.Command{
	Use: "add", Short: "add configuration",
	PreRun: preRun, Example: "cat <file> | aginx client add http",
	RunE: func(cmd *cobra.Command, args []string) error {
		defer util.Catch(func(err error) {
			fmt.Println(err)
		})

		bs, err := ioutil.ReadAll(os.Stdin)
		util.PanicIfError(err)
		if len(bs) == 0 {
			return fmt.Errorf("add content is empty: %s", err)
		}

		conf, err := nginx.ReaderReadable(nil, plugins.NewFile("", bs))
		util.PanicIfError(err)
		if len(conf.Body) > 0 {
			return fmt.Errorf("add content is empty")
		}

		err = aginx.Directive().Add(args, conf.Body...)
		util.PanicIfError(err)
		return err
	},
}

var modifyCmd = &cobra.Command{
	Use: "modify", Short: "modify configuration",
	PreRun: preRun, Example: "cat <file> | aginx client modify http",
	RunE: func(cmd *cobra.Command, args []string) error {
		defer util.Catch(func(err error) {
			fmt.Println(err)
		})

		bs, err := ioutil.ReadAll(os.Stdin)
		util.PanicIfError(err)
		if len(bs) == 0 {
			return fmt.Errorf("modify content is empty: %s", err)
		}

		conf, err := nginx.ReaderReadable(nil, plugins.NewFile("", bs))
		util.PanicIfError(err)
		if len(conf.Body) != 1 {
			return fmt.Errorf("the modify content must be only one")
		}

		err = aginx.Directive().Modify(args, conf.Body[0])
		util.PanicIfError(err)
		return err
	},
}

var deleteCmd = &cobra.Command{
	Use: "delete", Short: "delete configuration", PreRun: preRun, Args: cobra.MinimumNArgs(1),
	Example: "aginx client delete http server.server_name('api.aginx.io')",
	RunE: func(cmd *cobra.Command, args []string) error {
		return aginx.Directive().Delete(args...)
	},
}

var sslCmd = &cobra.Command{
	Use: "ssl", Short: "new ssl",
	PreRun: preRun, Args: cobra.ExactArgs(1), Example: "aginx client ssl api.aginx.io",
	RunE: func(cmd *cobra.Command, args []string) error {
		email := viper.GetString("email")
		lego, err := aginx.SSL().New(email, args[0])
		if err == nil {
			bs, _ := json.MarshalIndent(lego, "", "\t")
			fmt.Println(string(bs))
		}
		return err
	},
}

var getCmd = &cobra.Command{
	Use: "get", Short: "get file", PreRun: preRun, Args: cobra.ExactArgs(1),
	Example: "aginx client get conf.d/default.conf",
	RunE: func(cmd *cobra.Command, args []string) error {
		content, err := aginx.File().Get(args[0])
		if err == nil {
			fmt.Println(content)
		}
		return err
	},
}

var searchCmd = &cobra.Command{
	Use: "search", Short: "search file", PreRun: preRun,
	Example: "aginx client search conf.d/*.conf",
	RunE: func(cmd *cobra.Command, args []string) error {
		files, err := aginx.File().Search(args...)
		if err == nil {
			for fileName := range files {
				fmt.Println(fileName)
			}
		}
		return err
	},
}

var removeCmd = &cobra.Command{
	Use: "remove", Short: "remove file",
	PreRun: preRun, Args: cobra.ExactArgs(1),
	Example: "aginx client remove conf.d/default.conf",
	RunE: func(cmd *cobra.Command, args []string) error {
		return aginx.File().Remove(args[0])
	},
}
var uploadCmd = &cobra.Command{
	Use: "upload", Short: "upload files",
	PreRun: preRun, Args: cobra.RangeArgs(1, 2),
	Example: `cat file | aginx client upload conf.d/default.conf
aginx client update /user/path/test.conf hosts.d/test.conf`,
	RunE: func(cmd *cobra.Command, args []string) error {
		defer util.Catch(func(err error) {
			fmt.Println(err)
		})
		if len(args) == 1 {
			content, err := ioutil.ReadAll(os.Stdin)
			util.PanicIfError(err)
			if len(content) == 0 {
				return fmt.Errorf("upload content is empty: %s", err)
			}
			return aginx.File().NewWithContent(args[0], content)
		} else {
			return aginx.File().New(args[1], args[0])
		}
	},
}

var ClientCmd = &cobra.Command{
	Use: "client", Aliases: []string{"cli"}, Short: "the AGINX console",
}

func init() {
	ClientCmd.PersistentFlags().StringP("api", "i", "127.0.0.1:8011", "restful api address.")
	ClientCmd.PersistentFlags().StringP("security", "s", "", "base auth for restful api, example: user:passwd")

	ClientCmd.AddCommand(reloadCmd)
	ClientCmd.AddCommand(selectCmd, addCmd, modifyCmd, deleteCmd)
	ClientCmd.AddCommand(getCmd, removeCmd, searchCmd, uploadCmd)
	sslCmd.PersistentFlags().StringP("email", "u", "", "Register the current account to the ACME server.")
	ClientCmd.AddCommand(sslCmd)

	_ = viper.BindPFlags(ClientCmd.PersistentFlags())
}
