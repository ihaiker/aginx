package config

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestCharIt(t *testing.T) {
	it := newCharIterator("/etc/nginx/nginx.conf")
	for {
		char, lineno, has := it.next()
		if !has {
			break
		}
		fmt.Printf("%3s %4d\n", char, lineno)
	}
}

func TestTokenIt(t *testing.T) {
	it := newTokenIterator("/etc/nginx/nginx.conf", nil)
	for {
		if token, line, has := it.next(); has {
			fmt.Printf("%-4d:   %s\n", line, token)
		} else {
			break
		}
	}
}

func TestParseConfig(t *testing.T) {
	cfg, err := Parse("/etc/nginx/nginx.conf", nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(cfg.Pretty(0))
}

func TestParseAginx(t *testing.T) {
	cfg, err := Parse("../../aginx.conf", nil)
	if err != nil {
		t.Fatal(err)
	}
	bs, _ := json.MarshalIndent(cfg, "", "\t")
	fmt.Println(string(bs))
}
