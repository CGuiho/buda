//go:build windows

package maintenance

import (
	"os/exec"
	"syscall"
)

func startDetached(executable string, arguments ...string) error {
	command := exec.Command(executable, arguments...)
	command.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP | 0x00000008,
		HideWindow:    true,
	}
	if err := command.Start(); err != nil {
		return err
	}
	return command.Process.Release()
}
