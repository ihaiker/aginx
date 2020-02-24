package nginx

import (
	"bytes"
	"github.com/ihaiker/aginx/util"
	"io/ioutil"
	"path/filepath"
)

type Writer func(file string, content []byte) error
type Differ func(file string, content []byte) bool

func Write2NGINX(cfg *Configuration) error {
	if _, conf, err := GetInfo(); err != nil {
		return err
	} else {
		return WriteTo(filepath.Dir(conf), cfg)
	}
}

func WriteTo(path string, cfg *Configuration) error {
	return Write(cfg, FileDiffer(path), FileWriter(path))
}

func Write(cfg *Configuration, differ Differ, writer Writer) (err error) {
	content := cfg.BodyBytes()
	if differ(NGINX_CONF, content) {
		if err = writer(NGINX_CONF, content); err != nil {
			return
		}
	}
	if err = writeVirtual(cfg, writer, differ); err != nil {
		return
	}
	return
}

func writeVirtual(directive *Directive, writer Writer, differ Differ) error {
	for _, body := range directive.Body {
		switch body.Virtual {
		case Include:
			filePath := body.Args[0]
			content := bytes.NewBufferString("")
			for _, d := range body.Body {
				content.WriteString(d.Pretty(0))
				content.WriteString("\n")
				if err := writeVirtual(d, writer, differ); err != nil {
					return err
				}
			}
			if differ(filePath, content.Bytes()) {
				if err := writer(filePath, content.Bytes()); err != nil {
					return err
				}
			}
		default:
			if err := writeVirtual(body, writer, differ); err != nil {
				return err
			}
		}
	}
	return nil
}

func FileWriter(root string) Writer {
	return func(file string, content []byte) error {
		fp := filepath.Join(root, file)
		return util.WriteFile(fp, content)
	}
}

func FileDiffer(root string) Differ {
	return func(file string, content []byte) bool {
		if bs, err := ioutil.ReadFile(filepath.Join(root, file)); err == nil {
			return !bytes.Equal(bs, content)
		}
		return true
	}
}
