package nginx

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/xhaiker/codf"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Directive struct {
	Virtual bool         `json:"virtual,omitempty"`
	Name    string       `json:"name"`
	Args    []string     `json:"args,omitempty"`
	Body    []*Directive `json:"body,omitempty"`
}

func (d Directive) String() string {
	s := d.Name + " " + strings.Join(d.Args, " ")
	if d.Body != nil {
		if len(d.Body) == 0 {
			s += " {}"
		} else {
			s += fmt.Sprintf("{ %d }", len(d.Body))
		}
	}
	return s
}

type Configuration []*Directive

func (conf *Configuration) String() string {
	bs, _ := json.MarshalIndent(conf, "", "\t")
	return string(bs)
}

func nginx() (path, file string, err error) {
	writer := bytes.NewBufferString("")
	cmd := exec.Command("nginx", "-h")
	cmd.Stdout = writer
	cmd.Stderr = writer
	if err = cmd.Run(); err != nil {
		return
	}
	rd := bufio.NewReader(writer)
	for {
		lineBytes, _, err := rd.ReadLine()
		if err == io.EOF {
			break
		}
		line := strings.TrimLeft(string(lineBytes), " ")
		if strings.HasPrefix(line, "-p prefix") {
			idx := strings.Index(line, "default:")
			path = line[idx+9 : len(line)-1]
		} else if strings.HasPrefix(line, "-c filename") {
			idx := strings.Index(line, "default:")
			file = line[idx+9 : len(line)-1]
		}
	}
	return
}

func AnalysisNginx() (*Configuration, error) {
	path, conf, err := nginx()
	if err != nil {
		return nil, err
	}
	return AnalysisFromFile(path, conf)
}

func Analysis(root string, r io.Reader) (*Configuration, error) {
	l := codf.NewLexer(r)
	p := codf.NewParser()
	if err := p.Parse(l); err != nil {
		return nil, err
	}
	doc := p.Document()
	cfg := new(Configuration)
	for _, child := range doc.Children {
		if node, err := node(root, child); err == nil {
			*cfg = append(*cfg, node)
		} else {
			return nil, err
		}
	}
	return cfg, nil
}
func AnalysisFromFile(root, file string) (*Configuration, error) {
	path := file
	if !strings.HasPrefix(file, "/") {
		path = root + string(filepath.Separator) + file
	}
	if rd, err := os.OpenFile(path, os.O_RDONLY, os.ModeTemporary); err != nil {
		return nil, err
	} else {
		return Analysis(root, rd)
	}
}

func node(root string, child codf.Node) (direct *Directive, err error) {
	direct = new(Directive)
	switch child.(type) {
	case *codf.Section:
		s := child.(*codf.Section)
		direct.Name = s.Name()
		direct.Args = make([]string, len(s.Parameters()))
		for i, param := range s.Parameters() {
			direct.Args[i] = string(param.Token().Raw)
		}
		direct.Body = make([]*Directive, len(s.Nodes()))
		for i, n := range s.Nodes() {
			if direct.Body[i], err = node(root, n); err != nil {
				return
			}
		}
	case codf.ParamNode:
		s := child.(codf.ParamNode)
		direct.Name = s.Name()
		direct.Args = make([]string, len(s.Parameters()))
		for i, param := range s.Parameters() {
			direct.Args[i] = string(param.Token().Raw)
		}
		if direct.Name == "include" {
			if err = includes(root, direct); err != nil {
				return
			}
		}
	case codf.ExprNode:
		s := child.(codf.ExprNode)
		direct.Name = string(s.Token().Raw)
	}
	return
}
