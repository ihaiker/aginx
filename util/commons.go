package util

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
)

//Try handler(err)
func Try(fun func(), handler ...func(error)) {
	defer Catch(handler...)
	fun()
}

//Try handler(err) and finally
func TryFinally(fun func(), handler func(error), finallyFn func()) {
	defer func() {
		if finallyFn != nil {
			finallyFn()
		}
	}()
	Try(fun, handler)
}

func CatchError(err error) {
	Catch(func(e error) {
		err = e
	})
}

func Catch(fns ...func(error)) {
	if r := recover(); r != nil && len(fns) > 0 {
		if err, match := r.(error); match {
			for _, fn := range fns {
				fn(err)
			}
		} else {
			err := fmt.Errorf("%v", r)
			for _, fn := range fns {
				fn(err)
			}
		}
	}
}

//如果不为空，使用msg panic错误，
func PanicMessage(err interface{}, msg string) {
	if err != nil {
		panic(fmt.Errorf("%s : %v", msg, err))
	}
}

func PanicIfError(err interface{}) {
	if err != nil {
		panic(err)
	}
}

func RandomPort() (int, error) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	port := listener.Addr().(*net.TCPAddr).Port
	return port, nil
}

type NameReader struct {
	io.Reader
	Name string
}

func NamedReader(rd io.Reader, name string) *NameReader {
	return &NameReader{Reader: rd, Name: name}
}

func (nr *NameReader) String() string {
	return fmt.Sprintf("file(%s)", nr.Name)
}

func WriterFile(path string, content []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}
	return ioutil.WriteFile(path, content, 0666)
}
