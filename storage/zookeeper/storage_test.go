package zookeeper

import (
	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&nested.Formatter{
		TimestampFormat: "2006-01-02 15:04:05.000",
		FieldsOrder:     []string{"engine"},
	})
	logrus.SetOutput(os.Stdout)
}

func TestStore(t *testing.T) {
	api, err := New("127.0.0.1:2181", "aginx", "", "")
	assert.Nil(t, err)

	err = api.Store("nginx.conf", []byte("zookeeper configuration "+time.Now().Format(time.RFC3339)))
	assert.Nil(t, err)
}

func TestSearch(t *testing.T) {
	api, err := New("127.0.0.1:2181", "aginx", "", "")
	assert.Nil(t, err)

	files, err := api.Search("*")
	assert.Nil(t, err)

	for _, file := range files {
		t.Log(file.Name)
	}
}

func TestStart(t *testing.T) {
	api, err := New("127.0.0.1:2181", "aginx", "", "")
	assert.Nil(t, err)

	err = api.Start()
	assert.Nil(t, err)
}
