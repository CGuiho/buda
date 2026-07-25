package repository

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/CGuiho/buda/internal/config"
)

type InitOptions struct {
	WikiID       string
	Now          time.Time
	BeforeCommit func(Repository) error
}

type InitResult struct {
	Repository Repository `json:"repository"`
	Created    []string   `json:"created"`
	Unchanged  bool       `json:"unchanged"`
}

// Initialize creates the canonical OKF/Buda filesystem only. The command layer
// must validate qmd before calling it, then install embedded skills and bounded
// repository instructions after this succeeds.
func Initialize(wiki string, options InitOptions) (InitResult, error) {
	if strings.TrimSpace(options.WikiID) == "" {
		return InitResult{}, errors.New("--wiki-id is required")
	}
	root, err := ResolveTarget(wiki)
	if err != nil {
		return InitResult{}, err
	}
	if options.Now.IsZero() {
		options.Now = time.Now().UTC()
	}
	initialized := false
	rootCreated := false
	if info, err := os.Stat(root); err == nil {
		if !info.IsDir() {
			return InitResult{}, fmt.Errorf("wiki target %q is not a directory", root)
		}
		if _, err := os.Stat(filepath.Join(root, config.FileName)); err == nil {
			existing, err := Open(root)
			if err != nil {
				return InitResult{}, fmt.Errorf("target contains an invalid Buda configuration: %w", err)
			}
			if existing.Config.WikiID != options.WikiID {
				return InitResult{}, fmt.Errorf("wiki_id is immutable: target uses %q, requested %q", existing.Config.WikiID, options.WikiID)
			}
			initialized = true
		} else if !errors.Is(err, os.ErrNotExist) {
			return InitResult{}, err
		}
		if !initialized {
			entries, err := os.ReadDir(root)
			if err != nil {
				return InitResult{}, fmt.Errorf("inspect initialization target: %w", err)
			}
			for _, entry := range entries {
				if entry.Name() != ".git" {
					return InitResult{}, fmt.Errorf("refusing incompatible non-empty target %q (found %s)", root, entry.Name())
				}
			}
		}
	} else if errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(root, 0o755); err != nil {
			return InitResult{}, fmt.Errorf("create wiki root: %w", err)
		}
		rootCreated = true
	} else {
		return InitResult{}, fmt.Errorf("inspect initialization target: %w", err)
	}

	configuration := config.Default(options.WikiID)
	configurationData, err := config.Marshal(configuration)
	if err != nil {
		return InitResult{}, err
	}
	files := map[string][]byte{
		config.FileName: configurationData,
		filepath.Join(configuration.Bundle, "index.md"): []byte("---\nokf_version: \"0.2\"\n---\n# Knowledge\n\nThis Buda wiki uses Open Knowledge Format v0.2.\n"),
		filepath.Join(configuration.Bundle, "log.md"):   []byte(fmt.Sprintf("# Wiki Update Log\n\n## %s\n* **Initialization**: Created the Buda wiki.\n", options.Now.UTC().Format("2006-01-02"))),
	}
	directories := []string{
		filepath.Join(configuration.Bundle, "concepts"),
		filepath.Join(configuration.Bundle, "sources"),
		filepath.Join(configuration.Bundle, "references", "raw"),
		configuration.QMD.ProjectDirectory,
		configuration.Derived,
	}
	created := make([]string, 0, len(files)+len(directories)+1)
	createdTopLevel := make(map[string]bool)
	for _, relative := range directories {
		path, err := ResolveContained(root, relative)
		if err != nil {
			return InitResult{}, err
		}
		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			topLevel := strings.Split(filepath.ToSlash(relative), "/")[0]
			topLevelPath := filepath.Join(root, filepath.FromSlash(topLevel))
			if _, topErr := os.Stat(topLevelPath); errors.Is(topErr, os.ErrNotExist) {
				createdTopLevel[topLevelPath] = true
			}
			if err := os.MkdirAll(path, 0o755); err != nil {
				return InitResult{}, fmt.Errorf("create %q: %w", relative, err)
			}
			created = append(created, filepath.ToSlash(relative)+"/")
		} else if err != nil {
			return InitResult{}, err
		}
	}
	synthetic := Repository{
		Root: root, ConfigPath: filepath.Join(root, config.FileName),
		Bundle:     filepath.Join(root, configuration.Bundle),
		QMDProject: filepath.Join(root, configuration.QMD.ProjectDirectory),
		Derived:    filepath.Join(root, configuration.Derived),
		Collection: configuration.QMD.Collection, Config: configuration,
	}
	if options.BeforeCommit != nil {
		if err := options.BeforeCommit(synthetic); err != nil {
			for path := range createdTopLevel {
				_ = os.RemoveAll(path)
			}
			if rootCreated {
				_ = os.Remove(root)
			}
			return InitResult{}, err
		}
	}
	fileNames := make([]string, 0, len(files))
	for relative := range files {
		fileNames = append(fileNames, relative)
	}
	sort.Strings(fileNames)
	for _, relative := range fileNames {
		path, err := ResolveContained(root, relative)
		if err != nil {
			return InitResult{}, err
		}
		if data, err := os.ReadFile(path); err == nil {
			if relative == config.FileName && string(data) != string(files[relative]) && !initialized {
				return InitResult{}, fmt.Errorf("refusing to replace existing %s", config.FileName)
			}
			continue
		} else if !errors.Is(err, os.ErrNotExist) {
			return InitResult{}, err
		}
		if err := AtomicWriteFile(path, files[relative], 0o644); err != nil {
			return InitResult{}, err
		}
		created = append(created, filepath.ToSlash(relative))
	}
	gitignorePath := filepath.Join(root, ".gitignore")
	changed, err := ensureGitignore(gitignorePath, []string{".buda/", ".qmd/"})
	if err != nil {
		return InitResult{}, err
	}
	if changed {
		created = append(created, ".gitignore")
	}
	repository, err := Open(root)
	if err != nil {
		return InitResult{}, err
	}
	sort.Strings(created)
	return InitResult{Repository: repository, Created: created, Unchanged: len(created) == 0}, nil
}

func ensureGitignore(path string, required []string) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return false, fmt.Errorf("read .gitignore: %w", err)
	}
	lineEnding := "\n"
	if strings.Contains(string(data), "\r\n") {
		lineEnding = "\r\n"
	}
	normalized := strings.ReplaceAll(string(data), "\r\n", "\n")
	lines := strings.Split(normalized, "\n")
	present := map[string]bool{}
	for _, line := range lines {
		present[strings.TrimSpace(line)] = true
	}
	changed := false
	for _, entry := range required {
		if !present[entry] {
			if len(data) > 0 && !strings.HasSuffix(normalized, "\n") {
				normalized += "\n"
			}
			normalized += entry + "\n"
			present[entry] = true
			changed = true
		}
	}
	if !changed {
		return false, nil
	}
	if err := AtomicWriteFile(path, []byte(strings.ReplaceAll(normalized, "\n", lineEnding)), 0o644); err != nil {
		return false, err
	}
	return true, nil
}
