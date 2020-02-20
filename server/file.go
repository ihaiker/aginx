package server

import (
	"bytes"
	"github.com/ihaiker/aginx/storage"
	"github.com/ihaiker/aginx/util"
	"github.com/kataras/iris/v12"
	"io"
	"io/ioutil"
	"strings"
)

type fileController struct {
	engine storage.Engine
}

func (as *fileController) Search(queries []string) map[string]string {
	files := make(map[string]string, 0)
	readers, err := as.engine.Search(queries...)
	util.PanicIfError(err)
	for _, reader := range readers {
		bs, _ := ioutil.ReadAll(reader)
		name := reader.Name
		files[name] = string(bs)
	}
	return files
}

func (as *fileController) Remove(ctx iris.Context) int {
	file := ctx.URLParam("file")
	if strings.HasPrefix(file, "/") {
		panic("File path must be relative")
	}
	util.PanicMessage(as.engine.Remove(file), "remove file error")
	return iris.StatusNoContent
}

func (as *fileController) New(ctx iris.Context) int {
	filePath := ctx.FormValue("path")

	file, _, err := ctx.FormFile("file")
	util.PanicIfError(err)
	defer func() { _ = file.Close() }()

	out := bytes.NewBuffer(make([]byte, 0))
	_, err = io.Copy(out, file)
	util.PanicIfError(err)

	err = as.engine.Store(filePath, out.Bytes())
	util.PanicIfError(err)

	return iris.StatusNoContent
}
