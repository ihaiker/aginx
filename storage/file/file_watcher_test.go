package file

import (
	"fmt"
	"github.com/ihaiker/aginx/logs"
	"github.com/ihaiker/aginx/util"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func init() {
	logs.SetLevel(logrus.DebugLevel)
}
func TestWatcher(t *testing.T) {
	root, _ := os.Getwd()
	t.Log("root :", root)

	fw := NewFileWatcher(root)
	if err := fw.Start(); err != nil {
		t.Fatal(err)
	}
	gw := new(sync.WaitGroup)
	file := filepath.Join(root, "test.conf")
	gw.Add(1)
	go func() {
		defer gw.Done()
		{
			time.Sleep(time.Second)
			gw.Add(1)
			if err := ioutil.WriteFile(file, []byte("1"), os.ModePerm); err != nil {
				gw.Done()
			}

			time.Sleep(time.Second)

			gw.Add(1)
			if err := ioutil.WriteFile(file, []byte("2"), os.ModePerm); err != nil {
				gw.Done()
			}

			time.Sleep(time.Second)

			gw.Add(1)
			if err := os.Remove(file); err != nil {
				gw.Done()
			}
		}
	}()

	go func() {
		for event := range fw.Listener {
			fmt.Println(event.String())
			gw.Done()
		}
	}()

	err := util.Async(time.Second*5, func() error {
		gw.Wait()
		return nil
	})

	_ = fw.Stop()
	if err != nil {
		t.Fatal(err)
	}
}
