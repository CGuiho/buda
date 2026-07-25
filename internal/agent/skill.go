package agent

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/CGuiho/buda/internal/repository"
)

type SkillInstallation struct {
	Tool      string `json:"tool"`
	Scope     string `json:"scope"`
	Path      string `json:"path"`
	Installed bool   `json:"installed"`
	Current   bool   `json:"current"`
	Version   string `json:"version,omitempty"`
	Digest    string `json:"digest,omitempty"`
	Changed   bool   `json:"changed,omitempty"`
}

func (s *Service) SkillInstallations(local bool, wiki string) ([]SkillInstallation, error) {
	root, scope, err := s.scopeRoot(local, wiki)
	if err != nil {
		return nil, err
	}
	bundled, err := s.Skill()
	if err != nil {
		return nil, err
	}
	results := make([]SkillInstallation, 0, 2)
	for _, tool := range []string{"agents", "claude"} {
		path := filepath.Join(root, "."+tool, "skills", SkillID)
		result := SkillInstallation{Tool: tool, Scope: scope, Path: path}
		info, statErr := os.Stat(path)
		if statErr == nil && info.IsDir() {
			result.Installed = true
			result.Digest, result.Version = installedSkillMetadata(path)
			result.Current = result.Digest == bundled.Digest
		} else if statErr != nil && !os.IsNotExist(statErr) {
			return nil, mutation("inspect installed Buda skill", statErr)
		}
		results = append(results, result)
	}
	return results, nil
}

func (s *Service) InstallSkill(local bool, wiki string) ([]SkillInstallation, error) {
	root, scope, err := s.scopeRoot(local, wiki)
	if err != nil {
		return nil, err
	}
	stages := make([]directoryStage, 0, 2)
	for _, tool := range []string{"agents", "claude"} {
		destination := filepath.Join(root, "."+tool, "skills", SkillID)
		stage, err := prepareDirectoryStage(destination, ".buda-skill-new-*", func(path string) error {
			return copyEmbeddedTree(s.resources.Skill, path)
		})
		if err != nil {
			cleanupStages(stages)
			return nil, err
		}
		stage.tool = tool
		stages = append(stages, stage)
	}
	if err := commitStages(stages); err != nil {
		return nil, err
	}
	results, err := s.SkillInstallations(local, wiki)
	if err != nil {
		return nil, err
	}
	for index := range results {
		results[index].Changed = stages[index].changed
		results[index].Scope = scope
	}
	return results, nil
}

func (s *Service) UninstallSkill(local bool, wiki string) ([]SkillInstallation, error) {
	root, scope, err := s.scopeRoot(local, wiki)
	if err != nil {
		return nil, err
	}
	stages := make([]directoryStage, 0, 2)
	for _, tool := range []string{"agents", "claude"} {
		destination := filepath.Join(root, "."+tool, "skills", SkillID)
		stage, err := prepareRemovalStage(destination)
		if err != nil {
			cleanupStages(stages)
			return nil, err
		}
		stage.tool = tool
		stages = append(stages, stage)
	}
	if err := commitRemovals(stages); err != nil {
		return nil, err
	}
	results := make([]SkillInstallation, 0, 2)
	for _, stage := range stages {
		results = append(results, SkillInstallation{
			Tool: stage.tool, Scope: scope, Path: stage.destination,
			Installed: false, Current: false, Changed: stage.existed,
		})
	}
	return results, nil
}

func (s *Service) scopeRoot(local bool, wiki string) (string, string, error) {
	if local {
		if strings.TrimSpace(wiki) == "" {
			return "", "", usage("--local requires an explicit --wiki")
		}
		selected, err := repository.Open(wiki)
		if err != nil {
			return "", "", repositoryError("open selected wiki for local skill operation", err)
		}
		return selected.Root, "local", nil
	}
	resolver := s.homeDir
	if resolver == nil {
		resolver = os.UserHomeDir
	}
	home, err := resolver()
	if err != nil {
		return "", "", mutation("resolve home directory", err)
	}
	return filepath.Clean(home), "global", nil
}

type directoryStage struct {
	tool        string
	destination string
	staged      string
	backup      string
	existed     bool
	changed     bool
	committed   bool
}

func prepareDirectoryStage(destination, pattern string, populate func(string) error) (directoryStage, error) {
	stage := directoryStage{destination: destination}
	parent := filepath.Dir(destination)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return stage, mutation("create skill parent directory", err)
	}
	temporary, err := os.MkdirTemp(parent, pattern)
	if err != nil {
		return stage, mutation("create staged skill directory", err)
	}
	stage.staged = temporary
	if err := populate(temporary); err != nil {
		_ = os.RemoveAll(temporary)
		return stage, err
	}
	if info, statErr := os.Stat(destination); statErr == nil {
		if !info.IsDir() {
			_ = os.RemoveAll(temporary)
			return stage, mutation("installed Buda skill path is not a directory", nil)
		}
		stage.existed = true
	} else if !os.IsNotExist(statErr) {
		_ = os.RemoveAll(temporary)
		return stage, mutation("inspect installed Buda skill", statErr)
	}
	stage.backup, err = reserveSibling(parent, ".buda-skill-backup-*")
	if err != nil {
		_ = os.RemoveAll(temporary)
		return stage, err
	}
	newDigest, err := directoryDigest(temporary)
	if err != nil {
		cleanupStages([]directoryStage{stage})
		return stage, mutation("digest staged Buda skill", err)
	}
	oldDigest := ""
	if stage.existed {
		oldDigest, _ = directoryDigest(destination)
	}
	stage.changed = newDigest != oldDigest
	return stage, nil
}

