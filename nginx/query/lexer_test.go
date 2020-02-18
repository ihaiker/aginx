package query

import (
	"github.com/kr/pretty"
	"testing"
)

func TestLexer(t *testing.T) {
	//str := "http"
	//str := `server_name('ni.renzhen.la' | 'wo.renzhen.la')`
	//str := `server.server_name('ni.renzhen.la' | 'wo.renzhen.la')`
	//str := `@server.server_name('*.renzhen.la' | 'wo.renzhen.la').location("~")`
	//str := `server.[!server_name('*.renzhen.la' | 'wo.renzhen.la') & listen('80' | '443')]`
	//str := "*"
	//str := `server.server_name(^'www')`
	//str := `server.server_name($'www')`
	//str := `server.server_name()`
	//str := `server.server_name( name | age )`
	str := `server.server_name( 'name' | 'age' )`
	expr, err := Parser(str)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = pretty.Println(expr)
}
