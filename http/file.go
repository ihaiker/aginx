package http

import (
	"bytes"
	"github.com/ihaiker/aginx/plugins"
	"github.com/ihaiker/aginx/util"
	"github.com/kataras/iris/v12"
	"io"
	"strings"
)

type fileController struct {
	engine plugins.StorageEngine
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

func (as *fileController) New(ctx iris.Context) int {
	filePath := ctx.FormValue("path")

	file, _, err := ctx.FormFile("file")
	util.PanicIfError(err)
	defer func() { _ = file.Close() }()

	out := bytes.NewBuffer(make([]byte, 0))
	_, err = io.Copy(out, file)
	util.PanicIfError(err)

	err = as.engine.Put(filePath, out.Bytes())
	util.PanicIfError(err)

	return iris.StatusNoContent
}

func (as *fileController) Remove(ctx iris.Context) int {
	file := ctx.URLParam("file")
	if strings.HasPrefix(file, "/") {
		panic("Get path must be relative")
	}
	util.PanicMessage(as.engine.Remove(file), "remove file error")
	return iris.StatusNoContent
}
