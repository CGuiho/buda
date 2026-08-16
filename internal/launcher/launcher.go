// Package launcher implements the stable, non-versioned Buda entrypoint.
package launcher

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/CGuiho/buda/internal/installlayout"
)

type Pointer struct {
	Schema          int    `json:"schema"`
	Active          string `json:"active"`
	Previous        string `json:"previous,omitempty"`
	ActiveVersion   string `json:"active_version"`
	PreviousVersion string `json:"previous_version,omitempty"`
}

var pointerSemver = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?(?:\+[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?$`)

func ReadPointer(path string) (Pointer, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Pointer{}, err
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	var pointer Pointer
	if err := decoder.Decode(&pointer); err != nil {
		return Pointer{}, fmt.Errorf("decode active Buda pointer: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return Pointer{}, errors.New("pointer must contain exactly one JSON document")
		}
		return Pointer{}, fmt.Errorf("decode active Buda pointer: expected exactly one JSON document: %w", err)
	}
	if pointer.Schema != 1 {
		return Pointer{}, errors.New("unsupported active Buda pointer schema")
	}
	if err := validateRelative(pointer.Active); err != nil {
		return Pointer{}, fmt.Errorf("active payload: %w", err)
	}
	baseActive := filepath.Base(filepath.FromSlash(pointer.Active))
	if baseActive != "buda" && baseActive != "buda.exe" {
		return Pointer{}, fmt.Errorf("active payload filename %q is not a canonical Buda payload name", baseActive)
	}
	if !validVersion(pointer.ActiveVersion) || pointerVersion(pointer.Active) != pointer.ActiveVersion {
		return Pointer{}, errors.New("active payload version does not match active_version")
	}
	if pointer.Previous == "" && strings.TrimSpace(pointer.PreviousVersion) != "" {
		return Pointer{}, errors.New("previous_version requires a previous payload")
	}
	if pointer.Previous != "" {
		if err := validateRelative(pointer.Previous); err != nil {
			return Pointer{}, fmt.Errorf("previous payload: %w", err)
		}
		basePrev := filepath.Base(filepath.FromSlash(pointer.Previous))
		if basePrev != "buda" && basePrev != "buda.exe" {
			return Pointer{}, fmt.Errorf("previous payload filename %q is not a canonical Buda payload name", basePrev)
		}
		if !validVersion(pointer.PreviousVersion) || pointerVersion(pointer.Previous) != pointer.PreviousVersion {
			return Pointer{}, errors.New("previous payload version does not match previous_version")
		}
	}
	return pointer, nil
}

func pointerVersion(path string) string {
	parts := strings.Split(filepath.ToSlash(filepath.Clean(filepath.FromSlash(path))), "/")
	if len(parts) < 2 {
		return ""
	}
	return parts[0]
}

func validateRelative(value string) error {
	normalized := strings.ReplaceAll(value, "\\", "/")
	if strings.TrimSpace(value) == "" || filepath.IsAbs(value) || filepath.VolumeName(value) != "" || strings.ContainsRune(value, ':') || strings.ContainsRune(value, '\x00') || strings.HasPrefix(normalized, "/") {
		return errors.New("payload path must be relative")
	}
	for _, component := range strings.Split(normalized, "/") {
		if component == ".." {
			return errors.New("payload path traverses the versions root")
		}
	}
	clean := filepath.Clean(filepath.FromSlash(normalized))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return errors.New("payload path traverses the versions root")
	}
	return nil
}

func validVersion(value string) bool {
	match := pointerSemver.FindStringSubmatch(strings.TrimPrefix(strings.TrimSpace(value), "v"))
	if match == nil {
		return false
	}
	for _, identifier := range strings.Split(match[4], ".") {
		if len(identifier) > 1 && identifier[0] == '0' {
			return false
		}
	}
	return true
}

func Run(args []string, in io.Reader, out, errOut io.Writer, layout installlayout.Layout) int {
	if err := layout.Validate(); err != nil {
		fmt.Fprintln(errOut, err)
		return 1
	}
	pointer, err := ReadPointer(layout.Current)
	if err != nil {
		fmt.Fprintf(errOut, "buda launcher: %v\n", err)
		return 1
	}
	active, activeErr := resolvePayload(layout.Versions, pointer.Active)
	var status int
	started := false
	var runErr error
	if activeErr == nil {
		status, started, runErr = invoke(active, args, in, out, errOut)
	} else {
		runErr = activeErr
	}
	if started {
		return status
	}
	if pointer.Previous != "" {
		previous, resolveErr := resolvePayload(layout.Versions, pointer.Previous)
		if resolveErr != nil {
			fmt.Fprintf(errOut, "buda launcher: active payload failed (%v); fallback payload: %v\n", runErr, resolveErr)
			return 1
		}
		fallbackStatus, _, fallbackErr := invoke(previous, args, in, out, errOut)
		if fallbackErr == nil {
			return fallbackStatus
		}
		fmt.Fprintf(errOut, "buda launcher: active payload failed (%v); fallback failed (%v)\n", runErr, fallbackErr)
		return 1
	}
	fmt.Fprintf(errOut, "buda launcher: active payload failed: %v\n", runErr)
	return 1
}

func invoke(path string, args []string, in io.Reader, out, errOut io.Writer) (status int, started bool, err error) {
	info, statErr := os.Lstat(path)
	if statErr != nil {
		return 1, false, statErr
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return 1, false, errors.New("payload is a symlink")
	}
	if !info.Mode().IsRegular() {
		return 1, false, errors.New("payload is not a regular file")
	}
	command := exec.Command(path, args...)
	command.Stdin = in
	command.Stdout = out
	command.Stderr = errOut
	err = command.Run()
	if err == nil {
		return 0, true, nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode(), true, err
	}
	return 1, false, err
}

func resolvePayload(root, relative string) (string, error) {
	if err := validateRelative(relative); err != nil {
		return "", err
	}
	root, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	target := filepath.Join(root, filepath.FromSlash(relative))
	if resolvedRoot, rootErr := filepath.EvalSymlinks(root); rootErr == nil {
		root = filepath.Clean(resolvedRoot)
	}
	resolved, err := filepath.EvalSymlinks(target)
	if err != nil {
		return "", err
	}
	resolved = filepath.Clean(resolved)
	rel, err := filepath.Rel(root, resolved)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return "", errors.New("payload resolves outside the versions directory")
	}
	return target, nil
}

func Exit(status int) {
	if runtime.GOOS == "windows" && status < 0 {
		status = 1
	}
	os.Exit(status)
}
