package query

import (
	"github.com/alecthomas/participle"
)

type Arg struct {
	Comparison string `[@("!" | "@" | "^" | "$")]`
	Value      string `@(String|RawString|Ident)`
}

type ArgAddition struct {
	Operator string `@("&" | "|")`
	Arg      *Arg   `@@`
}

type Args struct {
	Arg  *Arg           `@@`
	Next []*ArgAddition `{ @@ }`
}

type Directive struct {
	Comparison string `( ( [@("!" | "@" | "^" | "$")]`
	Name       string `@Ident )`

	All string ` | @"*" )`

	Args *Args `["(" [@@] ")"]`
}

type Children struct {
	Directive *Directive  `( @@`
	Group     *ChildArray `| "[" @@ "]" )`
}

type ChildArray struct {
	First *Directive          `@@`
	Next  []*ChildrenAddition `@@`
}

type ChildrenAddition struct {
	Operator string     `@("&" | "|")`
	Next     *Directive `@@`
}

type Expression struct {
	Directive *Directive  `@@`
	Children  []*Children `("." @@)*`
}

func Parser(str string) (expr *Expression, err error) {
	expr = &Expression{}
	parser := participle.MustBuild(expr)
	err = parser.ParseString(str, expr)
	return
}
