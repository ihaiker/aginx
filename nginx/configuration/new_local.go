package configuration

import (
	"bytes"
	"github.com/ihaiker/aginx/nginx"
)

func DownWriterDiffer(cfg *nginx.Configuration, writer Writer, differ Differ) (err error) {
	content := cfg.Bytes()
	if differ("nginx.conf", content) {
		if err = writer("nginx.conf", content); err != nil {
			return
		}
	}
	if err = writeVirtual(cfg.Directive(), writer, differ); err != nil {
		return
	}
	return
}

func writeVirtual(directive *nginx.Directive, writerFn Writer, differ Differ) error {
	for _, body := range directive.Body {
		switch body.Virtual {
		case nginx.Include:
			filePath := body.Args[0]
			content := bytes.NewBufferString("")
			for _, d := range body.Body {
				content.WriteString(d.Pretty(0))
				if err := writeVirtual(d, writerFn, differ); err != nil {
					return err
				}
			}
			if differ(filePath, content.Bytes()) {
				if err := writerFn(filePath, content.Bytes()); err != nil {
					return err
				}
			}
		default:
			if err := writeVirtual(body, writerFn, differ); err != nil {
				return err
			}
		}
	}
	return nil
}
