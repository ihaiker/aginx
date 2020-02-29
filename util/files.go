package util

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
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
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}
	return ioutil.WriteFile(path, content, 0666)
}

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
		destNewPath := strings.Replace(path, srcPath, destPath, -1)
		if f.Mode()&os.ModeSymlink == os.ModeSymlink {
			linkPath, _ := filepath.EvalSymlinks(path)
			if linkInfo, err := os.Stat(linkPath); err != nil {
				return err
			} else if linkInfo.IsDir() {
				relative, _ := filepath.Rel(srcPath, path)
				destNewPath = filepath.Join(destPath, relative)
				return CopyDir(linkPath, destNewPath)
			}
		}
		if !f.IsDir() {
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

//清楚空行
func CleanEmptyLine(content []byte) []byte {
	reader := bufio.NewReader(bytes.NewBuffer(content))
	out := bytes.NewBufferString("")
	for {
		if line, _, err := reader.ReadLine(); err == io.EOF {
			break
		} else if strings.TrimSpace(string(line)) != "" {
			out.Write(line)
			out.WriteRune('\n')
		}
	}
	return out.Bytes()
}
