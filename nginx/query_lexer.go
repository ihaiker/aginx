package nginx

import (
	"github.com/alecthomas/participle"
)

type QueryArg struct {
	Comparison string `[@("!" | "@" | "^" | "$")]`
	Value      string `@(String|RawString|Ident)`
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
	Name       string `@Ident )`

	All string ` | @"*" )`

	Args *QueryArgs `["(" [@@] ")"]`
}

type QueryChildren struct {
	Directive *QueryDirective  `( @@`
	Group     *QueryChildArray `| "[" @@ "]" )`
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
	Directive *QueryDirective  `@@`
	Children  []*QueryChildren `("." @@)*`
}

func Parser(str string) (expr *Expression, err error) {
	expr = &Expression{}
	parser := participle.MustBuild(expr)
	err = parser.ParseString(str, expr)
	return
}
