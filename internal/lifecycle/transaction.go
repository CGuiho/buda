// Package lifecycle contains small crash-safe primitives shared by all
// installation operations. It deliberately knows no Buda domain data.
package lifecycle

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

type Lock struct {
	Path  string    `json:"path"`
	Token string    `json:"token"`
	PID   int       `json:"pid"`
	Start time.Time `json:"start"`
	file  *os.File
}

func AcquireLock(path string) (*Lock, error) {
	if strings.TrimSpace(path) == "" {
		return nil, errors.New("lock path must not be empty")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	var file *os.File
	var err error
	for {
		file, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if err == nil {
			break
		}
		if !errors.Is(err, os.ErrExist) {
			return nil, err
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil, fmt.Errorf("operation lock is held and cannot be inspected: %w", readErr)
		}
		var owner Lock
		if decodeErr := json.Unmarshal(data, &owner); decodeErr != nil || owner.PID <= 0 {
			return nil, fmt.Errorf("operation lock is held with invalid ownership metadata: %s", path)
		}
		if IsProcessAlive(owner.PID) {
			return nil, fmt.Errorf("operation lock is held: %s", path)
		}
		if removeErr := os.Remove(path); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			return nil, fmt.Errorf("remove stale operation lock: %w", removeErr)
		}
	}
	tokenBytes := make([]byte, 16)
	if _, err := rand.Read(tokenBytes); err != nil {
		file.Close()
		os.Remove(path)
		return nil, err
	}
	lock := &Lock{Path: path, Token: hex.EncodeToString(tokenBytes), PID: os.Getpid(), Start: time.Now().UTC(), file: file}
	data, _ := json.Marshal(lock)
	if _, err := file.Write(append(data, '\n')); err != nil {
		file.Close()
		os.Remove(path)
		return nil, err
	}
	if err := file.Sync(); err != nil {
		file.Close()
		os.Remove(path)
		return nil, err
	}
	return lock, nil
}

func (l *Lock) Release() error {
	if l == nil {
		return nil
	}
	if l.file != nil {
		_ = l.file.Close()
	}
	data, err := os.ReadFile(l.Path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	var owner Lock
	if err := json.Unmarshal(data, &owner); err != nil {
		return err
	}
	if owner.Token != l.Token {
		return fmt.Errorf("operation lock ownership changed: %s", l.Path)
	}
	return os.Remove(l.Path)
}

type Phase string

const (
	PhasePlanned                Phase = "planned"
	PhaseStaged                 Phase = "staged"
	PhaseCandidateVerified      Phase = "candidate-verified"
	PhaseProjectionsSnapshotted Phase = "projections-snapshotted"
	PhaseArtifactsReplaced      Phase = "artifacts-replaced"
	PhaseActivated              Phase = "activated"
	PhaseVerified               Phase = "verified"
	PhaseRollingBack            Phase = "rolling-back"
	PhaseRolledBack             Phase = "rolled-back"
	PhaseComplete               Phase = "complete"
)

type Journal struct {
	Schema          int       `json:"schema"`
	Operation       string    `json:"operation"`
	Phase           Phase     `json:"phase"`
	Token           string    `json:"token"`
	Version         string    `json:"version,omitempty"`
	PreviousVersion string    `json:"previous_version,omitempty"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func SaveJournal(path string, journal Journal) error {
	journal.Schema = 1
	journal.UpdatedAt = time.Now().UTC()
	data, err := json.MarshalIndent(journal, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	temp, err := os.CreateTemp(filepath.Dir(path), ".journal-*")
	if err != nil {
		return err
	}
	tmp := temp.Name()
	defer os.Remove(tmp)
	if _, err = temp.Write(append(data, '\n')); err != nil {
		temp.Close()
		return err
	}
	if err = temp.Sync(); err != nil {
		temp.Close()
		return err
	}
	if err = temp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func LoadJournal(path string) (Journal, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Journal{}, err
	}
	var value Journal
	if err := json.Unmarshal(data, &value); err != nil {
		return Journal{}, err
	}
	if value.Schema != 1 {
		return Journal{}, errors.New("unsupported transaction journal schema")
	}
	return value, nil
}

func AtomicWrite(path string, content []byte, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".atomic-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if err := tmp.Chmod(mode); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(content); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

// IsLinked reports whether the entry is a symbolic link or, on Windows, any
// reparse point such as a junction.
func IsLinked(info os.FileInfo) bool { return isLinked(info) }

func VerifyNoLinkedAncestors(path, root string) error {
	if strings.TrimSpace(path) == "" || strings.TrimSpace(root) == "" {
		return errors.New("path and root are required")
	}
	cleanPath := filepath.Clean(path)
	cleanRoot := filepath.Clean(root)
	rel, err := filepath.Rel(cleanRoot, cleanPath)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("path %s is outside root %s", cleanPath, cleanRoot)
	}
	curr := cleanPath
	for {
		info, err := os.Lstat(curr)
		if err == nil {
			if isLinked(info) {
				return fmt.Errorf("path or ancestor is a symlink or reparse point: %s", curr)
			}
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		if curr == cleanRoot || curr == filepath.Dir(curr) {
			break
		}
		curr = filepath.Dir(curr)
	}
	return nil
}

func SafeRemove(path, root string) error {
	if err := VerifyNoLinkedAncestors(path, root); err != nil {
		return err
	}
	rel, err := filepath.Rel(root, path)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return fmt.Errorf("refuse removal outside owned root: %s", path)
	}
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if isLinked(info) {
		return errors.New("refuse removal of symlink or reparse point")
	}
	return os.RemoveAll(path)
}

func IsProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	if runtime.GOOS == "windows" {
		return processAliveWindows(pid)
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = process.Signal(syscall.Signal(0))
	return err == nil
}
