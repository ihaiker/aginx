package nginx

import (
	"bytes"
	"fmt"
	"strings"
)

type Virtual string

const (
	Include Virtual = "include"
)

type Directive struct {
	Virtual Virtual      `json:"virtual,omitempty"`
	Name    string       `json:"name"`
	Args    []string     `json:"args,omitempty"`
	Body    []*Directive `json:"body,omitempty"`
}

type Configuration = Directive

func NewDirective(name string, args ...string) *Directive {
	return &Directive{Name: name, Args: args}
}

func (d *Directive) String() string {
	return d.Pretty(0)
}

func (d *Directive) BodyBytes() []byte {
	out := bytes.NewBufferString("")
	for _, body := range d.Body {
		out.WriteString(body.Pretty(0))
		out.WriteString("\n")
	}
	return out.Bytes()
}

func (d *Directive) noBody() bool {
	if len(d.Body) == 0 {
		return true
	} else {
		for _, body := range d.Body {
			if body.Virtual == "" {
				return false
			}
		}
		return true
	}
}

func (d *Directive) AddBody(name string, args ...string) *Directive {
	body := NewDirective(name, args...)
	d.AddBodyDirective(body)
	return body
}

func (d *Directive) AddBodyDirective(directive ...*Directive) {
	if d.Body == nil {
		d.Body = make([]*Directive, 0)
	}
	d.Body = append(d.Body, directive...)
}

func (d *Directive) Pretty(prefix int) string {
	prefixString := strings.Repeat(" ", prefix*4)
	if d.Virtual != "" {
		return ""
	} else {
		out := bytes.NewBufferString(prefixString)
		out.WriteString(d.Name)
		out.WriteString(" ")
		if len(d.Args) > 0 {
			out.WriteString(strings.Join(d.Args, " "))
		}

		if d.noBody() {
			out.WriteString(";")
		} else {
			out.WriteString(" {")
			for _, body := range d.Body {
				out.WriteString("\n")
				out.WriteString(body.Pretty(prefix + 1))
			}
			out.WriteString(fmt.Sprintf("\n%s}", prefixString))
		}
		return out.String()
	}
}

func (d *Directive) find(directives []*Directive, query string) ([]*Directive, error) {
	expr, err := Parser(query)
	if err != nil {
		return nil, fmt.Errorf("Search condition error：[%s]", query)
	}
	matched := make([]*Directive, 0)
	for _, directive := range directives {
		for _, body := range directive.Body {
			if expr.Match(body) {
				matched = append(matched, body)
			}
		}
	}
	return matched, nil
}

func (d *Directive) Select(queries ...string) ([]*Directive, error) {
	current := []*Directive{d}
	for _, query := range queries {
		directives, err := d.find(current, query)
		if err != nil {
			return nil, err
		}
		if directives == nil || len(directives) == 0 {
			return nil, ErrNotFound
		}
		current = directives
	}
	return current, nil
}
