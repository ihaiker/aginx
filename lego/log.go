package lego

import "fmt"

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
	fmt.Print(args...)
}
func (this *Stdlout) Println(args ...interface{}) {
	fmt.Println(args...)
}
func (this *Stdlout) Printf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}
