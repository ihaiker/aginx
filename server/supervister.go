package server

import (
	"github.com/ihaiker/aginx/nginx/configuration"
	fileStorage "github.com/ihaiker/aginx/storage/file"
	"github.com/ihaiker/aginx/util"
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type Supervister struct {
	startCmd *exec.Cmd
}

func (sp *Supervister) start() error {
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

func (sp *Supervister) Start() (err error) {
	_ = util.EBus.Subscribe(util.NginxReload, sp.Reload)

	if err = sp.start(); err != nil {
		logrus.Debug("start nginx error: ", err)
		if stopErr := sp.stop(); stopErr != nil {
			logrus.Debug("first stop nginx error: ", stopErr)
			return
		} else {
			logrus.Debug("first stop nginx")
		}
		err = sp.start()
	}
	logrus.Infof("start nginx %v", err)
	return
}

func (sp *Supervister) Reload() error {
	err := util.CmdRun("nginx", "-s", "reload")
	logrus.Info("reload nginx ", err)
	return err
}

func (sp *Supervister) Test(cfg *configuration.Configuration) (err error) {
	_, conf, _ := fileStorage.GetInfo()

	dir := filepath.Dir(conf)
	testRoot := filepath.Dir(os.TempDir()) + "/aginx"
	if err = os.RemoveAll(testRoot); err != nil {
		return
	}

	if err = util.CopyDir(dir, testRoot); err != nil {
		return
	}
	if err = configuration.Down(testRoot, cfg); err != nil {
		return
	}
	if err = util.CmdRun("nginx", "-t" /*"-p", path,*/, "-c", testRoot+"/nginx.conf"); err != nil {
		logrus.Info("nginx test error: ", err)
	}
	return
}

func (sp *Supervister) stop() error {
	return util.CmdRun("nginx", "-s", "quit")
}

func (sp *Supervister) Stop() error {
	if sp.startCmd != nil {
		return sp.startCmd.Process.Kill()
	} else {
		return sp.stop()
	}
}
