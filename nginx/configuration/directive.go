package configuration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

type Virtual string

const (
	File    Virtual = "file"
	Include Virtual = "include"
)

type Directive struct {
	Modify  bool         `json:"-"`
	Virtual Virtual      `json:"virtual,omitempty"`
	Name    string       `json:"name"`
	Args    []string     `json:"args,omitempty"`
	Body    []*Directive `json:"body,omitempty"`
}

func NewDirective(name string, args ...string) *Directive {
	return &Directive{Name: name, Args: args}
}

func (d *Directive) String() string {
	return d.Pretty(0)
}

func (d *Directive) Json() string {
	bs, _ := json.MarshalIndent(d, "", "\t")
	return string(bs)
}

func (d *Directive) HideBody() bool {
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

func (d *Directive) AddBodyDirective(directive *Directive) {
	if d.Body == nil {
		d.Body = make([]*Directive, 0)
	}
	d.Body = append(d.Body, directive)
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

		if d.HideBody() {
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

func (d *Directive) Query() string {
	if d.Virtual != "" {
		return ""
	} else {
		out := bytes.NewBufferString(d.Name)
		if len(d.Args) > 0 {
			out.WriteString("(")
			for i, arg := range d.Args {
				out.WriteString("'")
				out.WriteString(arg)
				out.WriteString("'")
				if i < len(d.Args)-1 {
					out.WriteString(" & ")
				}
			}
			out.WriteString(")")
		}

		if !d.HideBody() {
			bodyLen := len(d.Body)
			if bodyLen > 1 {
				out.WriteString(".[")
			}

			for i, body := range d.Body {
				out.WriteString(body.Query())
				if i < bodyLen-1 {
					out.WriteString(" & ")
				}
			}

			if bodyLen > 1 {
				out.WriteString("]")
			}
		}
		return out.String()
	}
}

type Configuration Directive

func (conf *Configuration) Directive() *Directive {
	return (*Directive)(conf)
}

func (conf Configuration) String() string {
	out := bytes.NewBufferString("")
	for _, body := range conf.Body {
		out.WriteString(body.Pretty(0))
		out.WriteString("\n")
	}
	return out.String()
}
