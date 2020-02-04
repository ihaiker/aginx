package nginx

import (
	"fmt"
	"os"
	"testing"
)

func TestAnalysis(t *testing.T) {
	r, err := os.OpenFile("../_test/nginx.conf", os.O_RDONLY, os.ModeTemporary)
	if err != nil {
		t.Fatal(err)
	}
	conf, err := Analysis("../_test", r)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(conf)
}

func TestNginx(t *testing.T) {
	conf, err := AnalysisNginx()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(conf.String())
}
