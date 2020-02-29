package dockerLabels

import (
	"encoding/json"
	"github.com/ihaiker/aginx/logs"
	"regexp"
	"strconv"
)

var logger = logs.New("register", "engine", "docker.labels")

var keyRegexp = regexp.MustCompile("aginx.domain(\\.(\\d+))?")
var valueRegexp = regexp.MustCompile("([a-zA-Z0-9-_\\.]*)(,(weight=(\\d+)))?(,(internal))?(,(ssl))?(,(virtual))?(,(nodes))?")

type label struct {
	Domain   string
	Port     int
	Weight   int  //服务器权重。在费Swarm节点下起作用
	AutoSSL  bool //自动生成证书文件
	Internal bool //使用内部地址
	Virtual  bool //虚拟VIP 只配置一个
	Nodes    bool //外部接口全节点配置，这里还可以筛选
}

type labels map[int]label

func (ls *labels) Has() bool {
	return len(*ls) > 0
}

func (ls *labels) String() string {
	bs, _ := json.Marshal(ls)
	return string(bs)
}

func findLabels(labs map[string]string, ignoreSwarmService bool) labels {
	lbs := labels{}
	if _, has := labs["com.docker.swarm.task.id"]; ignoreSwarmService && has {
		return lbs
	}
	for key, value := range labs {
		if keyRegexp.MatchString(key) && valueRegexp.MatchString(value) {
			domain := valueRegexp.FindStringSubmatch(value)
			port := keyRegexp.FindStringSubmatch(key)
			label := label{Domain: domain[1]}
			label.Weight, _ = strconv.Atoi(domain[4])
			label.Internal = domain[6] == "internal"
			label.Port, _ = strconv.Atoi(port[2])
			label.AutoSSL = domain[8] == "ssl"
			label.Virtual = domain[10] == "virtual"
			label.Nodes = domain[12] == "nodes"
			lbs[label.Port] = label
		}
	}
	return lbs
}
