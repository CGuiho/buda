package qmd

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"strings"
)

type OSRunner struct{}

func (OSRunner) Run(ctx context.Context, request Request) (ProcessResult, error) {
	command := exec.CommandContext(ctx, request.Executable, request.Arguments...)
	command.Dir = request.Directory
	// qmd resolves project-local configuration from PWD before process.cwd().
	// Set it explicitly so an inherited caller PWD cannot redirect qmd outside
	// the wiki selected by Buda, including through Windows npm launchers.
	command.Env = environmentWithPWD(command.Environ(), request.Directory)
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

func environmentWithPWD(environment []string, directory string) []string {
	result := make([]string, 0, len(environment)+1)
	for _, entry := range environment {
		key, _, _ := strings.Cut(entry, "=")
		if strings.EqualFold(key, "PWD") {
			continue
		}
		result = append(result, entry)
	}
	return append(result, "PWD="+directory)
}
