package std

import (
	"io"
	"os/exec"
)

type ProcessMetadata struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser
}

func ExecuteCommand(commandName string, commandArgs []string, env []string, cwd string) (metadata *ProcessMetadata, err error) {
	cmd := exec.Command(commandName, commandArgs...)
	cmd.Env = env
	cmd.Dir = cwd

	var stdout, stderr io.ReadCloser
	var stdin io.WriteCloser

	if stdout, err = cmd.StdoutPipe(); err != nil {
		return
	}

	if stderr, err = cmd.StderrPipe(); err != nil {
		return nil, err
	}

	if stdin, err = cmd.StdinPipe(); err != nil {
		return
	}

	if err = cmd.Start(); err != nil {
		return
	}

	metadata = &ProcessMetadata{cmd, stdin, stdout, stderr}

	return
}
