package configuration

import (
	"bytes"
	"os"
	"path/filepath"
)

type Writer func(file string, content []byte) error

func FileWriter(path string, content []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	_, err = f.Write(content)
	return err
}

func Down(root string, cfg *Configuration) error {
	return DownWriter(root, cfg, FileWriter)
}

func DownWriter(root string, cfg *Configuration, writerFn Writer) (err error) {
	content := cfg.String()
	if err = writerFn(root+"/nginx.conf", []byte(content)); err != nil {
		return
	}
	if err = writeVirtual(root, cfg.Directive(), writerFn); err != nil {
		return
	}
	return
}

func writeVirtual(testRoot string, directive *Directive, writerFn Writer) error {
	for _, body := range directive.Body {
		switch body.Virtual {
		case File:
			filePath := testRoot + "/" + body.Name
			content := body.Args[0]
			if err := writerFn(filePath, []byte(content)); err != nil {
				return err
			}
		case Include:
			filePath := testRoot + "/" + body.Name
			content := bytes.NewBufferString("")
			for _, d := range body.Body {
				content.WriteString(d.Pretty(0))
				if err := writeVirtual(testRoot, d, writerFn); err != nil {
					return err
				}
			}
			if err := writerFn(filePath, content.Bytes()); err != nil {
				return err
			}
		default:
			if err := writeVirtual(testRoot, body, writerFn); err != nil {
				return err
			}
		}
	}
	return nil
}
