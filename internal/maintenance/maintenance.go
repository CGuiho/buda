package maintenance

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/CGuiho/buda/internal/agent"
	"github.com/spf13/cobra"
)

const WorkerCommand = "__agent-maintenance"

// ShouldSchedule identifies successful public repository operations. The root,
// all agent routes, hidden workers, help, and version never schedule work.
func ShouldSchedule(command *cobra.Command) bool {
	if command == nil || command == command.Root() || command.Hidden {
		return false
	}
	top := command
	for top.Parent() != nil && top.Parent() != command.Root() {
		top = top.Parent()
	}
	if top.Hidden || top.Name() == "agent" || top.Name() == "uninstall" {
		return false
	}
	return true
}

// Schedule starts the hidden worker and returns immediately. The caller must
// deliberately ignore errors so bootstrap cannot alter foreground behavior.
func Schedule(executable, wiki string) error {
	if executable == "" || wiki == "" {
		return fmt.Errorf("maintenance worker requires executable and wiki")
	}
	return startDetached(executable, WorkerCommand, "--wiki", wiki)
}

// Run reconciles only embedded agent resources for the already-resolved wiki.
// It performs no qmd, network, repository discovery, or cross-repository work.
func Run(service *agent.Service, wiki string) error {
	if service == nil {
		return fmt.Errorf("maintenance worker requires agent service")
	}
	lock, acquired, err := acquireLock(wiki)
	if err != nil || !acquired {
		return err
	}
	defer func() {
		_ = lock.Close()
		_ = os.Remove(lock.Name())
	}()
	if _, err := service.InstallSkill(false, ""); err != nil {
		return fmt.Errorf("reconcile global Buda skill: %w", err)
	}
	if _, err := service.ApplyInstructions(wiki); err != nil {
		return fmt.Errorf("reconcile Buda instructions: %w", err)
	}
	return nil
}

func acquireLock(wiki string) (*os.File, bool, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, false, fmt.Errorf("resolve home directory: %w", err)
	}
	directory := filepath.Join(home, ".guiho", "buda")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return nil, false, fmt.Errorf("create Buda state directory: %w", err)
	}
	digest := sha256.Sum256([]byte(filepath.Clean(wiki)))
	path := filepath.Join(directory, "agent-maintenance-"+hex.EncodeToString(digest[:8])+".lock")
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err == nil {
		_, _ = fmt.Fprintf(file, "%d\n", os.Getpid())
		return file, true, nil
	}
	if !errors.Is(err, os.ErrExist) {
		return nil, false, fmt.Errorf("acquire Buda maintenance lock: %w", err)
	}
	info, statErr := os.Stat(path)
	if statErr == nil && time.Since(info.ModTime()) > 5*time.Minute {
		if removeErr := os.Remove(path); removeErr == nil {
			file, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
			if err == nil {
				_, _ = fmt.Fprintf(file, "%d\n", os.Getpid())
				return file, true, nil
			}
		}
	}
	return nil, false, nil
}
