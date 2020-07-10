package nginx

import (
	"github.com/ihaiker/aginx/logs"
	"github.com/ihaiker/aginx/nginx/config"
	"github.com/ihaiker/aginx/util"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const NGINX_CONF = "nginx.conf"

var logger = logs.New("nginx")

type Process struct {
	startCmd *exec.Cmd
}

func (sp *Process) start() error {
	err := util.Async(time.Second*5, func() (err error) {
		sp.startCmd, err = util.CmdStart("nginx", "-g", "daemon off;")
		if err == nil {
			err = util.CmdAfterWait(sp.startCmd)
		}
		if err != nil {
			sp.startCmd = nil
		}
		return
	})
	if err == util.ErrTimeout {
		return nil
	}
	return err
}

func (sp *Process) Start() (err error) {
	util.SubscribeFileChanged(sp.Reload)

	if err = sp.start(); err != nil {
		logger.Warn("start NGINX error ", err)
		err = sp.Stop()
		logger.WithError(err).Debug("first stop NGINX")
		err = sp.start()
	}
	logger.WithError(err).Info("start NGINX")
	return
}

func (sp *Process) Reload() error {
	err := util.CmdRun("nginx", "-s", "reload")
	logger.Info("reload NGINX ", err)
	return err
}

func (sp *Process) Test(cfg *config.Configuration, beforeHocks ...func(testDir string) error) (err error) {
	defer util.Catch(func(re error) { err = re })

	configDir := MustConfigDir()
	testDir := "/tmp/nginx"
	util.PanicIfError(os.RemoveAll(testDir))
	util.PanicIfError(util.CopyDir(configDir, testDir))
	util.PanicIfError(WriteTo(testDir, cfg))

	for _, beforeHock := range beforeHocks {
		util.PanicIfError(beforeHock(testDir))
	}

	util.PanicIfError(util.CmdRun("nginx", "-t" /*"-p", path,*/, "-c", filepath.Join(testDir, NGINX_CONF)))
	return
}

func (sp *Process) Stop() error {
	if sp.startCmd != nil {
		return sp.startCmd.Process.Kill()
	} else {
		return util.CmdRun("nginx", "-s", "quit")
	}
}
