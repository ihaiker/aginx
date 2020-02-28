package http

import (
	"bytes"
	"github.com/ihaiker/aginx/nginx"
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
	if filepath.Ext(filePath) == "conf" {
		_ = client.Add(nginx.Queries("http"), nginx.NewDirective("include", filePath))
		util.PanicIfError(as.process.Test(client.Configuration(), func(testDir string) error {
			path := filepath.Join(testDir, filePath)
			return util.WriteFile(path, bodys)
		}))
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
