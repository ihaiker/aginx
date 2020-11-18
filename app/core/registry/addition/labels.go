package addition

import (
	"github.com/ihaiker/aginx/v2/api"
	"github.com/ihaiker/aginx/v2/core/logs"
	"github.com/ihaiker/aginx/v2/core/nginx/config"
	"github.com/ihaiker/aginx/v2/core/util"
	"time"
)

var logger = logs.New("registry")

/*
	给特定的额服务附加labels，做到0侵入第三方服务.
	默认是从注册管理器的meta中获取数据，但是这样的侵入性太强，
	所以使用这种方式来确定新的labels这样可以做到零侵入

配置方式一：名称配置
	serverName {
	      aginx.domain.0: http://test.aginx.io
	      aginx.domain.1: http://t2.aginx.io
	 }
配置方式二：标签选择
	aginx.labels service=consul {
		aginx.domain.0: http://consul.aginx.io;
	}
*/
type LabelFinder func(name string, meta map[string]string) map[string]string
type tagSelect func(userLabel map[string]string) (additionalLabel map[string]string, match bool)

type cache struct {
	finder     LabelFinder
	expireTime time.Time
}

var cachies = map[string]cache{}

func Load(aginx api.Aginx, path string) LabelFinder {
	//不能实时加载，不太合适
	if c, has := cachies[path]; has {
		if c.expireTime.After(time.Now()) {
			return c.finder
		}
	}
	logger.Debug("加载label finder ", path)
	cachies[path] = cache{
		finder:     aginxLoad(aginx, path),
		expireTime: time.Now().Add(time.Minute * 5), //5分钟后过期
	}
	return cachies[path].finder
}

func aginxLoad(aginx api.Aginx, path string) LabelFinder {
	confFile, err := aginx.Files().Get(path)
	if err == nil {
		return load(confFile.Content)
	}
	return func(name string, meta map[string]string) map[string]string {
		return map[string]string{}
	}
}

func makeTagSelect(tags, labels map[string]string) tagSelect {
	return func(userLabel map[string]string) (map[string]string, bool) {
		if userLabel == nil {
			return nil, false
		}
		for name, value := range tags {
			if labelValue, has := userLabel[name]; !has || value != labelValue {
				return nil, false
			}
		}
		return labels, true
	}
}

func load(content []byte) LabelFinder {
	opt := &config.Options{Delimiter: true, RemoveBrackets: true, RemoveAnnotation: true}

	namedLabels := make(map[string]map[string]string)
	tagsLabels := make([]tagSelect, 0)

	if conf, err := config.ParseWith("labels.conf", content, opt); err == nil {
		for _, d := range conf.Body {
			if d.Name == "aginx.labels" { //labels选择器
				if len(d.Args) == 0 || len(d.Body) == 0 {
					logger.Warn("labels.conf 配置错误：", d.Name)
					continue //错误配置不管了
				}
				tags := map[string]string{}
				for _, arg := range d.Args {
					name, value := util.Split2(arg, "=")
					tags[name] = value
				}
				labels := map[string]string{}
				for _, label := range d.Body {
					labels[label.Name] = label.Args[0]
				}
				tagsLabels = append(tagsLabels, makeTagSelect(tags, labels))
			} else {
				namedLabels[d.Name] = map[string]string{}
				for _, label := range d.Body {
					namedLabels[d.Name][label.Name] = label.Args[0]
				}
			}
		}
	} else {
		logger.WithError(err).Warn("解析labels.conf配置文件异常")
	}

	return func(name string, meta map[string]string) map[string]string {
		if labels, has := namedLabels[name]; has {
			return labels
		}
		for _, tagsLabel := range tagsLabels {
			if labels, match := tagsLabel(meta); match {
				return labels
			}
		}
		return meta
	}
}
