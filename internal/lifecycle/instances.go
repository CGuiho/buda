package lifecycle

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Instance is an owned Buda payload process record. The executable path and
// current-user identity are verified again through the platform adapter before
// a lifecycle operation can terminate the process.
type Instance struct {
	Schema     int       `json:"schema"`
	PID        int       `json:"pid"`
	Executable string    `json:"executable"`
	Version    string    `json:"version"`
	User       string    `json:"user"`
	Started    time.Time `json:"started"`
}

// RegisterInstance creates one tokenized registry record and returns an
// ownership-checked cleanup function. Registry failures are intentionally
// visible to callers so an upgrade cannot claim process safety without it.
func RegisterInstance(directory, executable, version string) (func() error, error) {
	if strings.TrimSpace(directory) == "" || strings.TrimSpace(executable) == "" {
		return nil, errors.New("instance registry directory and executable are required")
	}
	absolute, err := filepath.Abs(executable)
	if err != nil {
		return nil, err
	}
	if resolved, resolveErr := filepath.EvalSymlinks(absolute); resolveErr == nil {
		absolute = resolved
	}
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return nil, err
	}
	var token [16]byte
	if _, err := rand.Read(token[:]); err != nil {
		return nil, err
	}
	path := filepath.Join(directory, fmt.Sprintf("%d-%s.json", os.Getpid(), hex.EncodeToString(token[:])))
	value := Instance{Schema: 1, PID: os.Getpid(), Executable: filepath.Clean(absolute), Version: strings.TrimSpace(version), User: CurrentProcessUser(), Started: time.Now().UTC()}
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	if err := AtomicWrite(path, append(data, '\n'), 0o600); err != nil {
		return nil, err
	}
	return func() error {
		current, readErr := os.ReadFile(path)
		if errors.Is(readErr, os.ErrNotExist) {
			return nil
		}
		if readErr != nil {
			return readErr
		}
		var owner Instance
		if err := json.Unmarshal(current, &owner); err != nil || owner.PID != value.PID || owner.Executable != value.Executable {
			return errors.New("instance registry ownership changed")
		}
		return os.Remove(path)
	}, nil
}

// TerminateOtherInstances stops only registered, current-user instances whose
// OS-observed executable path equals oldExecutable. It never follows a child
// process tree and never targets the invoking PID.
func TerminateOtherInstances(directory string, currentPID int, oldExecutable string) error {
	if strings.TrimSpace(directory) == "" || strings.TrimSpace(oldExecutable) == "" {
		return nil
	}
	expected, err := filepath.Abs(oldExecutable)
	if err != nil {
		return err
	}
	instances, err := readInstances(directory)
	if err != nil {
		return err
	}
	currentUser := CurrentProcessUser()
	for _, instance := range instances {
		if instance.PID <= 0 || instance.PID == currentPID || instance.User != currentUser {
			continue
		}
		observed, observeErr := ProcessExecutable(instance.PID)
		if errors.Is(observeErr, os.ErrProcessDone) || !IsProcessAlive(instance.PID) {
			continue
		}
		if observeErr != nil {
			return fmt.Errorf("inspect registered Buda process %d: %w", instance.PID, observeErr)
		}
		observed, err = filepath.Abs(observed)
		if err != nil || filepath.Clean(observed) != filepath.Clean(expected) {
			continue
		}
		if err := TerminateProcess(instance.PID); err != nil {
			return fmt.Errorf("terminate registered Buda process %d: %w", instance.PID, err)
		}
		deadline := time.Now().Add(3 * time.Second)
		for IsProcessAlive(instance.PID) && time.Now().Before(deadline) {
			time.Sleep(50 * time.Millisecond)
		}
		if IsProcessAlive(instance.PID) {
			return fmt.Errorf("registered Buda process %d did not terminate", instance.PID)
		}
	}
	return nil
}

func readInstances(directory string) ([]Instance, error) {
	entries, err := os.ReadDir(directory)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	instances := make([]Instance, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		path := filepath.Join(directory, entry.Name())
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil, readErr
		}
		var instance Instance
		if decodeErr := json.Unmarshal(data, &instance); decodeErr != nil || instance.Schema != 1 || instance.PID <= 0 || strings.TrimSpace(instance.Executable) == "" {
			return nil, fmt.Errorf("invalid Buda instance record %s", path)
		}
		if !IsProcessAlive(instance.PID) {
			_ = os.Remove(path)
			continue
		}
		instances = append(instances, instance)
	}
	return instances, nil
}
