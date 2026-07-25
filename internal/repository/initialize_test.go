package repository

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestInitializeCreatesPortableLayoutAndIsIdempotent(t *testing.T) {
	root := filepath.Join(t.TempDir(), "wiki")
	result, err := Initialize(root, InitOptions{WikiID: "wiki-test", Now: time.Date(2026, 7, 26, 1, 2, 3, 0, time.UTC)})
	if err != nil {
		t.Fatal(err)
	}
	for _, relative := range []string{"buda.yaml", ".gitignore", "knowledge/index.md", "knowledge/log.md", "knowledge/concepts", "knowledge/sources", "knowledge/references/raw", ".qmd", ".buda"} {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(relative))); err != nil {
			t.Errorf("missing %s: %v", relative, err)
		}
	}
	if result.Repository.Config.WikiID != "wiki-test" {
		t.Fatalf("result = %+v", result)
	}
	second, err := Initialize(root, InitOptions{WikiID: "wiki-test", Now: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	if !second.Unchanged {
		t.Fatalf("second init changed files: %+v", second.Created)
	}
}

func TestInitializeRefusesIncompatibleTargetAndWikiIDChange(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("existing"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Initialize(root, InitOptions{WikiID: "wiki"}); err == nil {
		t.Fatal("incompatible non-empty target accepted")
	}
	root = filepath.Join(t.TempDir(), "wiki")
	if _, err := Initialize(root, InitOptions{WikiID: "first"}); err != nil {
		t.Fatal(err)
	}
	if _, err := Initialize(root, InitOptions{WikiID: "second"}); err == nil {
		t.Fatal("wiki_id mutation accepted")
	}
}
