package util

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

/**
 * 拷贝文件夹,同时拷贝文件夹中的文件
 * @param srcPath  		需要拷贝的文件夹路径
 * @param destPath		拷贝到的位置
 */
func CopyDir(srcPath string, destPath string) error {
	if srcInfo, err := os.Stat(srcPath); err != nil {
		return err
	} else if !srcInfo.IsDir() {
		return errors.New("srcPath is not a correct directory！")
	}

	if destInfo, err := os.Stat(destPath); err != nil && !os.IsNotExist(err) {
		return err
	} else if err == nil && !destInfo.IsDir() {
		return errors.New("destInfo is not a correct directory！")
	}

	err := filepath.Walk(srcPath, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if !f.IsDir() {
			path := strings.Replace(path, "\\", "/", -1)
			destNewPath := strings.Replace(path, srcPath, destPath, -1)
			if _, err := copyFile(path, destNewPath); err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func copyFile(src, dest string) (w int64, err error) {
	srcFile, err := os.Open(src)
	if err != nil {
		return
	}
	defer func() { _ = srcFile.Close() }()

	if err = os.MkdirAll(filepath.Dir(dest), os.ModePerm); err != nil {
		return
	}

	dstFile, err := os.Create(dest)
	if err != nil {
		return
	}
	defer func() { _ = dstFile.Close() }()

	return io.Copy(dstFile, srcFile)
}
