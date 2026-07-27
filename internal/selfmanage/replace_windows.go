//go:build windows

package selfmanage

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
	"unsafe"
)

func replaceExecutable(executable, candidate, backup, targetVersion, checksum, wiki string, _ VerifyFunc) (bool, error) {
	helper, err := copySelfHelper(executable, ".buda-upgrade-helper-*.exe")
	if err != nil {
		return false, err
	}
	arguments := []string{"upgrade", "__replace-windows", "--pid", strconv.Itoa(os.Getpid()), "--executable", executable,
		"--candidate", candidate, "--backup", backup, "--target-version", targetVersion, "--checksum", checksum, "--helper", helper}
	if wiki != "" {
		arguments = append(arguments, "--selected-wiki", wiki)
	}
	command := exec.Command(helper, arguments...)
	command.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x08000000, HideWindow: true}
	command.Stdin, command.Stdout, command.Stderr = nil, nil, nil
	if err := command.Start(); err != nil {
		_ = os.Remove(helper)
		return false, fmt.Errorf("start Windows upgrade helper: %w", err)
	}
	if err := command.Process.Release(); err != nil {
		return false, fmt.Errorf("release Windows upgrade helper: %w", err)
	}
	return true, nil
}

func CompleteWindowsReplacement(executable, candidate, backup, targetVersion, checksum, helper, wiki string, parentPID int) error {
	defer scheduleDeleteAtReboot(helper)
	if err := waitForProcess(uint32(parentPID), 2*time.Minute); err != nil {
		return err
	}
	calculated, err := fileChecksum(candidate)
	if err != nil {
		return fmt.Errorf("hash staged Windows update: %w", err)
	}
	if calculated != checksum {
		return fmt.Errorf("staged Windows candidate checksum changed: expected %s, got %s", checksum, calculated)
	}
	if _, err := os.Stat(backup); err == nil {
		return fmt.Errorf("Windows upgrade backup already exists at %s", backup)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect Windows upgrade backup: %w", err)
	}
	if err := os.Rename(executable, backup); err != nil {
		return fmt.Errorf("backup current Windows executable: %w", err)
	}
	if err := os.Rename(candidate, executable); err != nil {
		if rollbackErr := os.Rename(backup, executable); rollbackErr != nil {
			return fmt.Errorf("activate Windows update: %w; rollback also failed: %v", err, rollbackErr)
		}
		return fmt.Errorf("activate Windows update: %w", err)
	}
	if err := VerifyExecutable(executable, targetVersion); err != nil {
		failed := executable + ".failed"
		_ = os.Remove(failed)
		if moveErr := os.Rename(executable, failed); moveErr != nil {
			return fmt.Errorf("%w; stage failed replacement: %v", err, moveErr)
		}
		if rollbackErr := os.Rename(backup, executable); rollbackErr != nil {
			return fmt.Errorf("%w; rollback also failed: %v", err, rollbackErr)
		}
		_ = os.Remove(failed)
		return err
	}
	if err := reconcileResources(executable, wiki); err != nil {
		if rollbackErr := rollbackFiles(executable); rollbackErr != nil {
			return fmt.Errorf("resource reconciliation failed: %w; automatic rollback also failed: %v", err, rollbackErr)
		}
		return fmt.Errorf("resource reconciliation failed: %w; executable automatically rolled back", err)
	}
	return nil
}

func RemoveExecutable(path string) (bool, error) {
	helper, err := copySelfHelper(path, ".buda-uninstall-helper-*.exe")
	if err != nil {
		return false, err
	}
	command := exec.Command(helper, "uninstall", "__remove-windows", "--pid", strconv.Itoa(os.Getpid()), "--executable", path, "--helper", helper)
	command.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x08000000, HideWindow: true}
	command.Stdin, command.Stdout, command.Stderr = nil, nil, nil
	if err := command.Start(); err != nil {
		_ = os.Remove(helper)
		return false, fmt.Errorf("start Windows uninstall helper: %w", err)
	}
	if err := command.Process.Release(); err != nil {
		return false, fmt.Errorf("release Windows uninstall helper: %w", err)
	}
	return true, nil
}

