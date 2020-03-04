package util

import (
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Service interface {
	Start() error
	Stop() error
}

type daemon struct {
	services []Service
}

func NewDaemon() *daemon {
	return &daemon{services: make([]Service, 0)}
}

func (d *daemon) Add(service ...Service) *daemon {
	d.services = append(d.services, service...)
	return d
}

func (d *daemon) AddStart(fns ...func() error) *daemon {
	for _, fn := range fns {
		d.Add(&funcService{StartFn: fn})
	}
	return d
}

func (d *daemon) AddStop(fns ...func() error) *daemon {
	for _, fn := range fns {
		d.Add(&funcService{StopFn: fn})
	}
	return d
}

func (d *daemon) Stop() error {
	for i := len(d.services) - 1; i >= 0; i-- {
		_ = d.services[i].Stop()
	}
	return nil
}

func (d *daemon) await() error {
	C := make(chan os.Signal)
	signal.Notify(C, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	for _ = range C {
		err := Async(time.Second*7, d.Stop)
		if err == ErrTimeout {
			os.Exit(1)
			return nil
		} else {
			return err
		}
	}
	return nil
}

func (d *daemon) Start() error {
	for idx, service := range d.services {
		if err := service.Start(); err != nil {
			for i := idx; i >= 0; i-- {
				_ = d.services[i].Stop()
			}
			return err
		}
	}
	return d.await()
}

type funcService struct {
	StartFn func() error
	StopFn  func() error
}

func (f *funcService) Start() error {
	if f.StartFn != nil {
		return f.StartFn()
	}
	return nil
}

func (f *funcService) Stop() error {
	if f.StopFn != nil {
		return f.StopFn()
	}
	return nil
}

func StartService(ob interface{}) error {
	if ob != nil {
		if sv, match := ob.(Service); match {
			return sv.Start()
		}
	}
	return nil
}

func StopService(ob interface{}) error {
	if ob != nil {
		if sv, match := ob.(Service); match {
			return sv.Stop()
		}
	}
	return nil
}
