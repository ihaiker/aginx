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

func StartService(ob interface{}) error {
	if sv, match := ob.(Service); match {
		return sv.Start()
	}
	return nil
}

func StopService(ob interface{}) error {
	if sv, match := ob.(Service); match {
		return sv.Stop()
	}
	return nil
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
