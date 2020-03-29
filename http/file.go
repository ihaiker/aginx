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

func (as *fileController) readFile(ctx iris.Context) []byte {
	file, _, err := ctx.FormFile("file")
	util.PanicIfError(err)
	defer func() { _ = file.Close() }()

	out := bytes.NewBuffer(make([]byte, 0))
	_, err = io.Copy(out, file)
	util.PanicIfError(err)
	return out.Bytes()
}

func (as *fileController) New(ctx iris.Context, client *nginx.Client) int {
	filePath := ctx.FormValue("path")
	if strings.HasPrefix(filePath, "/") {
		panic("path must be relative")
	}
	bodys := as.readFile(ctx)
	//如果是配置文件需要测试是否可用
	if filepath.Ext(filePath) == ".conf" {
		need := true
		if includes, err := client.Select("http", "include"); err == nil {
			for _, include := range includes {
				if matched, _ := filepath.Match(include.Args[0], filePath); matched {
					need = false
				}
			}
		}
		if need {
			_ = client.Add(nginx.Queries("http"), ngx.NewDirective("include", filePath))
		}
		util.PanicIfError(as.process.Test(client.Configuration(), func(testDir string) error {
			path := filepath.Join(testDir, filePath)
			return util.WriteFile(path, bodys)
		}))
		if need {
			_ = client.Delete("http", fmt.Sprintf("include('%s')", filePath))
		}
	}
	util.PanicIfError(as.engine.Put(filePath, bodys))
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
