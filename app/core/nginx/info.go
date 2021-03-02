package nginx

import (
	"bufio"
	"bytes"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
)

func Nginx() (bin, prefix, config string, err error) {
	if bin, err = Lookup(); err != nil {
		return
	}
	prefix, config, err = HelpInfo(bin)
	return
}

func Lookup() (string, error) {
	return exec.LookPath("nginx")
}

func HelpInfo(bin string) (prefix, config string, err error) {
	writer := bytes.NewBufferString("")
	cmd := exec.Command(bin, "-h")
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
			prefix = filepath.Dir(line[idx+9 : len(line)-1])
		} else if strings.HasPrefix(line, "-c filename") {
			idx := strings.Index(line, "default:")
			config = line[idx+9 : len(line)-1]
		}
	}
	if prefix == "" || prefix == "." {
		prefix = filepath.Dir(bin)
	}
	//bugfix: window路径问题
	prefix = strings.ReplaceAll(prefix, "\\", "/")

	if !strings.HasPrefix(config, "/") || !strings.HasPrefix(config, "\\") {
		config = filepath.Join(prefix, config)
	}
	//bugfix: window路径问题
	config = strings.ReplaceAll(config, "\\", "/")
	return
}
