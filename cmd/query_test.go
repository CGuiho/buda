package cmd

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CGuiho/buda/internal/qmd"
	"github.com/CGuiho/buda/internal/repository"
)

type fakeQMD struct {
	matches []qmd.Match
}

func (*fakeQMD) CheckCompatibility(context.Context) (qmd.Compatibility, error) {
	return qmd.Compatibility{Version: qmd.Version{Major: 2, Minor: 6, Patch: 3}}, nil
}
func (*fakeQMD) EnsureProject(context.Context) error { return nil }
func (*fakeQMD) Ready(context.Context) (qmd.Compatibility, error) {
	return qmd.Compatibility{Version: qmd.Version{Major: 2, Minor: 6, Patch: 3}}, nil
}
func (*fakeQMD) Update(context.Context) (qmd.IndexResult, error) {
	return qmd.IndexResult{}, nil
}
func (*fakeQMD) Embed(context.Context) (qmd.IndexResult, error) {
	return qmd.IndexResult{}, nil
}
func (fake *fakeQMD) Search(context.Context, qmd.SearchOptions) ([]qmd.Match, error) {
	return fake.matches, nil
}
func (*fakeQMD) Get(context.Context, string) (qmd.Document, error) {
	return qmd.Document{}, nil
}
func (*fakeQMD) Status(context.Context) (qmd.Diagnostic, error) {
	return qmd.Diagnostic{}, nil
}
func (*fakeQMD) Doctor(context.Context) (qmd.Diagnostic, error) {
	return qmd.Diagnostic{}, nil
}

func TestQueryJSONNormalizesOKFEvidenceWithoutReranking(t *testing.T) {
	root := makeWiki(t)
	concept := `---
type: Decision
title: API policy
sources:
  - id: policy
    resource: /sources/policy.md
buda:
  schema_version: "1"
  uid: uid-1
  wiki_id: wiki-1
---
The API uses durable IDs.[^policy]

[^policy]: Policy
`
	writeTestFile(t, filepath.Join(root, "knowledge", "concepts", "policy.md"), concept)
	fake := &fakeQMD{matches: []qmd.Match{{Rank: 1, DocumentID: "#abc123", Score: 0.73, Path: "concepts/policy.md", Title: "API policy", Snippet: "durable IDs"}}}
	factory := func(repository.Repository) (QMDClient, error) { return fake, nil }
	var output bytes.Buffer
	deps := Dependencies{Options: &Options{Wiki: root, JSON: true}, Out: &output}
	command := NewQueryCommand(deps, factory)
	command.SetOut(&output)
	command.SetArgs([]string{"--text", "identifiers", "--mode", "lexical"})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	value := output.String()
	for _, expected := range []string{`"score": 0.73`, `"path": "concepts/policy.md"`, `"claim_footnote_joined": true`, `"wiki_id": "wiki-1"`} {
		if !strings.Contains(value, expected) {
			t.Fatalf("output missing %s:\n%s", expected, value)
		}
	}
}

func makeWiki(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	configuration := `schema: 1
wiki_id: wiki-1
bundle: knowledge
qmd:
  executable: qmd
  collection: buda-wiki
  project_directory: .qmd
derived: .buda
`
	writeTestFile(t, filepath.Join(root, "buda.yaml"), configuration)
	writeTestFile(t, filepath.Join(root, "knowledge", "index.md"), "---\nokf_version: \"0.2\"\n---\n# Index\n")
	return root
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
