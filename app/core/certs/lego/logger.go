package lego

import (
	"fmt"
	"github.com/go-acme/lego/v3/log"
	"github.com/ihaiker/aginx/v2/core/logs"
)

var logger = logs.New("lego")

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
	logger.Print(args...)
}
func (this *Stdlout) Println(args ...interface{}) {
	logger.Println(args...)
}
func (this *Stdlout) Printf(format string, args ...interface{}) {
	logger.Printf(format, args...)
}

func init() {
	log.Logger = new(Stdlout)
}
