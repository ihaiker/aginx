package configuration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"
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

func (d Directive) String() string {
	return d.Pretty(0)
}

func (d Directive) Json() string {
	bs, _ := json.MarshalIndent(d, "", "\t")
	return string(bs)
}

func (d Directive) HideBody() bool {
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

func (d Directive) Pretty(prefix int) string {
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

type Configuration Directive

func (conf *Configuration) Directive() *Directive {
	return (*Directive)(conf)
}

func (conf Configuration) String() string {
	out := bytes.NewBufferString("# -*-*- " + time.Now().Format(time.RFC3339) + " -*-*- \n")
	for _, body := range conf.Body {
		out.WriteString(body.Pretty(0))
		out.WriteString("\n")
	}
	return out.String()
}
