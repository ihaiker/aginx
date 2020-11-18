package nginx

import (
	"bytes"
	"github.com/ihaiker/aginx/v2/core/nginx/config"
	"github.com/ihaiker/aginx/v2/core/util/files"
	"github.com/ihaiker/aginx/v2/plugins/storage"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Writer func(file string, content []byte) error
type Differ func(file string, content []byte) bool
type Remover func(file string) error

//把配置文件写到本地路径path下面
func Write2Path(path string, cfg *config.Configuration) error {
	return Write(cfg, fileDiffer(path), fileWriter(path), fileRemove(path))
}

//把配置文件写入指定的写入器内
func Write(cfg *config.Configuration, differ Differ, writer Writer, remover Remover) (err error) {
	content := cfg.BodyBytes()
	if differ("nginx.conf", content) {
		if err = writer("nginx.conf", content); err != nil {
			return
		}
	}
	if err = writeVirtual(cfg, writer, differ, remover); err != nil {
		return
	}
	return
}

func Write2Storage(cfg *config.Configuration, engine storage.Plugin) error {
	return Write(cfg, storageDiffer(engine), engine.Put, engine.Remove)
}

func writeVirtual(directive *config.Directive, writer Writer, differ Differ, remover Remover) error {
	for _, body := range directive.Body {
		switch body.Virtual {
		case config.Include:
			filePath := body.Args[0]
			if body.Body == nil || len(body.Body) == 0 { //include 的文件为空文件，删除
				if err := remover(filePath); err != nil {
					return err
				}
			} else {
				content := bytes.NewBufferString("")
				for _, d := range body.Body {
					content.WriteString(d.Pretty(0))
					content.WriteString("\n")
					if err := writeVirtual(d, writer, differ, remover); err != nil {
						return err
					}
				}
				if differ(filePath, content.Bytes()) {
					if err := writer(filePath, content.Bytes()); err != nil {
						return err
					}
				}
			}
		default:
			if err := writeVirtual(body, writer, differ, remover); err != nil {
				return err
			}
		}
	}
	return nil
}

func fileWriter(root string) Writer {
	return func(file string, content []byte) error {
		fp := filepath.Join(root, file)
		return files.WriteFile(fp, content)
	}
}

func fileDiffer(root string) Differ {
	return func(file string, content []byte) bool {
		if bs, err := ioutil.ReadFile(filepath.Join(root, file)); err == nil {
			return !bytes.Equal(bs, content)
		}
		return true
	}
}

func fileRemove(root string) Remover {
	return func(file string) error {
		fp := filepath.Join(root, file)
		return os.Remove(fp)
	}
}

func storageDiffer(engine storage.Plugin) Differ {
	return func(file string, content []byte) bool {
		if fileContent, err := engine.Get(file); err == nil {
			return !bytes.Equal(fileContent.Content, content)
		}
		return true
	}
}
