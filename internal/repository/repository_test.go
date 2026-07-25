package repository

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/CGuiho/buda/internal/config"
)

func TestOpenRequiresExplicitExactWiki(t *testing.T) {
	if _, err := Open(""); err == nil {
		t.Fatal("Open accepted implicit wiki")
	}
	root := t.TempDir()
	configuration, err := config.Marshal(config.Default("wiki-1"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, config.FileName), configuration, 0o644); err != nil {
		t.Fatal(err)
	}
	repository, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	if repository.Root != root || repository.Config.WikiID != "wiki-1" {
		t.Fatalf("Open() = %+v", repository)
	}
	child := filepath.Join(root, "knowledge")
	if err := os.Mkdir(child, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(child); err == nil {
		t.Fatal("Open searched an ancestor config")
	}
}

func TestResolveContainedRejectsEscapes(t *testing.T) {
	root := t.TempDir()
	for _, value := range []string{"../outside", filepath.Join("nested", "..", "..", "outside")} {
		if _, err := ResolveContained(root, value); err == nil {
			t.Fatalf("ResolveContained accepted %q", value)
		}
	}
	inside, err := ResolveContained(root, filepath.Join("knowledge", "concepts"))
	if err != nil {
		t.Fatal(err)
	}
	if inside != filepath.Join(root, "knowledge", "concepts") {
		t.Fatalf("inside = %q", inside)
	}
}

func TestResolveContainedRejectsExistingSymlinkEscape(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("creating symlinks may require Windows developer mode")
	}
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "escape")); err != nil {
		t.Fatal(err)
	}
	if _, err := ResolveContained(root, filepath.Join("escape", "file.md")); err == nil {
		t.Fatal("ResolveContained accepted symlink escape")
	}
}
