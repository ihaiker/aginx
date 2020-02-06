package util

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os/exec"
)

type ErrStdErr struct {
	*bytes.Buffer
}

func (std *ErrStdErr) Error() error {
	lastLine := ""
	reader := bufio.NewReader(std)
	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		lastLine = string(line)
	}
	return errors.New(lastLine)
}

func CmdAfterWait(cmd *exec.Cmd) error {
	err := cmd.Wait()
	if err != nil {
		return cmd.Stderr.(*ErrStdErr).Error()
	}
	return nil
}

func CmdRun(command string, args ...string) error {
	cmd, err := CmdStart(command, args...)
	if err != nil {
		return err
	}
	return CmdAfterWait(cmd)
}

func CmdStart(command string, args ...string) (*exec.Cmd, error) {
	cmd := exec.Command(command, args...)
	cmd.Stderr = &ErrStdErr{bytes.NewBufferString("")}
	return cmd, cmd.Start()
}
