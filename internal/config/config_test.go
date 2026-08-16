package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDecodeStrictAndSemanticValidation(t *testing.T) {
	valid := `schema: 1
wiki_id: research-notes
bundle: knowledge
qmd:
  executable: qmd
  collection: buda-wiki
  project_directory: .qmd
derived: .buda
`
	configuration, err := Decode(strings.NewReader(valid))
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if configuration.WikiID != "research-notes" {
		t.Fatalf("WikiID = %q", configuration.WikiID)
	}

	cases := []string{
		strings.Replace(valid, "derived: .buda", "derived: .buda\nunknown: true", 1),
		strings.Replace(valid, "bundle: knowledge", "bundle: ../outside", 1),
		strings.Replace(valid, "schema: 1", "schema: 2", 1),
		valid + "---\nschema: 1\n",
	}
	for _, input := range cases {
		if _, err := Decode(strings.NewReader(input)); err == nil {
			t.Fatalf("Decode() unexpectedly accepted:\n%s", input)
		}
	}
}

func TestResolvePrecedence(t *testing.T) {
	root := t.TempDir()
	cwd := filepath.Join(root, "project")
	home := filepath.Join(root, "home")
	if err := os.MkdirAll(filepath.Join(home, ".guiho", "buda"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(cwd, 0o755); err != nil {
		t.Fatal(err)
	}
	global := filepath.Join(home, ".guiho", "buda", GlobalFileName)
	if err := os.WriteFile(global, []byte("global"), 0o644); err != nil {
		t.Fatal(err)
	}
	resolved, err := Resolve("", cwd, home)
	if err != nil || resolved != global {
		t.Fatalf("Resolve global = %q, %v", resolved, err)
	}
	project := filepath.Join(cwd, FileName)
	if err := os.WriteFile(project, []byte("project"), 0o644); err != nil {
		t.Fatal(err)
	}
	resolved, err = Resolve("", cwd, home)
	if err != nil || resolved != project {
		t.Fatalf("Resolve project = %q, %v", resolved, err)
	}
}

func TestLoadEffectiveUsesGlobalBaselineAndProjectOverrides(t *testing.T) {
	home := t.TempDir()
	projectRoot := filepath.Join(t.TempDir(), "wiki")
	if err := os.MkdirAll(filepath.Dir(GlobalPath(home)), 0o755); err != nil {
		t.Fatal(err)
	}
	global, err := MarshalGlobal(GlobalConfig{
		Schema:  CurrentSchema,
		Bundle:  stringPtr("shared"),
		Derived: stringPtr(".derived"),
		QMD:     &QMDConfig{Executable: "qmd", Collection: "global", ProjectDirectory: ".qmd"},
		Agent:   &AgentConfig{Evolution: EvolutionConfig{Upgrade: PolicyAlwaysProceed}},
	}, "0.2.0")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(GlobalPath(home), global, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(projectRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	project, err := MarshalProject(ProjectConfig{Schema: CurrentSchema, WikiID: "selected", QMD: &QMDConfig{Collection: "project"}}, "0.2.0")
	if err != nil {
		t.Fatal(err)
	}
	projectPath := filepath.Join(projectRoot, FileName)
	if err := os.WriteFile(projectPath, project, 0o644); err != nil {
		t.Fatal(err)
	}
	effective, err := LoadEffective(projectPath, home)
	if err != nil {
		t.Fatal(err)
	}
	if effective.Bundle != "shared" || effective.QMD.Collection != "project" || effective.WikiID != "selected" || effective.EffectiveAgent().Evolution.Upgrade != PolicyAlwaysProceed {
		t.Fatalf("effective config = %#v", effective)
	}
}

func stringPtr(value string) *string { return &value }
