package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CGuiho/buda/internal/agent"
	"github.com/CGuiho/buda/internal/config"
	"github.com/CGuiho/buda/internal/qmd"
	"github.com/CGuiho/buda/internal/repository"
)

type domainFakeQMD struct {
	ensure int
	update int
	search []qmd.Match
}

func (*domainFakeQMD) CheckCompatibility(context.Context) (qmd.Compatibility, error) {
	return qmd.Compatibility{Version: qmd.Version{Major: 2, Minor: 6, Patch: 3}}, nil
}
func (fake *domainFakeQMD) EnsureProject(context.Context) error { fake.ensure++; return nil }
func (*domainFakeQMD) Ready(context.Context) (qmd.Compatibility, error) {
	return qmd.Compatibility{Version: qmd.Version{Major: 2, Minor: 6, Patch: 3}}, nil
}
func (fake *domainFakeQMD) Update(context.Context) (qmd.IndexResult, error) {
	fake.update++
	return qmd.IndexResult{Capability: "index", Output: "ok"}, nil
}
func (*domainFakeQMD) Embed(context.Context) (qmd.IndexResult, error) { return qmd.IndexResult{}, nil }
func (fake *domainFakeQMD) Search(context.Context, qmd.SearchOptions) ([]qmd.Match, error) {
	return append([]qmd.Match(nil), fake.search...), nil
}
func (*domainFakeQMD) Get(context.Context, string) (qmd.Document, error) { return qmd.Document{}, nil }
func (*domainFakeQMD) Status(context.Context) (qmd.Diagnostic, error) {
	return qmd.Diagnostic{Capability: "status", Version: "2.6.0", Output: "ready"}, nil
}
func (*domainFakeQMD) Doctor(context.Context) (qmd.Diagnostic, error) { return qmd.Diagnostic{}, nil }

func TestInitCommandPreflightsQMDAndInstallsAgentResources(t *testing.T) {
	wiki := filepath.Join(t.TempDir(), "wiki")
	home := t.TempDir()
	client := &domainFakeQMD{}
	output := &bytes.Buffer{}
	globalPath := config.GlobalPath(home)
	if err := os.MkdirAll(filepath.Dir(globalPath), 0o755); err != nil {
		t.Fatal(err)
	}
	globalData, err := config.MarshalGlobal(config.DefaultGlobal(), "test")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(globalPath, globalData, 0o644); err != nil {
		t.Fatal(err)
	}
	deps := domainDeps(output, home)
	root := NewRootCommand(deps, BuildInfo{Version: "test"}, NewInitCommand(deps, func(repository.Repository) (QMDClient, error) { return client, nil }))
	root.SetArgs([]string{"--wiki", wiki, "--json", "init", "--wiki-id", "wiki-test"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if client.ensure != 1 {
		t.Fatalf("EnsureProject calls = %d", client.ensure)
	}
	for _, path := range []string{
		filepath.Join(wiki, "buda.yaml"), filepath.Join(wiki, "knowledge", "index.md"),
		filepath.Join(home, ".agents", "skills", agent.SkillID, "SKILL.md"),
		filepath.Join(home, ".claude", "skills", agent.SkillID, "SKILL.md"),
		filepath.Join(wiki, "AGENTS.md"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("missing initialized path %s: %v", path, err)
		}
	}
	var document map[string]any
	if err := json.Unmarshal(output.Bytes(), &document); err != nil {
		t.Fatalf("JSON output: %v\n%s", err, output)
	}
}

func TestCaptureCommandRefreshesQMDAndEmitsOneJSONDocument(t *testing.T) {
	wiki := initializedWiki(t)
	client := &domainFakeQMD{}
	output := &bytes.Buffer{}
	deps := domainDeps(output, t.TempDir())
	root := NewRootCommand(deps, BuildInfo{Version: "test"}, NewCaptureCommand(deps, func(repository.Repository) (QMDClient, error) { return client, nil }))
	root.SetArgs([]string{
		"--wiki", wiki, "--json", "capture", "--target", "concepts/decision.md",
		"--title", "Decision", "--description", "A decision.", "--type", "Decision",
		"--actor", "human:owner", "--text", "Use stable identifiers.",
	})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if client.ensure != 1 || client.update != 1 {
		t.Fatalf("qmd calls ensure=%d update=%d", client.ensure, client.update)
	}
	decoder := json.NewDecoder(bytes.NewReader(output.Bytes()))
	var document map[string]any
	if err := decoder.Decode(&document); err != nil {
		t.Fatal(err)
	}
	var extra any
	if err := decoder.Decode(&extra); err == nil {
		t.Fatal("capture emitted more than one JSON document")
	}
}

func TestStatusKeepsCanonicalAndQMDResultsSeparate(t *testing.T) {
	wiki := initializedWiki(t)
	output := &bytes.Buffer{}
	deps := domainDeps(output, t.TempDir())
	root := NewRootCommand(deps, BuildInfo{Version: "test"}, NewStatusCommand(deps, func(repository.Repository) (QMDClient, error) { return &domainFakeQMD{}, nil }))
	root.SetArgs([]string{"--wiki", wiki, "--json", "status"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	var document map[string]any
	if err := json.Unmarshal(output.Bytes(), &document); err != nil {
		t.Fatal(err)
	}
	if document["canonical"] == nil || document["qmd"] == nil {
		t.Fatalf("status output merged or omitted domains: %#v", document)
	}
}

func domainDeps(output *bytes.Buffer, home string) Dependencies {
	return Dependencies{
		In: strings.NewReader("yes\n"), Out: output, Err: &bytes.Buffer{}, Options: &Options{},
		Interactive:         func() bool { return true },
		HomeDir:             func() (string, error) { return home, nil },
		Agents:              agent.NewService(agent.DefaultResources(), agent.WithHomeDir(func() (string, error) { return home, nil })),
		Executable:          func() (string, error) { return "buda", nil },
		ScheduleMaintenance: func(string, string) error { return nil },
	}
}

func initializedWiki(t *testing.T) string {
	t.Helper()
	wiki := filepath.Join(t.TempDir(), "wiki")
	if _, err := repository.Initialize(wiki, repository.InitOptions{WikiID: "wiki-test"}); err != nil {
		t.Fatal(err)
	}
	return wiki
}
