package qmd

import (
	"fmt"
	"strings"
)

type CommandError struct {
	Capability string `json:"capability"`
	Version    string `json:"version,omitempty"`
	ExitCode   int    `json:"exit_code"`
	Stderr     string `json:"stderr"`
	Cause      error  `json:"-"`
}

func (err *CommandError) Error() string {
	detail := err.Stderr
	if detail == "" && err.Cause != nil {
		detail = err.Cause.Error()
	}
	if detail == "" {
		detail = "qmd did not provide diagnostics"
	}
	if err.ExitCode != 0 {
		return fmt.Sprintf("qmd %s failed with exit code %d: %s", err.Capability, err.ExitCode, detail)
	}
	return fmt.Sprintf("qmd %s failed: %s", err.Capability, detail)
}

func (err *CommandError) Unwrap() error { return err.Cause }

func sanitizeStderr(value []byte) string {
	lines := strings.FieldsFunc(strings.TrimSpace(string(value)), func(r rune) bool {
		return r == '\n' || r == '\r'
	})
	if len(lines) > 3 {
		lines = lines[:3]
	}
	return strings.Join(lines, "; ")
}
