// Package uninstall plans and applies only manifest-proven Buda-owned paths.
package uninstall

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/CGuiho/buda/internal/agent"
	"github.com/CGuiho/buda/internal/artifact"
	"github.com/CGuiho/buda/internal/installlayout"
	"github.com/CGuiho/buda/internal/lifecycle"
)

type Item struct {
	Path      string `json:"path"`
	Root      string `json:"root"`
	Action    string `json:"action"`
	Owner     string `json:"owner"`
	Preserved bool   `json:"preserved"`
}

type Plan struct {
	Items          []Item `json:"items"`
	Wiki           string `json:"wiki,omitempty"`
	PreserveConfig bool   `json:"preserve_config"`
	PreserveData   bool   `json:"preserve_data"`
}

func BuildPlan(layout installlayout.Layout, wiki string, preserveConfig, preserveData bool) (Plan, error) {
	if err := layout.Validate(); err != nil {
		return Plan{}, err
	}
	plan := Plan{Wiki: wiki, PreserveConfig: preserveConfig, PreserveData: preserveData}
	add := func(path, root, owner string, preserve bool) {
		action := "REMOVE"
		if preserve {
			action = "PRESERVE"
		}
		plan.Items = append(plan.Items, Item{Path: filepath.Clean(path), Root: filepath.Clean(root), Action: action, Owner: owner, Preserved: preserve})
	}

	// The shared .guiho and bin directories are never targets. Only Buda's
	// stable launcher may be removed from the shared binary directory.
	add(layout.Launcher, layout.Bin, "buda stable launcher", false)
	// Global agent-skill projections are CLI-owned leaves beneath the user's
	// home. Their shared tool directories and parent home are never targets.
	for _, tool := range []string{".agents", ".claude"} {
		add(filepath.Join(layout.Home, tool, "skills", agent.SkillID), layout.Home, "buda global agent skill", false)
	}

	manifestPath := layout.InstalledManifest
	if info, err := os.Stat(manifestPath); err == nil && !info.IsDir() {
		installedManifest, loadErr := artifact.Load(manifestPath)
		if loadErr != nil {
			return Plan{}, fmt.Errorf("refuse uninstall with invalid installed manifest: %w", loadErr)
		}
		for _, entry := range installedManifest.Artifacts {
			preserve := preserveData && entry.Persistent
			add(filepath.Join(layout.CLIHome, filepath.FromSlash(entry.InstalledPath)), layout.CLIHome, "manifest-owned "+string(entry.Ownership)+" artifact", preserve)
		}
	} else if err == nil && info.IsDir() {
		return Plan{}, errors.New("refuse uninstall with an installed-artifacts.json directory")
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return Plan{}, fmt.Errorf("inspect installed Buda manifest: %w", err)
	} else if _, cliErr := os.Stat(layout.CLIHome); cliErr == nil {
		return Plan{}, errors.New("refuse uninstall without a valid installed-artifacts.json ownership manifest")
	} else if !errors.Is(cliErr, os.ErrNotExist) {
		return Plan{}, fmt.Errorf("inspect Buda CLI home: %w", cliErr)
	}

	// Never remove the CLI-home directory recursively. It is a container that
	// may hold user-created files not represented by the Buda ownership
	// manifest. Enumerate the known lifecycle/data leaves instead and leave the
	// empty container behind when no preserved content remains.
	add(filepath.Join(layout.CLIHome, "data"), layout.CLIHome, "buda persistent data", preserveData)
	add(filepath.Join(layout.CLIHome, "database"), layout.CLIHome, "buda database", preserveData)
	for _, child := range []string{"versions", "state", "current.json", "installed-artifacts.json", "cache.json"} {
		add(filepath.Join(layout.CLIHome, child), layout.CLIHome, "buda replaceable lifecycle state", false)
	}
	add(layout.GlobalConfig, layout.CLIHome, "buda global configuration", preserveConfig)
	add(layout.CLIHome, layout.GUIHO, "buda CLI home container", true)

	if strings.TrimSpace(wiki) != "" {
		root, err := filepath.Abs(wiki)
		if err != nil {
			return Plan{}, err
		}
		root = filepath.Clean(root)
		add(filepath.Join(root, "buda.yaml"), root, "selected wiki configuration", preserveConfig)
		add(filepath.Join(root, "AGENTS.md"), root, "managed Buda AGENTS instruction block", false)
		for _, tool := range []string{".agents", ".claude"} {
			add(filepath.Join(root, tool, "skills", agent.SkillID), root, "buda selected-wiki agent skill", false)
		}
		if info, statErr := os.Stat(filepath.Join(root, "CLAUDE.md")); statErr == nil && !info.IsDir() {
			add(filepath.Join(root, "CLAUDE.md"), root, "managed Buda CLAUDE instruction block", false)
		} else if statErr != nil && !errors.Is(statErr, os.ErrNotExist) {
			return Plan{}, fmt.Errorf("inspect selected project instruction projection: %w", statErr)
		}
	}
	sort.Slice(plan.Items, func(i, j int) bool { return plan.Items[i].Path < plan.Items[j].Path })
	return plan, nil
}

