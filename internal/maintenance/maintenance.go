package maintenance

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/CGuiho/buda/internal/agent"
	"github.com/spf13/cobra"
)

const WorkerCommand = "__agent-maintenance"

// ShouldSchedule identifies successful normal operations. The root is included
// so a plain invocation can reconcile the global skill without selecting a
// wiki. Agent routes, hidden workers, and uninstall never schedule work.
func ShouldSchedule(command *cobra.Command) bool {
	if command == nil || command.Hidden {
		return false
	}
	if os.Getenv("BUDA_DISABLE_MAINTENANCE") == "1" {
		return false
	}
	for current := command; current != nil; current = current.Parent() {
		if current.Hidden || current.Name() == "uninstall" || current.Name() == "upgrade" {
			return false
		}
	}
	top := command
	for top.Parent() != nil && top.Parent() != command.Root() {
		top = top.Parent()
	}
	if top.Hidden || top.Name() == "agent" {
		return false
	}
	return true
}

// Schedule starts the hidden worker and returns immediately. The caller must
// deliberately ignore errors so bootstrap cannot alter foreground behavior.
func Schedule(executable, wiki string) error {
	if executable == "" {
		return fmt.Errorf("maintenance worker requires executable")
	}
	arguments := []string{WorkerCommand}
	if strings.TrimSpace(wiki) != "" {
		arguments = append(arguments, "--wiki", wiki)
	}
	return startDetached(executable, arguments...)
}

type runOptions struct {
	stateDirectory string
}

// RunOption customizes maintenance execution without changing its operational
// boundary. It is primarily useful for isolated tests.
type RunOption func(*runOptions)

// WithStateDirectory stores the worker lock in directory instead of the
// default ~/.guiho/buda location.
func WithStateDirectory(directory string) RunOption {
	return func(options *runOptions) { options.stateDirectory = directory }
}

// Run reconciles the embedded global skill and, when wiki is explicit, the
// instruction block for exactly that already-resolved wiki. It performs no
// qmd, network, repository discovery, or cross-repository work.
func Run(service *agent.Service, wiki string, options ...RunOption) error {
	if service == nil {
		return fmt.Errorf("maintenance worker requires agent service")
	}
	settings := runOptions{}
	for _, option := range options {
		option(&settings)
	}
	if strings.TrimSpace(wiki) != "" && (!filepath.IsAbs(wiki) || filepath.Clean(wiki) != wiki) {
		return fmt.Errorf("maintenance worker requires an absolute canonical --wiki path")
	}
	globalLock, acquired, err := acquireLock("global-agent-resources", settings.stateDirectory)
	if err != nil {
		return err
	}
	if acquired {
		if _, err := service.InstallSkill(false, ""); err != nil {
			releaseLock(globalLock)
			return fmt.Errorf("reconcile global Buda skill: %w", err)
		}
		releaseLock(globalLock)
	}
	if strings.TrimSpace(wiki) != "" {
		wikiLock, wikiAcquired, err := acquireLock("wiki-instructions:"+wiki, settings.stateDirectory)
		if err != nil {
			return err
		}
		if wikiAcquired {
			defer releaseLock(wikiLock)
			if _, err := service.ApplyInstructions(wiki); err != nil {
				return fmt.Errorf("reconcile Buda instructions: %w", err)
			}
		}
	}
	return nil
}

func releaseLock(lock *os.File) {
	if lock == nil {
		return
	}
	_ = lock.Close()
	_ = os.Remove(lock.Name())
}

func acquireLock(scope, directory string) (*os.File, bool, error) {
	if strings.TrimSpace(directory) == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, false, fmt.Errorf("resolve home directory: %w", err)
		}
		directory = filepath.Join(home, ".guiho", "buda")
	}
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return nil, false, fmt.Errorf("create Buda state directory: %w", err)
	}
	digest := sha256.Sum256([]byte(scope))
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
