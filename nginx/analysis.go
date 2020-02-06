package nginx

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/xhaiker/codf"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Directive struct {
	Virtual bool         `json:"virtual,omitempty"`
	Name    string       `json:"name"`
	Args    []string     `json:"args,omitempty"`
	Body    []*Directive `json:"body,omitempty"`
}

func NewDirective(name string, args ...string) *Directive {
	return &Directive{
		Name: name,
		Args: args,
	}
}

func (d Directive) String() string {
	return d.Pretty(0)
}

func (d Directive) Pretty(prefix int) string {
	prefixString := strings.Repeat(" ", prefix*4)
	if d.Virtual {
		return ""
	} else {
		out := bytes.NewBufferString(prefixString)
		out.WriteString(d.Name)
		out.WriteString(" ")
		if len(d.Args) > 0 {
			out.WriteString(strings.Join(d.Args, " "))
		}
		if d.Name == "include" {
			out.WriteString(";")
		} else if d.Body != nil {
			out.WriteString(" {")
			for _, body := range d.Body {
				out.WriteString("\n")
				out.WriteString(body.Pretty(prefix + 1))
			}
			out.WriteString(fmt.Sprintf("\n%s}", prefixString))
		} else {
			out.WriteString(";")
		}
		return out.String()
	}
}

type Configuration Directive

func (conf *Configuration) Directive() *Directive {
	return (*Directive)(conf)
}

func (conf Configuration) String() string {
	out := bytes.NewBufferString("# " + time.Now().Format(time.RFC3339) + "\n")
	for _, body := range conf.Body {
		out.WriteString(body.Pretty(0))
		out.WriteString("\n")
	}
	return out.String()
}

func GetInfo() (path, file string, err error) {
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
	/*
		if strings.HasPrefix(file, "/") {
			path = filepath.Dir(file)
		}
	*/
	return
}

func AnalysisNginx() (*Configuration, error) {
	path, conf, err := GetInfo()
	if err != nil {
		return nil, err
	}
	return AnalysisFromFile(path, conf)
}

type nameReader struct {
	io.Reader
	namefn func() string
}

func (n nameReader) Name() string {
	return n.namefn()
}

func NamedReader(rd io.Reader, name string) codf.NamedReader {
	return nameReader{Reader: rd, namefn: func() string {
		return name
	}}
}

func AnalysisFromFile(root, file string) (*Configuration, error) {
	path := file
	if !strings.HasPrefix(file, "/") {
		path = root + string(filepath.Separator) + file
	}
	if rd, err := os.OpenFile(path, os.O_RDONLY, os.ModeTemporary); err != nil {
		return nil, err
	} else {
		return Analysis(root, NamedReader(rd, path))
	}
}

func Analysis(root string, r codf.NamedReader) (*Configuration, error) {
	l := codf.NewLexer(r)
	p := codf.NewParser()
	if err := p.Parse(l); err != nil {
		return nil, err
	}
	doc := p.Document()
	cfg := &Configuration{
		Name: r.Name(),
		Body: make([]*Directive, 0),
	}
	for _, child := range doc.Children {
		if node, err := analysisNode(root, child); err == nil {
			cfg.Body = append(cfg.Body, node)
		} else {
			return nil, err
		}
	}
	return cfg, nil
}

func searchFiles(root, file string) []string {
	path := file
	if !strings.HasPrefix(file, "/") {
		path = root + string(filepath.Separator) + file
	}
	files, err := filepath.Glob(path)
	if err != nil {
		return []string{}
	}
	return files
}

func includes(root string, node *Directive) error {
	for _, arg := range node.Args {
		files := searchFiles(root, arg)
		for _, file := range files {
			includeDirective := &Directive{Virtual: true, Name: file}
			if doc, err := AnalysisFromFile(root, file); err != nil {
				return err
			} else {
				includeDirective.Body = doc.Body
			}
			node.Body = append(node.Body, includeDirective)
		}
	}
	return nil
}

func analysisNode(root string, child codf.Node) (direct *Directive, err error) {
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
			if direct.Body[i], err = analysisNode(root, n); err != nil {
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
