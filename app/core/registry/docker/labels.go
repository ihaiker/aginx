package docker

import (
	"fmt"
	"github.com/ihaiker/aginx/v2/core/util/errors"
	"net/url"
	"strconv"
	"strings"
)

type label struct {
	Domain   string
	Port     int    //监听的端口
	Weight   int    //服务器权重。在费Swarm节点下起作用
	AutoSSL  bool   //自动生成证书文件
	Internal bool   //使用内部地址
	Networks string //网络优先选择地址
	Source   string //原文
	Template string //使用模板
	Provider string
}

func findLabels(labs map[string]string) ([]label, error) {
	//swarm服务的容器，不用监控
	if _, has := labs["com.docker.swarm.task.id"]; has {
		return nil, nil
	}

	labels := make([]label, 0)
	for key, value := range labs {
		if !strings.HasPrefix(key, "aginx.domain") {
			continue
		}
		tag, err := url.Parse(value)
		if err != nil || tag.Host == "" { //即使解析不出错，如果未提供Scheme前缀的也是同样host为空
			return nil, errors.New("标签定义错误：%s", value)
		}
		label := label{}
		label.Source = fmt.Sprintf("%s=%s", key, value)
		label.Domain = tag.Host
		label.AutoSSL = tag.Scheme == "https"

		if port := tag.Query().Get("port"); port != "" {
			if label.Port, err = strconv.Atoi(port); err != nil {
				return nil, err
			}
		}

		if weight := tag.Query().Get("weight"); weight != "" {
			if label.Weight, err = strconv.Atoi(weight); err != nil {
				return nil, err
			}
		}

		label.Networks = tag.Query().Get("networks")
		label.Internal = tag.Query().Get("internal") == "true"
		label.Template = tag.Query().Get("template")
		label.Provider = tag.Query().Get("provider")
		labels = append(labels, label)
	}
	return labels, nil
}
