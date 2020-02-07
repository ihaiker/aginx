package lego

import (
	"fmt"
	"github.com/go-acme/lego/v3/log"
	"github.com/sirupsen/logrus"
)

type Stdlout struct {
}

func (this *Stdlout) Fatal(args ...interface{}) {
	panic(fmt.Sprint(args...))
}
func (this *Stdlout) Fatalln(args ...interface{}) {
	panic(fmt.Sprintln(args...))
}
func (this *Stdlout) Fatalf(format string, args ...interface{}) {
	panic(fmt.Sprintf(format, args...))
}
func (this *Stdlout) Print(args ...interface{}) {
	logrus.Print(args...)
}
func (this *Stdlout) Println(args ...interface{}) {
	logrus.Println(args...)
}
func (this *Stdlout) Printf(format string, args ...interface{}) {
	logrus.Printf(format, args...)
}

func init() {
	log.Logger = new(Stdlout)
}
