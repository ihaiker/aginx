package client

import (
	"fmt"
	"github.com/ihaiker/aginx/v2/api"
	"github.com/ihaiker/aginx/v2/core/nginx"
	"github.com/ihaiker/aginx/v2/core/nginx/config"
	"github.com/ihaiker/aginx/v2/core/nginx/query"
	"github.com/ihaiker/aginx/v2/core/util/errors"
	"github.com/ihaiker/aginx/v2/plugins/storage"
	"strings"
)

type clientDirective struct {
	engine storage.Plugin
	daemon nginx.Daemon
}

func (cd *clientDirective) Batch(batchs []*api.DirectiveBatch) error {
	if len(batchs) == 0 {
		return nil
	}

	conf, err := nginx.Configuration(cd.engine)
	if err != nil {
		return err
	}

	for _, one := range batchs {

		switch one.Type {
		case api.BatchAdd:
			{
				paths, err := query.Selects(conf, one.Queries...)
				if err != nil {
					return err
				}
				for _, directive := range one.Directives {
					logger.Infof("add (%s) in (%s)", directive.Name, strings.Join(one.Queries, " "))
				}
				for _, directive := range paths {
					directive.Body = append(directive.Body, one.Directives...)
				}
			}
		case api.BatchModify:
			{
				if len(one.Directives) == 0 {
					return fmt.Errorf("修改只是允许包含一项")
				}

				paths, err := query.Selects(conf, one.Queries...)
				if err != nil {
					return err
				}
				logger.Infof("modify in (%s)", strings.Join(one.Queries, " "))

				directive := one.Directives[0]
				for _, selectDirective := range paths {
					selectDirective.Name = directive.Name
					selectDirective.Args = directive.Args
					selectDirective.Body = directive.Body
				}
			}
		case api.BatchDelete:
			{
				queries := one.Queries

				if len(queries) == 0 {
					return fmt.Errorf("无法确认删除路径")
				}

				//前面定位，最后一个是删除对象
				finder := queries[0 : len(queries)-1]
				directives, err := query.Selects(conf, finder...)
				if err != nil || len(directives) == 0 { //查找错误或者没有找到直接返回
					return err
				}

				//删除对象描述
				delExpr, err := query.Lexer(queries[len(queries)-1])
				if err != nil {
					return err
				}

				logger.Infof("delete in (%s)", strings.Join(one.Queries, " "))

				for _, directive := range directives {
					//搜索需要删除的的索引
					deleteDirectiveIdx := make([]int, 0)
					for i, body := range directive.Body {
						if delExpr.Match(body) {
							deleteDirectiveIdx = append(deleteDirectiveIdx, i)
						}
					}
					//删除内容
					for i := len(deleteDirectiveIdx) - 1; i >= 0; i-- {
						idx := deleteDirectiveIdx[i]
						directive.Body = append(directive.Body[:idx], directive.Body[idx+1:]...)
					}
				}
			}
		}
	}
	return cd.testAndReload(conf)
}

func (cd *clientDirective) Select(queries ...string) ([]*config.Directive, error) {
	conf, err := nginx.Configuration(cd.engine)
	if err != nil {
		return nil, errors.Wrap(err, "解析配置文件")
	}
	if len(queries) == 0 {
		return conf.Body, nil
	}
	return query.Selects(conf, queries...)
}

func (cd *clientDirective) testAndReload(conf *config.Configuration) error {
	err := cd.daemon.Test(conf)
	if err == nil {
		//写入内容
		if err = nginx.Write2Storage(conf, cd.engine); err != nil {
			return err
		}
		return cd.daemon.Reload()
	}
	return err
}

func (cd *clientDirective) Add(queries []string, addDirectives ...*config.Directive) error {
	return cd.Batch([]*api.DirectiveBatch{
		{
			Type:       api.BatchAdd,
			Queries:    queries,
			Directives: addDirectives,
		},
	})
}

func (cd *clientDirective) Delete(queries ...string) error {
	return cd.Batch([]*api.DirectiveBatch{
		{
			Type:    api.BatchDelete,
			Queries: queries,
		},
	})
}

func (cd *clientDirective) Modify(queries []string, directive *config.Directive) error {
	return cd.Batch([]*api.DirectiveBatch{
		{
			Type:       api.BatchModify,
			Queries:    queries,
			Directives: []*config.Directive{directive},
		},
	})
}
