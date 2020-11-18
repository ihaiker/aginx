package services

import (
	"github.com/ihaiker/aginx/v2/core/util/errors"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Service interface {
	Start() error
	Stop() error
}

type manager struct {
	services []Service
}

func Manager() *manager {
	return &manager{services: make([]Service, 0)}
}

func (d *manager) Add(service ...interface{}) *manager {
	for _, server := range service {
		if ss, match := server.(Service); match {
			d.services = append(d.services, ss)
		}
	}
	return d
}

func (d *manager) AddStart(fns ...func() error) *manager {
	for _, fn := range fns {
		d.Add(&funcService{StartFn: fn})
	}
	return d
}

func (d *manager) AddStop(fns ...func() error) *manager {
	for _, fn := range fns {
		d.Add(&funcService{StopFn: fn})
	}
	return d
}

func (d *manager) Stop() error {
	for i := len(d.services) - 1; i >= 0; i-- {
		_ = d.services[i].Stop()
	}
	return nil
}

func (d *manager) await() error {
	C := make(chan os.Signal)
	signal.Notify(C, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	for _ = range C {
		err := errors.Async(time.Second*7, d.Stop)
		if err == errors.ErrTimeout {
			os.Exit(1)
			return nil
		} else {
			return err
		}
	}
	return nil
}

func (d *manager) Start() error {
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

func Start(ob interface{}) error {
	if ob != nil {
		if sv, match := ob.(Service); match {
			return sv.Start()
		}
	}
	return nil
}

func Stop(ob interface{}) error {
	if ob != nil {
		if sv, match := ob.(Service); match {
			return sv.Stop()
		}
	}
	return nil
}
