package qmd

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
)

type OSRunner struct{}

func (OSRunner) Run(ctx context.Context, request Request) (ProcessResult, error) {
	command := exec.CommandContext(ctx, request.Executable, request.Arguments...)
	command.Dir = request.Directory
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	result := ProcessResult{Stdout: stdout.Bytes(), Stderr: stderr.Bytes()}
	if err == nil {
		return result, nil
	}
	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		result.ExitCode = exitError.ExitCode()
		return result, nil
	}
	return result, err
}