func prepareRemovalStage(destination string) (directoryStage, error) {
	stage := directoryStage{destination: destination}
	info, err := os.Stat(destination)
	if os.IsNotExist(err) {
		return stage, nil
	}
	if err != nil {
		return stage, mutation("inspect installed Buda skill", err)
	}
	if !info.IsDir() {
		return stage, mutation("installed Buda skill path is not a directory", nil)
	}
	stage.existed = true
	stage.changed = true
	stage.backup, err = reserveSibling(filepath.Dir(destination), ".buda-skill-remove-*")
	return stage, err
}

func commitStages(stages []directoryStage) error {
	for index := range stages {
		stage := &stages[index]
		if !stage.changed {
			_ = os.RemoveAll(stage.staged)
			stage.staged = ""
			continue
		}
		if stage.existed {
			if err := os.Rename(stage.destination, stage.backup); err != nil {
				_ = rollbackStages(stages)
				return mutation("stage installed Buda skill", err)
			}
		}
		if err := os.Rename(stage.staged, stage.destination); err != nil {
			_ = rollbackStages(stages)
			return mutation("install staged Buda skill", err)
		}
		stage.staged = ""
		stage.committed = true
	}
	cleanupStages(stages)
	return nil
}

func commitRemovals(stages []directoryStage) error {
	for index := range stages {
		stage := &stages[index]
		if !stage.existed {
			continue
		}
		if err := os.Rename(stage.destination, stage.backup); err != nil {
			_ = rollbackRemovals(stages)
			return mutation("stage Buda skill removal", err)
		}
		stage.committed = true
	}
	cleanupStages(stages)
	return nil
}

func rollbackStages(stages []directoryStage) error {
	var failures []string
	for index := len(stages) - 1; index >= 0; index-- {
		stage := &stages[index]
		if stage.committed {
			if err := os.RemoveAll(stage.destination); err != nil {
				failures = append(failures, err.Error())
			}
		}
		if stage.existed {
			if _, err := os.Stat(stage.backup); err == nil {
				if err := os.Rename(stage.backup, stage.destination); err != nil {
					failures = append(failures, err.Error())
				}
			}
		}
	}
	cleanupStages(stages)
	if len(failures) > 0 {
		return fmt.Errorf("%s", strings.Join(failures, "; "))
	}
	return nil
}

func rollbackRemovals(stages []directoryStage) error {
	var failures []string
	for index := len(stages) - 1; index >= 0; index-- {
		stage := &stages[index]
		if stage.committed {
			if err := os.Rename(stage.backup, stage.destination); err != nil {
				failures = append(failures, err.Error())
			}
		}
	}
	if len(failures) > 0 {
		return fmt.Errorf("%s", strings.Join(failures, "; "))
	}
	return nil
}

func cleanupStages(stages []directoryStage) {
	for _, stage := range stages {
		if stage.staged != "" {
			_ = os.RemoveAll(stage.staged)
		}
		if stage.backup != "" {
			_ = os.RemoveAll(stage.backup)
		}
	}
}

func reserveSibling(parent, pattern string) (string, error) {
	path, err := os.MkdirTemp(parent, pattern)
	if err != nil {
		return "", mutation("reserve skill transaction path", err)
	}
	if err := os.Remove(path); err != nil {
		return "", mutation("release skill transaction path", err)
	}
	return path, nil
}

func copyEmbeddedTree(source fs.FS, destination string) error {
	return fs.WalkDir(source, ".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return mutation("walk embedded Buda skill", walkErr)
		}
		if path == "." {
			return nil
		}
		target := filepath.Join(destination, filepath.FromSlash(path))
		if entry.IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return mutation("create staged Buda skill directory", err)
			}
			return nil
		}
		content, err := fs.ReadFile(source, path)
		if err != nil {
			return mutation("read embedded Buda skill file", err)
		}
		if err := os.WriteFile(target, content, 0o644); err != nil {
			return mutation("write staged Buda skill file", err)
		}
		return nil
	})
}

func directoryDigest(root string) (string, error) {
	var paths []string
	if err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() {
			paths = append(paths, path)
		}
		return nil
	}); err != nil {
		return "", err
	}
	sort.Strings(paths)
	hash := sha256.New()
	for _, path := range paths {
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return "", err
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		_, _ = hash.Write([]byte(filepath.ToSlash(relative)))
		_, _ = hash.Write([]byte{0})
		_, _ = hash.Write(content)
		_, _ = hash.Write([]byte{0})
	}
	return "sha256:" + hex.EncodeToString(hash.Sum(nil)), nil
}

func installedSkillMetadata(path string) (digest, version string) {
	digest, _ = directoryDigest(path)
	content, err := os.ReadFile(filepath.Join(path, "SKILL.md"))
	if err != nil {
		return digest, ""
	}
	frontmatter, _, ok := splitFrontmatter(string(content))
	if !ok {
		return digest, ""
	}
	for _, line := range strings.Split(frontmatter, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "version:") {
			return digest, strings.Trim(strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "version:")), "\"")
		}
	}
	return digest, ""
}
