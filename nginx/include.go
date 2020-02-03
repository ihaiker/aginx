package nginx

import (
	"path/filepath"
	"strings"
)

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
				includeDirective.Body = *doc
			}
			node.Body = append(node.Body, includeDirective)
		}
	}
	return nil
}
