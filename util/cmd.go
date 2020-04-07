package util

import (
	"bytes"
	"errors"
	"os/exec"
)

type StdError struct {
	*bytes.Buffer
}

func (std *StdError) Error() error {
	return errors.New(std.String())
}

func CmdAfterWait(cmd *exec.Cmd) error {
	err := cmd.Wait()
	if err != nil {
		return cmd.Stderr.(*StdError).Error()
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
	cmd.Stderr = &StdError{bytes.NewBufferString("")}
	cmd.Stdout = &StdError{bytes.NewBufferString("")}
	if err := cmd.Start(); err != nil {
		return nil, err
	} else {
		return cmd, nil
	}
}
