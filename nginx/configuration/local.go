package configuration

import (
	"bytes"
	"github.com/ihaiker/aginx/util"
)

type Writer func(file string, content []byte) error

func Down(root string, cfg *Configuration) error {
	return DownWriter(root, cfg, util.WriterFile)
}

//TODO 判断异同后写文件

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
