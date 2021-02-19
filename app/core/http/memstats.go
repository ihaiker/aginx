package http

import (
	"github.com/ihaiker/aginx/v2/core/nginx"
	"github.com/kataras/iris/v12"
	"github.com/shirou/gopsutil/v3/process"
	"os"
	"time"
)

type processInfo struct {
	Time     string  `json:"time"`
	Mem      uint64  `json:"mem"`
	NginxMem uint64  `json:"nginxMem"`
	CPU      float64 `json:"cpu"`
	NginxCpu float64 `json:"nginxCpu"`
}

var infos []*processInfo

type memInfoController struct {
	daemon nginx.Daemon
}

func (c *memInfoController) start() {
	infos = make([]*processInfo, 0)
	go c.readMemStats()
}

func (c *memInfoController) readMemStats() {
	pid := int32(os.Getpid())
	var err error
	var mem *process.MemoryInfoStat
	var p *process.Process
	for {
		time.Sleep(time.Second)

		pi := &processInfo{
			Time: time.Now().Format("2006-01-02 15:04:05"),
		}
		//程序本身
		{
			p, err = process.NewProcess(pid)
			if err != nil {
				logger.Debug("获取进程信息错误：", err)
				continue
			}
			mem, err = p.MemoryInfo()
			if err != nil {
				logger.Debug("获取内存信息错误：", err)
				continue
			}
			pi.Mem = mem.RSS
			pi.CPU, err = p.CPUPercent()
			if err != nil {
				logger.Debug("获取CPU信息错误：", err)
				continue
			}
		}
		//nginx
		if c.daemon != nil {
			p, err = process.NewProcess(c.daemon.PID())
			if err != nil {
				logger.Debug("获取进程信息错误：", err)
				continue
			}
			mem, err = p.MemoryInfo()
			if err != nil {
				logger.Debug("获取内存信息错误：", err)
				continue
			}
			pi.NginxMem = mem.RSS
			pi.NginxCpu, err = p.CPUPercent()
			if err != nil {
				logger.Debug("获取CPU信息错误：", err)
				continue
			}
		}

		if len(infos) >= 60*30 {
			infos = infos[1:]
		}
		infos = append(infos, pi)
	}
}

func (c *memInfoController) memstats(ctx iris.Context) interface{} {
	limit := ctx.URLParamIntDefault("limit", 1)
	if limit == 1 {
		return infos[len(infos)-1]
	} else {
		return infos
	}
}
