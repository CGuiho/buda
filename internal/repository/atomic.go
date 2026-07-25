package repository

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// AtomicWriteFile writes and syncs a same-directory candidate before replacing
// the destination. If an existing destination must be moved aside, a failed
// replacement restores it.
func AtomicWriteFile(path string, data []byte, mode os.FileMode) error {
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return fmt.Errorf("create parent directory %q: %w", directory, err)
	}
	temporary, err := os.CreateTemp(directory, ".buda-write-*")
	if err != nil {
		return fmt.Errorf("create atomic candidate for %q: %w", path, err)
	}
	temporaryPath := temporary.Name()
	cleanup := func() { _ = os.Remove(temporaryPath) }
	defer cleanup()
	if _, err := temporary.Write(data); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("write atomic candidate for %q: %w", path, err)
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("sync atomic candidate for %q: %w", path, err)
	}
	if err := temporary.Chmod(mode); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("set atomic candidate mode for %q: %w", path, err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close atomic candidate for %q: %w", path, err)
	}

	backup := ""
	if _, err := os.Lstat(path); err == nil {
		backupFile, createErr := os.CreateTemp(directory, ".buda-backup-*")
		if createErr != nil {
			return fmt.Errorf("reserve atomic backup for %q: %w", path, createErr)
		}
		backup = backupFile.Name()
		if err := backupFile.Close(); err != nil {
			return err
		}
		if err := os.Remove(backup); err != nil {
			return err
		}
		if err := os.Rename(path, backup); err != nil {
			return fmt.Errorf("stage existing file %q: %w", path, err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect destination %q: %w", path, err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		if backup != "" {
			_ = os.Rename(backup, path)
		}
		return fmt.Errorf("replace %q atomically: %w", path, err)
	}
	temporaryPath = ""
	if backup != "" {
		if err := os.Remove(backup); err != nil {
			return fmt.Errorf("remove atomic backup for %q: %w", path, err)
		}
	}
	return nil
}
