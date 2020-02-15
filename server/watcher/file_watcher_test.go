package watcher

import (
	"github.com/ihaiker/aginx/logs"
	ignore2 "github.com/ihaiker/aginx/server/ignore"
	"github.com/ihaiker/aginx/storage"
	"github.com/ihaiker/aginx/storage/consul"
	"github.com/ihaiker/aginx/storage/etcd"
	"github.com/ihaiker/aginx/storage/file"
	"github.com/ihaiker/aginx/storage/zookeeper"
	"github.com/ihaiker/aginx/util"
	"github.com/sirupsen/logrus"
	"net/url"
	"testing"
)

func init() {
	logs.SetLevel(logrus.DebugLevel)
}
func TestWatcher(t *testing.T) {
	fw := new(FileWatcher)
	fw.ignore = ignore2.Cluster()

	config := "consul://127.0.0.1:8500/aginx"
	//config := "zk://127.0.0.1:2182/aginx"
	//config := "etcd://127.0.0.1:2379/aginx"

	fw.engine = getEngine(config, fw.ignore)
	if s, match := fw.engine.(util.Service); match {
		_ = s.Start()
	}
	if err := fw.Start(); err != nil {
		t.Fatal(err)
	}
}

func getEngine(cfgStr string, ignore ignore2.Ignore) (engine storage.Engine) {
	config, _ := url.Parse(cfgStr)
	switch config.Scheme {
	case "consul":
		engine, _ = consul.New(config, ignore)
	case "zk":
		engine, _ = zookeeper.New(config, ignore)
	case "etcd":
		engine, _ = etcd.New(config, ignore)
	default:
		engine, _ = file.System()
	}
	return engine
}

func TestF(t *testing.T) {
	engine := getEngine("etcd", ignore2.Empty())
	matchs, _ := engine.Search("*", "*/*", "*/*/*", "*/*/*/*")
	for _, match := range matchs {
		t.Log(match)
	}
}
