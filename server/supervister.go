package server

import (
	"github.com/ihaiker/aginx/nginx"
	"github.com/ihaiker/aginx/util"
	"github.com/sirupsen/logrus"
	"os/exec"
	"time"
)

type Supervister struct {
	startCmd *exec.Cmd
}

func (sp *Supervister) Start() error {
	err := util.Async(time.Second*5, func() (err error) {
		sp.startCmd, err = util.CmdStart("nginx", "-g", "daemon off;")
		if err != nil {
			sp.startCmd = nil
			return
		}
		return util.CmdAfterWait(sp.startCmd)
	})
	if err == util.ErrTimeout {
		return nil
	}
	logrus.Info("start nginx error: ", err)
	return err
}

func (sp *Supervister) Reload() error {
	err := util.CmdRun("nginx", "-s", "reload")
	logrus.Info("reload nginx ", err)
	return err
}

func (sp *Supervister) Test(conf *nginx.Configuration) (err error) {
	//path, file, _ := nginx.GetInfo()
	//aginx := path + "/aginx"
	//if err = os.MkdirAll(aginx,os.ModeDir); err != nil {
	//	return
	//}
	err = util.CmdRun("nginx", "-t")
	logrus.Info("nginx test ", err)
	return err
}

func (sp *Supervister) Stop() error {
	if sp.startCmd != nil {
		return sp.startCmd.Process.Kill()
	}
	return nil
}
