//go:build !windows

package lifecycle

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"syscall"
)

func processAliveWindows(pid int) bool { return false }

func CurrentProcessUser() string { return strconv.Itoa(os.Getuid()) }

func ProcessExecutable(pid int) (string, error) {
	path, err := os.Readlink(fmt.Sprintf("/proc/%d/exe", pid))
	if errors.Is(err, os.ErrNotExist) {
		return "", os.ErrProcessDone
	}
	return path, err
}

func TerminateProcess(pid int) error {
	if err := syscall.Kill(pid, syscall.SIGTERM); err != nil && !errors.Is(err, os.ErrProcessDone) {
		return err
	}
	return nil
}
