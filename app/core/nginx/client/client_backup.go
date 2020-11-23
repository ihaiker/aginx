package client

import (
	"archive/zip"
	"bytes"
	"fmt"
	"github.com/ihaiker/aginx/v2/api"
	"github.com/ihaiker/aginx/v2/core/nginx"
	"github.com/ihaiker/aginx/v2/core/util/errors"
	"github.com/ihaiker/aginx/v2/core/util/files"
	"github.com/ihaiker/aginx/v2/plugins/storage"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type clientBackup struct {
	engine   storage.Plugin
	daemon   nginx.Daemon
	dir      string
	limit    int
	dayLimit int
}

func (c *clientBackup) List() ([]*api.BackupFile, error) {
	names := make([]*api.BackupFile, 0)
	if !files.Exists(c.dir) {
		return names, nil
	}
	fis, err := ioutil.ReadDir(c.dir)
	if err != nil {
		return nil, err
	}
	for _, fi := range fis {
		name := fi.Name()
		if !strings.HasSuffix(name, ".zip") {
			logger.Warn("备份文件夹下存不确定文件：", name)
			continue
		}
		backupDate := strings.Replace(fi.Name(), ".zip", "", 1)
		if rc, err := zip.OpenReader(filepath.Join(c.dir, fi.Name())); err != nil {
			return names, err
		} else {
			names = append([]*api.BackupFile{{Comment: rc.Comment, Name: backupDate}}, names...)
			_ = rc.Close()
		}
	}
	return names, nil
}

func (c *clientBackup) Delete(dateString string) error {
	path := filepath.Join(c.dir, fmt.Sprintf("%s.zip", dateString))
	if !files.Exists(path) {
		return errors.Wrap(errors.ErrNotFound, "备份未发现")
	}
	return os.Remove(path)
}

func (c *clientBackup) Backup(comment string) (*api.BackupFile, error) {
	fs, err := c.engine.Search()
	if err != nil {
		return nil, err
	}

	name := time.Now().Format("20060102T150405")

	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)
	_ = zipWriter.SetComment(comment)
	for _, file := range fs {
		zipFile, err := zipWriter.Create(file.Name)
		if err != nil {
			return nil, err
		}
		if _, err = zipFile.Write([]byte(file.Content)); err != nil {
			return nil, err
		}
		logger.Debug("备份文件：", file.Name, " to ", name)
	}
	// Make sure to check the error on Close.
	if err = zipWriter.Close(); err != nil {
		return nil, err
	}

	path := filepath.Join(c.dir, fmt.Sprintf("%s.zip", name))
	if err = os.MkdirAll(c.dir, 0777); err != nil {
		return nil, err
	}
	if err = ioutil.WriteFile(path, buf.Bytes(), 0666); err == nil {
		logger.Info("备份成功：", name)
	}
	return &api.BackupFile{Comment: comment, Name: name}, err
}

func (c *clientBackup) Rollback(dateString string) error {
	path := filepath.Join(c.dir, fmt.Sprintf("%s.zip", dateString))
	if !files.Exists(path) {
		return errors.Wrap(errors.ErrNotFound, "备份未发现")
	}
	zipReader, err := zip.OpenReader(path)
	if err != nil {
		return err
	}
	defer zipReader.Close()

	for _, f := range zipReader.File {
		if rc, err := f.Open(); err == nil {
			logger.Debug("恢复文件：", f.FileHeader.Name)
			bs := bytes.NewBuffer([]byte{})
			if _, err = io.Copy(bs, rc); err != nil {
				return err
			}
			if err = c.engine.Put(f.FileHeader.Name, bs.Bytes()); err != nil {
				return err
			}
			_ = rc.Close()
		}
	}
	return c.daemon.Reload()
}
