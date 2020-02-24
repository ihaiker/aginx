package client

import (
	"github.com/ihaiker/aginx/nginx"
	"github.com/ihaiker/aginx/storage"
	"github.com/ihaiker/aginx/util"
	"github.com/xhaiker/codf"
)

func Readable(store storage.Engine) (*nginx.Configuration, error) {
	reader, err := store.Get("nginx.conf")
	if err != nil {
		return nil, err
	}
	return ReaderReadable(store, reader)
}

func ReaderReadable(store storage.Engine, reader *util.NameReader) (*nginx.Configuration, error) {
	parser := codf.NewParser()
	if err := parser.Parse(codf.NewLexer(reader)); err != nil {
		return nil, err
	}
	doc := parser.Document()
	cfg := &nginx.Configuration{
		Name: reader.Name,
		Body: make([]*nginx.Directive, 0),
	}
	for _, child := range doc.Children {
		if node, err := analysisNode(store, child); err == nil {
			cfg.Body = append(cfg.Body, node)
		} else {
			return nil, err
		}
	}
	return cfg, nil
}

func analysisNode(store storage.Engine, child codf.Node) (directive *nginx.Directive, err error) {
	directive = new(nginx.Directive)
	switch child.(type) {
	case *codf.Section:
		s := child.(*codf.Section)
		directive.Name = s.Name()
		directive.Args = make([]string, len(s.Parameters()))
		for i, param := range s.Parameters() {
			directive.Args[i] = string(param.Token().Raw)
		}
		directive.Body = make([]*nginx.Directive, len(s.Nodes()))
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

func includes(store storage.Engine, node *nginx.Directive) error {
	files, err := store.Search(node.Args...)
	if err != nil {
		return err
	}
	for _, file := range files {
		includeDirective := &nginx.Directive{Virtual: nginx.Include, Name: "file", Args: Queries(file.Name)}
		if doc, err := ReaderReadable(store, file); err != nil {
			return err
		} else {
			includeDirective.Body = doc.Body
		}
		node.Body = append(node.Body, includeDirective)
	}
	return nil
}

func virtual(store storage.Engine, directive *nginx.Directive) (err error) {
	switch directive.Name {
	case "include":
		if err = includes(store, directive); err != nil {
			return
		}
	}
	return err
}
