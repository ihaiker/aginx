package lego

import (
	"github.com/go-acme/lego/v3/log"
)

func Lego(defaultPath string) error {
	log.Logger = new(Stdlout)
	//app := cli.NewApp()
	//app.Name = "lego"
	//app.HelpName = "lego"
	//app.Usage = "Let's Encrypt client written in Go"
	//app.EnableBashCompletion = true
	//app.Version = "aginx lego"
	//
	//cli.VersionPrinter = func(c *cli.Context) {
	//	fmt.Printf("lego version %s %s/%s\n", c.App.Version, runtime.GOOS, runtime.GOARCH)
	//}
	//
	//app.Flags = cmd.CreateFlags(defaultPath)
	//app.Before = cmd.Before
	//app.Commands = cmd.CreateCommands()
	//
	//return app.Run(os.Args)
	return nil
}
