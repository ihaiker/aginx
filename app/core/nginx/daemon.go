package nginx

import (
	"fmt"
	"github.com/ihaiker/aginx/v2/core/logs"
	"github.com/ihaiker/aginx/v2/core/nginx/config"
	"github.com/ihaiker/aginx/v2/core/util/cmds"
	"github.com/ihaiker/aginx/v2/core/util/errors"
	"github.com/ihaiker/aginx/v2/core/util/files"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

var logger = logs.New("nginx")

//控制nginx启动
type Daemon interface {
	Start() error
	Reload() error
	Stop() error
	Test(cfg *config.Configuration, hocks ...func(testDir string) error) (err error)
}

//自动寻找可执行程序和配置文件
func LookupDaemon() (Daemon, error) {
	nginxBin, err := Lookup()
	if err != nil {
		return nil, err
	}
	prefix, nginxConf, err := HelpInfo(nginxBin)
	if err != nil {
		return nil, err
	}
	return NewDaemon(nginxBin, prefix, nginxConf)
}

//初始一个nginx管理器，bin为nginx可以执行程序，conf为配置文件
func NewDaemon(bin, prefix, conf string) (Daemon, error) {
	if !strings.HasSuffix(conf, "nginx.conf") {
		conf = filepath.Join(conf, "nginx.conf")
	}
	if !files.Exists(conf) {
		return nil, fmt.Errorf("not found %s", conf)
	}
	if !files.Exists(bin) {
		return nil, fmt.Errorf("not found %s", bin)
	}
	return &localDaemon{bin: bin, prefix: prefix, nginxConf: conf}, nil
}

type localDaemon struct {
	daemon      *exec.Cmd
	bin, prefix string
	nginxConf   string
}

func (sp *localDaemon) params(args ...string) []string {
	if sp.prefix == "" {
		return append([]string{"-p", sp.prefix, "-c", sp.nginxConf}, args...)
	} else {
		return append([]string{"-c", sp.nginxConf}, args...)
	}
}

func (sp *localDaemon) start(wait time.Duration) error {
	err := errors.Async(wait, func() (err error) {
		if sp.daemon, err = cmds.CmdStart(sp.bin, sp.params("-g", "daemon off;")...); err == nil {
			err = cmds.CmdAfterWait(sp.daemon)
		}
		if err != nil {
			sp.daemon = nil
		}
		return
	})
	if err == errors.ErrTimeout {
		return nil
	}
	logger.Info("start nginx")
	return err
}

func (sp *localDaemon) Start() (err error) {
	if err = sp.start(time.Second * 5); err != nil {
		logger.Warn("start nginx error ", err)
		if err = sp.Stop(); err != nil {
			return
		}
		logger.Debug("first stop nginx")
		err = sp.start(time.Second * 5)
	}
	if err == nil {
		logger.Debug("start nginx daemon")
	}
	return
}

func (sp *localDaemon) Reload() error {
	if sp.daemon == nil {
		//未启动不用，或者这个地方只是不需要托管的其他节点，例如：纯api节点，纯registry节点
		return nil
	}
	err := cmds.CmdRun(sp.bin, sp.params("-s", "reload")...)
	if err != nil {
		logger.Debug("reload nginx")
	}
	return err
}

func (sp *localDaemon) Test(cfg *config.Configuration, hocks ...func(testDir string) error) (err error) {
	configDir := filepath.Dir(sp.nginxConf)
	testDir := filepath.Join(os.TempDir(), "nginx")

	//把测试文件夹全部删除
	if err = os.RemoveAll(testDir); err != nil {
		return
	}
	if err = files.CopyDir(configDir, testDir); err != nil {
		return
	}

	//如果不为空就写入并且测试
	if cfg != nil {
		if err = Write2Path(testDir, cfg); err != nil {
			return
		}
	}

	for _, hock := range hocks {
		if err = hock(testDir); err != nil {
			return
		}
	}
	conf := filepath.Join(testDir, "nginx.conf")
	if err = cmds.CmdRun("nginx", "-c", conf, "-t"); err != nil {
		logger.WithError(err).Warn("测试")
		//去除测试目录内容
		err = fmt.Errorf(strings.ReplaceAll(err.Error(), testDir, "<storage>"))
	}
	return
}

func (sp *localDaemon) Stop() error {
	logger.Info("stop nginx daemon")
	if sp.daemon != nil {
		return sp.daemon.Process.Signal(syscall.SIGQUIT)
	} else {
		return cmds.CmdRun(sp.bin, sp.params("-s", "quit")...)
	}
}
