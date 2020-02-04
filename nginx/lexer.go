package nginx

import (
	"github.com/alecthomas/participle"
)

type Arg struct {
	Comparison string `[@("!" | "@" | "^" | "$"]`
	Value      string `@(Ident|String|RawString)`
}

type ArgAddition struct {
	Operator string `@("&" | "|")`
	Arg      *Arg   `@@`
}

type Args struct {
	Arg  *Arg           `@@`
	Next []*ArgAddition `{ @@ }`
}

type DirectiveExpr struct {
	Comparison string `( ( [@("!" | "@" | "^" | "$")]`
	Name       string `@Ident )`

	All string ` | @"*" )`

	Args *Args `["(" @@ ")"]`
}

type Children struct {
	Directive *DirectiveExpr `( @@`
	Group     *ChildArray    `| "[" @@ "]" )`
}

type ChildArray struct {
	First *DirectiveExpr      `@@`
	Next  []*ChildrenAddition `@@`
}

type ChildrenAddition struct {
	Operator string         `@("&" | "|")`
	Next     *DirectiveExpr `@@`
}

type Expression struct {
	Directive *DirectiveExpr `@@`
	Children  []*Children    `("." @@)*`
}

func Parser(str string) (expr *Expression, err error) {
	expr = &Expression{}
	parser := participle.MustBuild(expr)
	err = parser.ParseString(str, expr)
	return
}
