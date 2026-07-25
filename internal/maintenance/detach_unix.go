//go:build !windows

package maintenance

import (
	"os/exec"
	"syscall"
)

func startDetached(executable string, arguments ...string) error {
	command := exec.Command(executable, arguments...)
	command.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := command.Start(); err != nil {
		return err
	}
	return command.Process.Release()
}
