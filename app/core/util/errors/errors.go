package errors

import (
	"errors"
	"fmt"
	"net"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var (
	ErrNotFound            = Wrap(errors.New("not found"), "未发现")
	ErrRootCannotBeDeleted = Wrap(errors.New("root cannot be deleted"), "不可删除根节点")
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

type WrapError struct {
	Err     error
	Message string
}

func (w WrapError) Error() string {
	if w.Err == nil {
		return w.Message
	} else if w.Message == "" {
		return w.Err.Error()
	}
	return fmt.Sprintf("%s: %s", w.Message, w.Err)
}

func New(format string, object ...interface{}) error {
	return &WrapError{
		Err: fmt.Errorf(format, object...),
	}
}

func Wrap(err error, message string) error {
	if _, match := err.(*WrapError); match {
		return err
	} else {
		return &WrapError{
			Err: err, Message: message,
		}
	}
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

func Assert(check bool, format string, args ...interface{}) {
	if !check {
		panic(&WrapError{Message: fmt.Sprintf(format, args...)})
	}
}

//如果不为空，使用msg panic错误，
func PanicMessage(err error, msg string) {
	if err != nil {
		panic(Wrap(err, msg))
	}
}

func PanicIfError(err error) {
	_, file, line, _ := runtime.Caller(1)
	PanicMessage(err, fmt.Sprintf("%s:%d", file, line))
}

func Panic(err error) {
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

func Stack() string {
	stackBuf := make([]uintptr, 50)
	length := runtime.Callers(3, stackBuf[:])
	stack := stackBuf[:length]
	trace := ""
	frames := runtime.CallersFrames(stack)
	for {
		frame, more := frames.Next()
		if strings.HasSuffix(frame.File, "errors/errors.go") ||
			strings.HasSuffix(frame.File, "/src/runtime/panic.go") ||
			strings.HasSuffix(frame.File, "/testing/testing.go") ||
			frame.Function == "runtime.goexit" || frame.Function == "" {

		} else if strings.HasPrefix(frame.Function, "github.com/ihaiker/aginx") {
			trace = trace + fmt.Sprintf("\t%s:%d %s\n", frame.File, frame.Line, filepath.Base(frame.Function))
		}

		if !more {
			break
		}
	}
	return trace
}

var (
	ErrTimeout = errors.New("timeout")
)

func Async(d time.Duration, fn func() error) (err error) {
	errChan := make(chan error)
	go func() {
		defer Catch()
		errChan <- fn()
	}()

	select {
	case err = <-errChan:
	case <-time.After(d):
		err = ErrTimeout
	}
	close(errChan)
	return
}

func IsNotFound(err error) bool {
	return err != nil && strings.Contains(err.Error(), "not found")
}
