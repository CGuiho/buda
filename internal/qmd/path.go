package qmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func containedPath(root, candidate string) (string, error) {
	root = filepath.Clean(root)
	candidate = filepath.Clean(candidate)
	relative, err := filepath.Rel(root, candidate)
	if err != nil {
		return "", err
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) || filepath.IsAbs(relative) {
		return "", fmt.Errorf("path %q escapes %q", candidate, root)
	}
	return candidate, nil
}

// ResolveResultPath converts a qmd display or virtual path into a contained,
// repository-relative canonical concept path.
func ResolveResultPath(bundleRoot, collection, qmdPath string) (string, error) {
	value := strings.TrimSpace(strings.ReplaceAll(qmdPath, "\\", "/"))
	prefix := "qmd://" + collection + "/"
	if strings.HasPrefix(value, "qmd://") {
		if !strings.HasPrefix(value, prefix) {
			return "", fmt.Errorf("qmd result belongs to another collection: %q", qmdPath)
		}
		value = strings.TrimPrefix(value, prefix)
	}
	if strings.HasPrefix(value, collection+"/") {
		value = strings.TrimPrefix(value, collection+"/")
	}
	bundleName := filepath.Base(filepath.Clean(bundleRoot))
	if strings.HasPrefix(value, bundleName+"/") {
		value = strings.TrimPrefix(value, bundleName+"/")
	}
	if value == "" || filepath.IsAbs(filepath.FromSlash(value)) {
		return "", fmt.Errorf("invalid qmd result path %q", qmdPath)
	}
	absolute := filepath.Join(bundleRoot, filepath.FromSlash(value))
	if _, err := containedPath(bundleRoot, absolute); err != nil {
		return "", fmt.Errorf("reject qmd result path %q: %w", qmdPath, err)
	}
	if _, err := os.Lstat(absolute); err == nil {
		resolvedBundle, bundleErr := filepath.EvalSymlinks(bundleRoot)
		resolvedResult, resultErr := filepath.EvalSymlinks(absolute)
		if bundleErr != nil || resultErr != nil {
			return "", fmt.Errorf("resolve qmd result path %q", qmdPath)
		}
		if _, err := containedPath(resolvedBundle, resolvedResult); err != nil {
			return "", fmt.Errorf("reject qmd result symlink %q: %w", qmdPath, err)
		}
	}
	return filepath.ToSlash(value), nil
}