type movedItem struct {
	source     string
	quarantine string
}

type fileSnapshot struct {
	path    string
	content []byte
	existed bool
}

// Apply executes the plan transactionally. Every removable path is first
// proven safe against symlink/reparse ancestors, then moved or copied into a
// unique Buda-owned quarantine directory under stagingRoot (normally the
// shared .guiho/.temp directory, so moves stay on one filesystem). Instruction
// blocks are edited only after snapshots allow byte-exact restoration.
//
// On any failure the previous state is restored before returning. If a
// restoration step itself fails, the quarantine directory is intentionally
// retained and reported so no user data can be lost.
func Apply(plan Plan, stagingRoot string) error {
	for _, item := range plan.Items {
		if strings.TrimSpace(item.Path) == "" || strings.TrimSpace(item.Root) == "" {
			return errors.New("uninstall plan contains an empty path or ownership root")
		}
		if err := lifecycle.VerifyNoLinkedAncestors(item.Path, item.Root); err != nil {
			return err
		}
	}
	if strings.TrimSpace(stagingRoot) == "" {
		return errors.New("uninstall staging root is required")
	}
	if err := os.MkdirAll(stagingRoot, 0o755); err != nil {
		return fmt.Errorf("create uninstall staging root: %w", err)
	}
	quarantineDir, err := os.MkdirTemp(stagingRoot, "buda-uninstall-")
	if err != nil {
		return fmt.Errorf("create uninstall quarantine directory: %w", err)
	}
	quarantineRetained := false
	cleanupQuarantine := func() {
		if !quarantineRetained {
			_ = os.RemoveAll(quarantineDir)
		}
	}
	defer cleanupQuarantine()

	instructionSnapshots, err := snapshotInstructions(plan)
	if err != nil {
		return err
	}

	var moved []movedItem
	restore := func() error {
		var failures []error
		for i := len(moved) - 1; i >= 0; i-- {
			item := moved[i]
			if restoreErr := restoreItem(item); restoreErr != nil {
				failures = append(failures, restoreErr)
			}
		}
		for _, snapshot := range instructionSnapshots {
			if restoreErr := restoreFileSnapshot(snapshot); restoreErr != nil {
				failures = append(failures, restoreErr)
			}
		}
		if len(failures) > 0 {
			quarantineRetained = true
			return fmt.Errorf("uninstall rollback incomplete (quarantine retained at %s): %v", quarantineDir, failures)
		}
		return nil
	}

	// Managed instruction blocks are edited first because that step still
	// reads the selected wiki configuration, which is quarantined afterwards.
	for _, item := range plan.Items {
		if item.Preserved || !strings.Contains(item.Owner, "instruction block") {
			continue
		}
		if err := removeManagedInstruction(item.Path); err != nil {
			_ = restore()
			return err
		}
	}

	for i, item := range plan.Items {
		if item.Preserved || strings.Contains(item.Owner, "instruction block") {
			continue
		}
		info, statErr := os.Lstat(item.Path)
		if errors.Is(statErr, os.ErrNotExist) {
			continue
		}
		if statErr != nil {
			_ = restore()
			return fmt.Errorf("inspect %s: %w", item.Path, statErr)
		}
		if lifecycle.IsLinked(info) {
			_ = restore()
			return fmt.Errorf("refuse removal of symlink or reparse point: %s", item.Path)
		}
		quarantineTarget := filepath.Join(quarantineDir, fmt.Sprintf("item-%d-%s", i, filepath.Base(item.Path)))
		entry := movedItem{source: item.Path, quarantine: quarantineTarget}
		if err := moveToQuarantine(item.Path, quarantineTarget); err != nil {
			_ = restore()
			return fmt.Errorf("quarantine %s: %w", item.Path, err)
		}
		moved = append(moved, entry)
	}
	return nil
}

