package main

import (
	"fmt"
	"github.com/ihaiker/aginx/v2/cmd"
	"math/rand"
	"os"
	"runtime"
	"time"
)

var (
	VERSION        = "v0.0.1"
	BUILD_TIME     = "2012-12-12 12:12:12"
	GITLOG_VERSION = "0000"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	rand.Seed(time.Now().Unix())
	if err := cmd.Execute(VERSION, BUILD_TIME, GITLOG_VERSION); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
