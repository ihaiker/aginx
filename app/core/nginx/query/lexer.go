package query

import (
	"github.com/alecthomas/participle"
)

type QueryArg struct {
	Comparison string `[@("!" | "@" | "^" | "$")]`
	Value      string `@(String|RawString|Ident|"_"|"*")`
}

type QueryArgAddition struct {
	Operator string    `@("&" | "|")`
	Arg      *QueryArg `@@`
}

type QueryArgs struct {
	Arg  *QueryArg           `@@`
	Next []*QueryArgAddition `{ @@ }`
}

type QueryDirective struct {
	Comparison string `( ( [@("!" | "@" | "^" | "$")]`
	Name       string `@Ident )` //指令名称

	All string ` | @"*" )`

	Args *QueryArgs `["(" [@@] ")"]`
}

type QueryChildren struct {
	//server.[server_name('_') & listen('8081' | '8080')]
	//后置指令一个，
	Directive *QueryDirective `( @@`
	//后置指令多个
	Group *QueryChildArray `| "[" @@ "]" )`
}

type QueryChildArray struct {
	First *QueryDirective          `@@`
	Next  []*QueryChildrenAddition `@@`
}

type QueryChildrenAddition struct {
	Operator string          `@("&" | "|")`
	Next     *QueryDirective `@@`
}

type Expression struct {
	Directive *QueryDirective  `@@`        //前置指令 例如：http
	Children  []*QueryChildren `("." @@)*` //后置指令 例如 server.server_name 中的 .server_name
}

func Lexer(str string) (expr *Expression, err error) {
	expr = &Expression{}
	parser := participle.MustBuild(expr)
	err = parser.ParseString(str, expr)
	return
}