func CompleteWindowsRemoval(executable, helper string, parentPID int) error {
	defer scheduleDeleteAtReboot(helper)
	if err := waitForProcess(uint32(parentPID), 2*time.Minute); err != nil {
		return err
	}
	if err := os.Remove(executable); err != nil {
		return fmt.Errorf("remove Buda executable: %w", err)
	}
	return nil
}

func Rollback(executable string) (bool, error) {
	helper, err := copySelfHelper(executable, ".buda-rollback-helper-*.exe")
	if err != nil {
		return false, err
	}
	command := exec.Command(helper, "upgrade", "__rollback-windows", "--pid", strconv.Itoa(os.Getpid()), "--executable", executable, "--helper", helper)
	command.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x08000000, HideWindow: true}
	command.Stdin, command.Stdout, command.Stderr = nil, nil, nil
	if err := command.Start(); err != nil {
		_ = os.Remove(helper)
		return false, fmt.Errorf("start Windows rollback helper: %w", err)
	}
	if err := command.Process.Release(); err != nil {
		return false, fmt.Errorf("release Windows rollback helper: %w", err)
	}
	return true, nil
}

func CompleteWindowsRollback(executable, helper string, parentPID int) error {
	defer scheduleDeleteAtReboot(helper)
	if err := waitForProcess(uint32(parentPID), 2*time.Minute); err != nil {
		return err
	}
	return rollbackFiles(executable)
}

func copySelfHelper(executable, pattern string) (string, error) {
	source, err := os.Open(executable)
	if err != nil {
		return "", fmt.Errorf("open Buda executable for helper: %w", err)
	}
	defer source.Close()
	file, err := os.CreateTemp(filepath.Dir(executable), pattern)
	if err != nil {
		return "", fmt.Errorf("create Buda helper: %w", err)
	}
	path := file.Name()
	if _, err := io.Copy(file, source); err != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return "", fmt.Errorf("copy Buda helper: %w", err)
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return "", fmt.Errorf("sync Buda helper: %w", err)
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(path)
		return "", fmt.Errorf("close Buda helper: %w", err)
	}
	return path, nil
}

func reconcileResources(executable, wiki string) error {
	commands := [][]string{{"agent", "skill", "update"}}
	if wiki != "" {
		commands = append(commands, []string{"agent", "instruction", "update", "--wiki", wiki})
	}
	for _, arguments := range commands {
		output, err := exec.Command(executable, arguments...).CombinedOutput()
		if err != nil {
			return fmt.Errorf("run %v: %w (%s)", arguments, err, string(output))
		}
	}
	return nil
}

func fileChecksum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func waitForProcess(pid uint32, timeout time.Duration) error {
	const synchronizeAccess = 0x00100000
	const waitObject0 = 0x00000000
	const waitTimeout = 0x00000102
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	openProcess := kernel32.NewProc("OpenProcess")
	wait := kernel32.NewProc("WaitForSingleObject")
	closeHandle := kernel32.NewProc("CloseHandle")
	handle, _, callErr := openProcess.Call(synchronizeAccess, 0, uintptr(pid))
	if handle == 0 {
		if errno, ok := callErr.(syscall.Errno); ok && errno == syscall.Errno(87) {
			return nil
		}
		return fmt.Errorf("open running Buda process: %w", callErr)
	}
	defer closeHandle.Call(handle)
	result, _, waitErr := wait.Call(handle, uintptr(timeout.Milliseconds()))
	switch result {
	case waitObject0:
		return nil
	case waitTimeout:
		return fmt.Errorf("timed out waiting for running Buda process to exit")
	default:
		return fmt.Errorf("wait for running Buda process: %w", waitErr)
	}
}

func scheduleDeleteAtReboot(path string) {
	const moveFileDelayUntilReboot = 0x00000004
	pointer, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return
	}
	_, _, _ = syscall.NewLazyDLL("kernel32.dll").NewProc("MoveFileExW").Call(uintptr(unsafe.Pointer(pointer)), 0, moveFileDelayUntilReboot)
}
