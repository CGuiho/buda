// Package uninstall plans and applies only manifest-proven Buda-owned paths.
package uninstall

import (
	"errors"
	"fmt"
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

func Apply(plan Plan) error {
	for _, item := range plan.Items {
		if item.Preserved {
			continue
		}
		if strings.Contains(item.Owner, "instruction block") {
			if err := removeManagedInstruction(item.Path); err != nil {
				return err
			}
			continue
		}
		if strings.TrimSpace(item.Path) == "" || strings.TrimSpace(item.Root) == "" {
			return errors.New("uninstall plan contains an empty path or ownership root")
		}
		if err := lifecycle.SafeRemove(item.Path, item.Root); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove %s: %w", item.Path, err)
		}
	}
	return nil
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
