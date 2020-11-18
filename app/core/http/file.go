package http

import (
	"bytes"
	"github.com/ihaiker/aginx/v2/api"
	"github.com/ihaiker/aginx/v2/core/util/errors"
	"github.com/ihaiker/aginx/v2/plugins/storage"
	"github.com/kataras/iris/v12"
	"io"
	"strings"
)

type fileController struct {
	aginx api.Aginx
}

func (as *fileController) New(ctx iris.Context) int {
	filePath := ctx.FormValue("path")
	errors.Assert(filePath != "" && !strings.HasPrefix(filePath, "/"), "文件路径错误")

	var bodyContext []byte
	file, _, err := ctx.FormFile("file")
	if err != nil && strings.Contains(err.Error(), "no such file") {
		bodyContext = []byte(ctx.FormValueDefault("fileContext", ""))
	} else {
		errors.PanicMessage(err, "获取上传文件")
		defer file.Close()
		out := bytes.NewBuffer(make([]byte, 0))
		_, err = io.Copy(out, file)
		errors.PanicMessage(err, "获得上传内容")
		bodyContext = out.Bytes()
	}
	errors.Assert(len(bodyContext) != 0, "文件内容不能为空")
	err = as.aginx.Files().NewWithContent(filePath, bodyContext)
	errors.Panic(err)
	return iris.StatusNoContent
}

func (as *fileController) Search(ctx iris.Context) []*storage.File {
	q := ctx.Request().URL.Query()["q"]
	files, err := as.aginx.Files().Search(q...)
	errors.Panic(err)
	return files
}

func (as *fileController) Get(ctx iris.Context) *storage.File {
	q := ctx.URLParam("q")
	file, err := as.aginx.Files().Get(q)
	errors.Panic(err)
	return file
}

func (as *fileController) Remove(ctx iris.Context) int {
	q := ctx.URLParam("q")
	err := as.aginx.Files().Remove(q)
	errors.Panic(err)
	return iris.StatusNoContent
}
