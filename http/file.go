package http

import (
	"bytes"
	"fmt"
	"github.com/ihaiker/aginx/nginx"
	ngx "github.com/ihaiker/aginx/nginx/config"
	"github.com/ihaiker/aginx/plugins"
	"github.com/ihaiker/aginx/util"
	"github.com/kataras/iris/v12"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type fileController struct {
	engine  plugins.StorageEngine
	process *nginx.Process
}

func (as *fileController) Search(queries []string) map[string]string {
	files := make(map[string]string, 0)
	cfgFiles, err := as.engine.Search(queries...)
	util.PanicIfError(err)
	for _, cfgFile := range cfgFiles {
		name := cfgFile.Name
		files[name] = string(cfgFile.Content)
	}
	return files
}

//解析上传文件，
//  1、格式化文件，
//  2、检查执行Test方法（Test方法会全部复制文件配置目录）
func (as *fileController) readFile(ctx iris.Context) *ngx.Configuration {
	filePath := ctx.FormValue("path")
	if filePath == "" || strings.HasPrefix(filePath, "/") {
		panic("path must be relative")
	}

	file, _, err := ctx.FormFile("file")
	util.PanicIfError(err)
	defer func() { _ = file.Close() }()

	out := bytes.NewBuffer(make([]byte, 0))
	_, err = io.Copy(out, file)
	util.PanicIfError(err)
	cfg, err := ngx.ParseWith(filePath, out.Bytes())
	util.PanicMessage(err, "parse error")
	return cfg
}

//是否配置文件已经在include的范围内
func (as *fileController) needAddInclude(client *nginx.Client, filePath string) bool {
	if includes, err := client.Select("http", "include"); err == nil {
		for _, include := range includes {
			if matched, _ := filepath.Match(include.Args[0], filePath); matched {
				return false
			}
		}
	}
	return true
}

func (as *fileController) New(ctx iris.Context, client *nginx.Client) int {
	cfg := as.readFile(ctx)
	filePath := cfg.Name
	contentBytes := cfg.BodyBytes()

	//如果是配置文件需要测试是否可用
	if filepath.Ext(filePath) == ".conf" {
		needAddInclude := as.needAddInclude(client, filePath)

		if needAddInclude {
			_ = client.Add(nginx.Queries("http"), ngx.NewDirective("include", filePath))
		}
		util.PanicIfError(as.process.Test(client.Configuration(), func(testDir string) error {
			path := filepath.Join(testDir, filePath)
			return util.WriteFile(path, contentBytes)
		}))
		if needAddInclude {
			_ = client.Delete("http", fmt.Sprintf("include('%s')", filePath))
		}
	}
	util.PanicIfError(as.engine.Put(filePath, contentBytes))
	util.PanicIfError(as.process.Reload())
	return iris.StatusNoContent
}

func (as *fileController) Remove(ctx iris.Context, client *nginx.Client) int {
	file := ctx.URLParam("file")
	if strings.HasPrefix(file, "/") {
		panic("Get path must be relative")
	}
	util.PanicIfError(as.process.Test(client.Configuration(), func(testDir string) error {
		path := filepath.Join(testDir, file)
		return os.Remove(path)
	}))
	util.PanicMessage(as.engine.Remove(file), "remove file error")
	util.PanicIfError(as.process.Reload())
	return iris.StatusNoContent
}
