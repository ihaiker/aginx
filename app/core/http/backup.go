package http

import (
	"github.com/ihaiker/aginx/v2/api"
	"github.com/ihaiker/aginx/v2/core/util/errors"
	"github.com/kataras/iris/v12"
)

type backupController struct {
	aginx api.Aginx
}

func (bc *backupController) list() []*api.BackupFile {
	names, err := bc.aginx.Backup().List()
	errors.Panic(err)
	return names
}
func (bc *backupController) delete(ctx iris.Context) int {
	name := ctx.URLParam("name")
	err := bc.aginx.Backup().Delete(name)
	errors.Panic(err)
	return iris.StatusNoContent
}
func (bc *backupController) backup(ctx iris.Context) *api.BackupFile {
	comment := ctx.URLParam("comment")
	name, err := bc.aginx.Backup().Backup(comment)
	errors.Panic(err)
	return name
}

func (bc *backupController) rollback(ctx iris.Context) int {
	name := ctx.URLParam("name")
	err := bc.aginx.Backup().Rollback(name)
	errors.Panic(err)
	return iris.StatusNoContent
}
