package nginx_test

import (
	"encoding/json"
	"fmt"
	"github.com/ihaiker/aginx/lego"
	"github.com/ihaiker/aginx/nginx"
	fileStorage "github.com/ihaiker/aginx/storage/file"
	"testing"
)

func TestService_List(t *testing.T) {
	engine := fileStorage.New(nginx.MustConf())
	legoStorage, _ := lego.NewManager(engine)
	client, _ := nginx.NewClient("", engine, legoStorage, nil)
	service := nginx.NewService(client)
	ss := service.ListService()
	for _, s := range ss {
		bs, _ := json.MarshalIndent(s, "", "\t")
		fmt.Println(string(bs))
	}
}
