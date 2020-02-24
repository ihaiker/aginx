package registry

import (
	"fmt"
	aginx "github.com/ihaiker/aginx/api"
	"github.com/ihaiker/aginx/plugins"
	"github.com/ihaiker/aginx/registry/docker"
	"github.com/ihaiker/aginx/util"
	"github.com/spf13/cobra"
	"net"
	"strings"
)

func RegisterFlags(cmd *cobra.Command) {
	docker.AddFlags(cmd)
}

func withBridge(cmd *cobra.Command, reg plugins.Register) *RegisterBridge {
	address := util.GetString(cmd, "api", ":8011")
	host, port, err := net.SplitHostPort(address)
	util.PanicIfError(err)
	if host == "" || host == "0.0.0.0" {
		host = "127.0.0.1"
	}
	api := aginx.New(fmt.Sprintf("http://%s:%s", host, port))
	if security := util.GetString(cmd, "security", ""); security != "" {
		userAndPwd := strings.SplitN(security, ":", 2)
		api.Auth(userAndPwd[0], userAndPwd[1])
	}

	return &RegisterBridge{
		Register:           reg,
		LocalTemplateDir:   util.GetString(cmd, "template", ""),
		StorageTemplateDir: util.GetString(cmd, "storage-template", ""),
		Aginx:              api,
	}
}

func FindRegistry(cmd *cobra.Command) *RegisterBridge {
	if register := docker.Cmd(cmd); register != nil {
		logger.Info("use docker registry ")
		return withBridge(cmd, register)
	} else {
		return nil
	}

}
