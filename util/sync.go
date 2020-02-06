package util

import (
	"errors"
	"time"
)

var (
	ErrTimeout = errors.New("timeout")
)

func Async(d time.Duration, fn func() error) (err error) {
	errChan := make(chan error)
	go func() {
		defer func() {
			_ = recover()
		}()
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
