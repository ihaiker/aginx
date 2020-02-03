package nginx

import (
	"os"
	"testing"
)

func TestAnalysis(t *testing.T) {
	r, err := os.OpenFile("../_test/nginx.nginx", os.O_RDONLY, os.ModeTemporary)
	if err != nil {
		t.Fatal(err)
	}
	conf, err := Analysis("../_test", r)
	if err != nil {
		t.Fatal(err)
	}
	//for _, directive := range *conf {
	//	t.Log(directive)
	//}
	t.Log(conf)
}

func TestNginx(t *testing.T) {
	conf, err := AnalysisNginx()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(conf)
}
