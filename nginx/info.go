package nginx

import (
	"bufio"
	"bytes"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
)

func MustConf() string {
	if _, conf, err := GetInfo(); err != nil {
		return "/etc/nginx/nginx.conf"
	} else {
		return conf
	}
}

func MustConfigDir() string {
	return filepath.Dir(MustConf())
}

func GetInfo() (path, file string, err error) {
	writer := bytes.NewBufferString("")
	cmd := exec.Command("nginx", "-h")
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
			path = filepath.Dir(line[idx+9 : len(line)-1])
		} else if strings.HasPrefix(line, "-c filename") {
			idx := strings.Index(line, "default:")
			file = line[idx+9 : len(line)-1]
		}
	}
	if !strings.HasPrefix(file, "/") {
		file = path + "/" + file
	}
	return
}
