package configuration

import (
	"bytes"
	"os"
	"path/filepath"
)

func Down(root string, cfg *Configuration) (err error) {
	content := cfg.String()
	if err = writeFile(root+"/nginx.conf", content); err != nil {
		return
	}
	if err = writeVirtual(root, cfg.Directive()); err != nil {
		return
	}
	return
}

func writeVirtual(testRoot string, directive *Directive) error {
	for _, body := range directive.Body {
		switch body.Virtual {
		case File:
			filePath := testRoot + "/" + body.Name
			content := body.Args[0]
			if err := writeFile(filePath, content); err != nil {
				return err
			}
		case Include:
			filePath := testRoot + "/" + body.Name
			content := bytes.NewBufferString("")
			for _, d := range body.Body {
				content.WriteString(d.Pretty(0))
				if err := writeVirtual(testRoot, d); err != nil {
					return err
				}
			}
			if err := writeFile(filePath, content.String()); err != nil {
				return err
			}
		default:
			if err := writeVirtual(testRoot, body); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeFile(path, content string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	_, err = f.WriteString(content)
	return err
}
