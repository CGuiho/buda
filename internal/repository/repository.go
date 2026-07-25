// Package repository resolves one explicit Buda wiki and enforces path
// containment. It never searches ancestors, siblings, remembered paths, or a
// global registry.
package repository

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/CGuiho/buda/internal/config"
)

type Repository struct {
	Root       string        `json:"root"`
	ConfigPath string        `json:"config_path"`
	Bundle     string        `json:"bundle"`
	QMDProject string        `json:"qmd_project_directory"`
	Derived    string        `json:"derived"`
	Collection string        `json:"collection"`
	Config     config.Config `json:"config"`
}

// ResolveTarget resolves an explicit initialization target. It permits an
// absent final path but rejects missing/blank input and file targets.
func ResolveTarget(wiki string) (string, error) {
	if strings.TrimSpace(wiki) == "" {
		return "", errors.New("--wiki is required; Buda never selects a wiki implicitly")
	}
	absolute, err := filepath.Abs(wiki)
	if err != nil {
		return "", fmt.Errorf("resolve wiki %q: %w", wiki, err)
	}
	if info, err := os.Stat(absolute); err == nil {
		if !info.IsDir() {
			return "", fmt.Errorf("wiki path %q is not a directory", absolute)
		}
		resolved, err := filepath.EvalSymlinks(absolute)
		if err != nil {
			return "", fmt.Errorf("resolve wiki symlinks %q: %w", absolute, err)
		}
		return filepath.Clean(resolved), nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("inspect wiki %q: %w", absolute, err)
	}
	parent, err := evalExistingParent(filepath.Dir(absolute))
	if err != nil {
		return "", err
	}
	return filepath.Join(parent, filepath.Base(absolute)), nil
}

// Open resolves exactly the selected root and its root buda.yaml. It does not
// search upward if the selected path is a child of a wiki.
func Open(wiki string) (Repository, error) {
	root, err := ResolveTarget(wiki)
	if err != nil {
		return Repository{}, err
	}
	info, err := os.Stat(root)
	if err != nil {
		return Repository{}, fmt.Errorf("open wiki %q: %w", root, err)
	}
	if !info.IsDir() {
		return Repository{}, fmt.Errorf("wiki path %q is not a directory", root)
	}
	configPath := filepath.Join(root, config.FileName)
	configuration, err := config.Load(configPath)
	if err != nil {
		return Repository{}, fmt.Errorf("selected wiki must contain exactly one root %s: %w", config.FileName, err)
	}
	bundle, err := ResolveContained(root, configuration.Bundle)
	if err != nil {
		return Repository{}, fmt.Errorf("resolve bundle: %w", err)
	}
	qmdProject, err := ResolveContained(root, configuration.QMD.ProjectDirectory)
	if err != nil {
		return Repository{}, fmt.Errorf("resolve qmd project directory: %w", err)
	}
	derived, err := ResolveContained(root, configuration.Derived)
	if err != nil {
		return Repository{}, fmt.Errorf("resolve derived directory: %w", err)
	}
	return Repository{
		Root:       root,
		ConfigPath: configPath,
		Bundle:     bundle,
		QMDProject: qmdProject,
		Derived:    derived,
		Collection: configuration.QMD.Collection,
		Config:     configuration,
	}, nil
}

// ResolveContained resolves a repository-relative path and rejects lexical or
// existing-symlink escapes.
func ResolveContained(root, relative string) (string, error) {
	if filepath.IsAbs(relative) || strings.TrimSpace(relative) == "" {
		return "", fmt.Errorf("path %q must be repository-relative", relative)
	}
	rootAbsolute, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("resolve repository root: %w", err)
	}
	target := filepath.Join(rootAbsolute, filepath.Clean(relative))
	if !isWithin(rootAbsolute, target) {
		return "", fmt.Errorf("path %q escapes selected wiki %q", relative, rootAbsolute)
	}
	resolvedRoot, err := filepath.EvalSymlinks(rootAbsolute)
	if err != nil {
		return "", fmt.Errorf("resolve repository root symlinks: %w", err)
	}
	resolvedTarget, err := resolveWithExistingParent(target)
	if err != nil {
		return "", err
	}
	if !isWithin(resolvedRoot, resolvedTarget) {
		return "", fmt.Errorf("path %q resolves outside selected wiki %q", relative, resolvedRoot)
	}
	return filepath.Clean(resolvedTarget), nil
}

func isWithin(root, target string) bool {
	relative, err := filepath.Rel(root, target)
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) && !filepath.IsAbs(relative)
}

func resolveWithExistingParent(target string) (string, error) {
	if _, err := os.Lstat(target); err == nil {
		resolved, err := filepath.EvalSymlinks(target)
		if err != nil {
			return "", fmt.Errorf("resolve path symlinks %q: %w", target, err)
		}
		return resolved, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("inspect path %q: %w", target, err)
	}
	parent, err := evalExistingParent(filepath.Dir(target))
	if err != nil {
		return "", err
	}
	return filepath.Join(parent, filepath.Base(target)), nil
}

func evalExistingParent(path string) (string, error) {
	missing := make([]string, 0)
	cursor := filepath.Clean(path)
	for {
		if _, err := os.Lstat(cursor); err == nil {
			resolved, err := filepath.EvalSymlinks(cursor)
			if err != nil {
				return "", fmt.Errorf("resolve path symlinks %q: %w", cursor, err)
			}
			for index := len(missing) - 1; index >= 0; index-- {
				resolved = filepath.Join(resolved, missing[index])
			}
			return resolved, nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("inspect path %q: %w", cursor, err)
		}
		parent := filepath.Dir(cursor)
		if parent == cursor {
			return "", fmt.Errorf("no existing parent for %q", path)
		}
		missing = append(missing, filepath.Base(cursor))
		cursor = parent
	}
}
