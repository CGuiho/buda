package pack

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func TestCreateIsDeterministicAndContainsOnlyBundle(t *testing.T) {
	root := t.TempDir()
	bundle := filepath.Join(root, "knowledge")
	mustWrite(t, filepath.Join(bundle, "concepts", "b.md"), "B\n")
	mustWrite(t, filepath.Join(bundle, "index.md"), "# Index\n")
	mustWrite(t, filepath.Join(root, "buda.yaml"), "secret: outside\n")
	mustWrite(t, filepath.Join(root, ".qmd", "index.sqlite"), "derived")
	first := filepath.Join(root, ".buda", "first.zip")
	second := filepath.Join(root, ".buda", "second.zip")
	for _, output := range []string{first, second} {
		result, err := Create(Options{WikiRoot: root, BundleRoot: bundle, BundleName: "knowledge", WikiID: "wiki-1", Output: output})
		if err != nil {
			t.Fatal(err)
		}
		if result.Files != 2 {
			t.Fatalf("files = %d", result.Files)
		}
	}
	firstBytes, _ := os.ReadFile(first)
	secondBytes, _ := os.ReadFile(second)
	if !bytes.Equal(firstBytes, secondBytes) {
		t.Fatal("identical canonical bundle produced different archives")
	}
	reader, err := zip.OpenReader(first)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	var names []string
	for _, file := range reader.File {
		names = append(names, file.Name)
	}
	sort.Strings(names)
	want := []string{"checksums.txt", "knowledge/concepts/b.md", "knowledge/index.md", "manifest.json"}
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("archive names = %#v, want %#v", names, want)
	}
}

func TestCreateRejectsOutputInsideBundle(t *testing.T) {
	root := t.TempDir()
	bundle := filepath.Join(root, "knowledge")
	mustWrite(t, filepath.Join(bundle, "index.md"), "# Index\n")
	_, err := Create(Options{WikiRoot: root, BundleRoot: bundle, WikiID: "wiki", Output: filepath.Join(bundle, "pack.zip")})
	if err == nil {
		t.Fatal("expected output containment error")
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