// moveToQuarantine prefers a same-filesystem rename. When source and
// quarantine live on different volumes it falls back to copy plus verified
// removal so cross-device targets stay transactional.
func moveToQuarantine(source, target string) error {
	if err := os.Rename(source, target); err == nil {
		return nil
	}
	info, statErr := os.Stat(source)
	if statErr != nil {
		return statErr
	}
	if info.IsDir() {
		if err := copyTree(source, target); err != nil {
			return err
		}
	} else {
		if err := copyFileAtomic(source, target); err != nil {
			return err
		}
	}
	if err := os.RemoveAll(source); err != nil {
		_ = os.RemoveAll(target)
		return err
	}
	return nil
}

func restoreItem(item movedItem) error {
	if err := os.MkdirAll(filepath.Dir(item.source), 0o755); err != nil {
		return err
	}
	if err := os.Rename(item.quarantine, item.source); err == nil {
		return nil
	}
	info, statErr := os.Stat(item.quarantine)
	if statErr != nil {
		return statErr
	}
	if info.IsDir() {
		return copyTree(item.quarantine, item.source)
	}
	return copyFileAtomic(item.quarantine, item.source)
}

func snapshotInstructions(plan Plan) ([]fileSnapshot, error) {
	var snapshots []fileSnapshot
	for _, item := range plan.Items {
		if item.Preserved || !strings.Contains(item.Owner, "instruction block") {
			continue
		}
		data, err := os.ReadFile(item.Path)
		if errors.Is(err, os.ErrNotExist) {
			snapshots = append(snapshots, fileSnapshot{path: item.Path, existed: false})
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("snapshot managed instruction %s: %w", item.Path, err)
		}
		snapshots = append(snapshots, fileSnapshot{path: item.Path, content: data, existed: true})
	}
	return snapshots, nil
}

func restoreFileSnapshot(snapshot fileSnapshot) error {
	if !snapshot.existed {
		return nil
	}
	return lifecycle.AtomicWrite(snapshot.path, snapshot.content, 0o644)
}

func copyTree(source, destination string) error {
	return filepath.Walk(source, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, relative)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return copyFileAtomic(path, target)
	})
}

func copyFileAtomic(source, destination string) error {
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return err
	}
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	output, err := os.Create(destination)
	if err != nil {
		return err
	}
	if _, err := io.Copy(output, input); err != nil {
		output.Close()
		return err
	}
	return output.Close()
}

func removeManagedInstruction(path string) error {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}
	root := filepath.Dir(path)
	service := agent.NewService(agent.DefaultResources())
	if _, err := service.RemoveInstructions(root); err != nil {
		return fmt.Errorf("remove managed instruction from %s: %w", path, err)
	}
	return nil
}
