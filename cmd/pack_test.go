package cmd

import (
	"archive/zip"
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestPackCommandCreatesCanonicalOnlyArchive(t *testing.T) {
	root := makeWiki(t)
	writeTestFile(t, filepath.Join(root, "outside.txt"), "not canonical")
	output := filepath.Join(root, ".buda", "wiki.zip")
	var stdout bytes.Buffer
	deps := Dependencies{Options: &Options{Wiki: root, JSON: true}, Out: &stdout}
	command := NewPackCommand(deps)
	command.SetOut(&stdout)
	command.SetArgs([]string{"--output", output})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), `"sharing_safety_certified": false`) {
		t.Fatalf("missing safety boundary: %s", stdout.String())
	}
	reader, err := zip.OpenReader(output)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	for _, file := range reader.File {
		if file.Name == "outside.txt" || file.Name == "buda.yaml" {
			t.Fatalf("archive included repository file %q", file.Name)
		}
	}
}
