package files

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

func DiffWriteFile(path string, content []byte) (write bool, err error) {
	var bs []byte
	if bs, err = ioutil.ReadFile(path); err == nil {
		if bytes.Equal(content, bs) {
			return
		}
	}
	err = WriteFile(path, content)
	write = true
	return
}

func WriteFile(path string, content []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0777); err != nil {
		return err
	}
	return ioutil.WriteFile(path, content, 0666)
}

/**
 * 拷贝文件夹,同时拷贝文件夹中的文件
 * @param srcPath  		需要拷贝的文件夹路径
 * @param destPath		拷贝到的位置，
 */
func CopyDir(srcPath string, destPath string) error {
	if !IsDir(srcPath) {
		return errors.New("srcPath is not a directory or not exists！")
	}

	if Exists(destPath) && !IsDir(destPath) {
		return errors.New("destInfo is not a directory or not exists！")
	}
	if err := os.MkdirAll(destPath, 0777); err != nil {
		return err
	}
	if files, err := Walk(srcPath); err != nil {
		return err
	} else {
		for _, file := range files {
			srcRelPath, _ := filepath.Rel(srcPath, file.Name)
			destName := filepath.Join(destPath, srcRelPath)
			if _, err := copyFile(file.Name, destName); err != nil {
				return err
			}
		}
	}
	return nil
}

func copyFile(src, dest string) (w int64, err error) {
	srcFile, err := os.Open(src)
	if err != nil {
		return
	}
	defer func() { _ = srcFile.Close() }()

	if err = os.MkdirAll(filepath.Dir(dest), 0777); err != nil {
		return
	}

	dstFile, err := os.Create(dest)
	if err != nil {
		return
	}
	defer func() { _ = dstFile.Close() }()

	return io.Copy(dstFile, srcFile)
}

// 判断所给路径文件/文件夹是否存在
func Exists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func IsDir(path string) bool {
	if s, err := os.Stat(path); err == nil && s.IsDir() {
		return true
	}
	return false
}
