package server

import (
	"github.com/ihaiker/aginx/logs"
	"github.com/ihaiker/aginx/nginx/configuration"
	fileStorage "github.com/ihaiker/aginx/storage/file"
	"github.com/ihaiker/aginx/util"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

var logger = logs.New("server")

type Supervister struct {
	startCmd *exec.Cmd
}

func (sp *Supervister) start() error {
	err := util.Async(time.Second*5, func() (err error) {
		sp.startCmd, err = util.CmdStart("nginx", "-g", "daemon off;")
		sp.startCmd.Stdout = os.Stdout
		sp.startCmd.Stderr = os.Stderr
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

	_ = util.EBus.Subscribe(util.StorageFileChanged, func() (err error) {
		if err = util.CmdRun("nginx", "-t"); err != nil {
			logger.Error("nginx configuration test error ", err)
			return
		}
		err = sp.Reload()
		return
	})

	if err = sp.start(); err != nil {
		logger.WithField("-", "supervister").Warn("start nginx error ", err)
		err = sp.stop()
		logger.WithField("-", "supervister").WithError(err).Debug("first stop nginx")
		err = sp.start()
	}
	logger.WithField("-", "supervister").WithError(err).Info("start nginx")
	return
}

func (sp *Supervister) Reload() error {
	err := util.CmdRun("nginx", "-s", "reload")
	logger.WithField("-", "supervister").Info("reload nginx ", err)
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
		logger.WithField("-", "supervister").Info("nginx test error: ", err)
		return
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
