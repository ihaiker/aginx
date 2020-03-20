package nginx

import (
	"bytes"
	"github.com/ihaiker/aginx/plugins"
	"github.com/ihaiker/aginx/util"
	"github.com/xhaiker/codf"
	"path/filepath"
	"strings"
)

func Readable(store plugins.StorageEngine) (*Configuration, error) {
	reader, err := store.Get("nginx.conf")
	if err != nil {
		return nil, util.Wrap(err, "get nginx.conf")
	}
	return ReaderReadable(store, reader)
}

func ReaderReadable(store plugins.StorageEngine, cfgFile *plugins.ConfigurationFile) (*Configuration, error) {
	parser := codf.NewParser()
	if err := parser.Parse(codf.NewLexer(bytes.NewBuffer(cfgFile.Content))); err != nil {
		return nil, util.Wrap(err, "parse config: "+cfgFile.Name)
	}
	doc := parser.Document()
	cfg := &Configuration{
		Name: cfgFile.Name,
		Body: make([]*Directive, 0),
	}
	for _, child := range doc.Children {
		node, err := analysisNode(store, child)
		if err != nil {
			return nil, err
		}
		cfg.Body = append(cfg.Body, node)
	}
	return cfg, nil
}

func analysisNode(store plugins.StorageEngine, child codf.Node) (directive *Directive, err error) {
	directive = new(Directive)
	switch child.(type) {
	case *codf.Section:
		s := child.(*codf.Section)
		directive.Name = s.Name()
		directive.Args = make([]string, len(s.Parameters()))
		for i, param := range s.Parameters() {
			directive.Args[i] = string(param.Token().Raw)
		}
		directive.Body = make([]*Directive, len(s.Nodes()))
		for i, n := range s.Nodes() {
			if directive.Body[i], err = analysisNode(store, n); err != nil {
				return
			}
		}
	case codf.ParamNode:
		s := child.(codf.ParamNode)
		directive.Name = s.Name()
		directive.Args = make([]string, len(s.Parameters()))
		for i, param := range s.Parameters() {
			directive.Args[i] = string(param.Token().Raw)
		}
		err = virtual(store, directive)
	case codf.ExprNode:
		s := child.(codf.ExprNode)
		directive.Name = string(s.Token().Raw)
	}
	return
}

func includes(store plugins.StorageEngine, node *Directive) error {
	files, err := store.Search(node.Args...)
	if err != nil {
		return err
	}
	for _, file := range files {
		includeDirective := &Directive{Virtual: Include, Name: "file", Args: Queries(file.Name)}
		if doc, err := ReaderReadable(store, file); err != nil {
			return err
		} else {
			includeDirective.Body = doc.Body
		}
		node.Body = append(node.Body, includeDirective)
	}
	return nil
}

func virtual(store plugins.StorageEngine, directive *Directive) (err error) {
	switch directive.Name {
	case "include":
		configDir := MustConfigDir()
		for i, arg := range directive.Args {
			if strings.HasPrefix(arg, configDir) {
				directive.Args[i], _ = filepath.Rel(configDir, arg)
			}
		}
		if err = includes(store, directive); err != nil {
			return
		}
	}
	return err
}
