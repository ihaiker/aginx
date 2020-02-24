package util

import (
	"errors"
	"fmt"
	"net"
)

func Safe(fn func()) (err error) {
	Try(fn, func(e error) {
		err = e
	})
	return
}

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

func AssertTrue(check bool, msg string) {
	if !check {
		panic(errors.New(msg))
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
