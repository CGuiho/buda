//go:build !windows

package selfmanage

import (
	"fmt"
	"os"
)

func replaceExecutable(executable, candidate, backup, targetVersion, _, _ string, verify VerifyFunc) (bool, error) {
	if err := os.Rename(executable, backup); err != nil {
		return false, fmt.Errorf("backup current executable: %w", err)
	}
	if err := os.Rename(candidate, executable); err != nil {
		if rollbackErr := os.Rename(backup, executable); rollbackErr != nil {
			return false, fmt.Errorf("activate update: %w; rollback also failed: %v", err, rollbackErr)
		}
		return false, fmt.Errorf("activate update: %w", err)
	}
	if err := verify(executable, targetVersion); err != nil {
		failed := executable + ".failed"
		_ = os.Remove(failed)
		if moveErr := os.Rename(executable, failed); moveErr != nil {
			return false, fmt.Errorf("%w; stage failed replacement: %v", err, moveErr)
		}
		if rollbackErr := os.Rename(backup, executable); rollbackErr != nil {
			return false, fmt.Errorf("%w; rollback also failed: %v", err, rollbackErr)
		}
		_ = os.Remove(failed)
		return false, err
	}
	return false, nil
}

func RemoveExecutable(path string) (bool, error) {
	if err := os.Remove(path); err != nil {
		return false, fmt.Errorf("remove Buda executable: %w", err)
	}
	return false, nil
}

func Rollback(executable string) (bool, error) {
	return false, rollbackFiles(executable)
}
