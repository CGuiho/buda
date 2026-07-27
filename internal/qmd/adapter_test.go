package qmd

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type queuedRunner struct {
	requests []Request
	results  []ProcessResult
}

func (runner *queuedRunner) Run(_ context.Context, request Request) (ProcessResult, error) {
	runner.requests = append(runner.requests, request)
	result := runner.results[0]
	runner.results = runner.results[1:]
	return result, nil
}

func testAdapter(t *testing.T, runner Runner) *Adapter {
	t.Helper()
	root := t.TempDir()
	bundle := filepath.Join(root, "knowledge")
	project := filepath.Join(root, ".qmd")
	if err := os.MkdirAll(bundle, 0o755); err != nil {
		t.Fatal(err)
	}
	adapter, err := New(Config{WikiRoot: root, BundleRoot: bundle, ProjectDirectory: project, Collection: "buda-wiki"}, runner)
	if err != nil {
		t.Fatal(err)
	}
	return adapter
}

func TestSearchUsesExplicitCollectionAndPreservesOrderAndScore(t *testing.T) {
	runner := &queuedRunner{results: []ProcessResult{{Stdout: []byte(`[
  {"docid":"#second","score":0.73,"file":"knowledge/concepts/b.md","title":"B","snippet":"two"},
  {"docid":"#first","score":0.51,"file":"knowledge/concepts/a.md","title":"A","snippet":"one"}
]`)}}}
	adapter := testAdapter(t, runner)
	matches, err := adapter.Search(context.Background(), SearchOptions{Mode: ModeLexical, Text: "policy", Limit: 2})
	if err != nil {
		t.Fatal(err)
	}
	if got := []string{matches[0].DocumentID, matches[1].DocumentID}; !reflect.DeepEqual(got, []string{"#second", "#first"}) {
		t.Fatalf("order changed: %v", got)
	}
	if matches[0].Score != 0.73 || matches[1].Score != 0.51 {
		t.Fatalf("scores changed: %#v", matches)
	}
	wantArguments := []string{"search", "--format", "json", "-c", "buda-wiki", "-n", "2", "policy"}
	if !reflect.DeepEqual(runner.requests[0].Arguments, wantArguments) {
		t.Fatalf("arguments = %#v, want %#v", runner.requests[0].Arguments, wantArguments)
	}
}

func TestSearchParsesPinnedQMD253Fixture(t *testing.T) {
	fixture, err := os.ReadFile(filepath.Join("testdata", "search-2.5.3.json"))
	if err != nil {
		t.Fatal(err)
	}
	runner := &queuedRunner{results: []ProcessResult{{Stdout: fixture}}}
	adapter := testAdapter(t, runner)
	matches, err := adapter.Search(context.Background(), SearchOptions{Mode: ModeHybrid, Text: "API policy"})
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 1 || matches[0].Score != 0.73 || matches[0].Path != "concepts/api-policy.md" || matches[0].Line != 12 {
		t.Fatalf("matches = %#v", matches)
	}
}

func TestSearchRejectsEscapedOrForeignPaths(t *testing.T) {
	for _, path := range []string{"../outside.md", "qmd://another/secret.md"} {
		t.Run(path, func(t *testing.T) {
			runner := &queuedRunner{results: []ProcessResult{{Stdout: []byte(`[{"docid":"#x","score":1,"file":` + quote(path) + `}]`)}}}
			adapter := testAdapter(t, runner)
			if _, err := adapter.Search(context.Background(), SearchOptions{Text: "x"}); err == nil {
				t.Fatal("expected containment error")
			}
		})
	}
}

func TestGetUsesCollectionQualifiedVirtualPath(t *testing.T) {
	runner := &queuedRunner{results: []ProcessResult{{Stdout: []byte(`[{"file":"knowledge/concepts/a.md","title":"A","body":"body"}]`)}}}
	adapter := testAdapter(t, runner)
	document, err := adapter.Get(context.Background(), "concepts/a.md")
	if err != nil {
		t.Fatal(err)
	}
	if document.Path != "concepts/a.md" || document.Body != "body" {
		t.Fatalf("document = %#v", document)
	}
	if got := runner.requests[0].Arguments[1]; got != "qmd://buda-wiki/concepts/a.md" {
		t.Fatalf("get target = %q", got)
	}
}

func TestGetUsesQMDGetForStableDocumentID(t *testing.T) {
	output := "qmd://buda-wiki/concepts/a.md  #a1b2c3\nFolder Context: Architecture decisions\n---\n\nbody\n"
	runner := &queuedRunner{results: []ProcessResult{{Stdout: []byte(output)}}}
	adapter := testAdapter(t, runner)
	document, err := adapter.Get(context.Background(), "#a1b2c3")
	if err != nil {
		t.Fatal(err)
	}
	if document.Path != "concepts/a.md" || document.Context != "Architecture decisions" || document.Body != "body\n" {
		t.Fatalf("document = %#v", document)
	}
	wantArguments := []string{"get", "#a1b2c3", "--no-line-numbers"}
	if !reflect.DeepEqual(runner.requests[0].Arguments, wantArguments) {
		t.Fatalf("arguments = %#v, want %#v", runner.requests[0].Arguments, wantArguments)
	}
}

func TestGetParsesPinnedQMD253Fixture(t *testing.T) {
	fixture, err := os.ReadFile(filepath.Join("testdata", "multi-get-2.5.3.json"))
	if err != nil {
		t.Fatal(err)
	}
	runner := &queuedRunner{results: []ProcessResult{{Stdout: fixture}}}
	adapter := testAdapter(t, runner)
	document, err := adapter.Get(context.Background(), "concepts/api-policy.md")
	if err != nil {
		t.Fatal(err)
	}
	if document.Path != "concepts/api-policy.md" || !strings.Contains(document.Body, "durable external identifiers") {
		t.Fatalf("document = %#v", document)
	}
}

func TestVersionRange(t *testing.T) {
	for _, test := range []struct {
		output string
		ok     bool
	}{{"qmd 2.5.0\n", true}, {"qmd 2.5.3 (5323277)\n", true}, {"qmd 2.9.3 (abc123)\n", true}, {"qmd 2.4.9\n", false}, {"qmd 3.0.0\n", false}} {
		t.Run(test.output, func(t *testing.T) {
			runner := &queuedRunner{results: []ProcessResult{{Stdout: []byte(test.output)}}}
			adapter := testAdapter(t, runner)
			_, err := adapter.Version(context.Background())
			if (err == nil) != test.ok {
				t.Fatalf("error = %v, ok = %v", err, test.ok)
			}
		})
	}
}

func TestCompatibilityRequiresEveryUsedCapability(t *testing.T) {
	help := "qmd search\nqmd vsearch\nqmd query\nqmd get\nqmd multi-get\nqmd init\nqmd status\nqmd update\nqmd embed\nqmd collection add/list/remove/rename/show\n--format\n-c, --collection"
	runner := &queuedRunner{results: []ProcessResult{{Stdout: []byte("qmd 2.5.3\n")}, {Stdout: []byte(help)}}}
	adapter := testAdapter(t, runner)
	compatibility, err := adapter.CheckCompatibility(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if compatibility.Version.Major != 2 || len(compatibility.Capabilities) != 12 {
		t.Fatalf("compatibility = %#v", compatibility)
	}
}

func TestCompatibilityRejectsRuntimeWithoutQMDGet(t *testing.T) {
	help := "qmd search\nqmd vsearch\nqmd query\nqmd multi-get\nqmd init\nqmd status\nqmd update\nqmd embed\nqmd collection add/list/remove/rename/show\n--format\n-c, --collection"
	runner := &queuedRunner{results: []ProcessResult{{Stdout: []byte("qmd 2.5.3\n")}, {Stdout: []byte(help)}}}
	adapter := testAdapter(t, runner)
	_, err := adapter.CheckCompatibility(context.Background())
	if err == nil || !strings.Contains(err.Error(), `lacks required capability "qmd get"`) {
		t.Fatalf("error = %v", err)
	}
}

func TestCompatibilityUsesExistingParentBeforeWikiInitialization(t *testing.T) {
	parent := t.TempDir()
	wiki := filepath.Join(parent, "not-created", "wiki")
	runner := &queuedRunner{results: []ProcessResult{{Stdout: []byte("qmd 2.5.3\n")}}}
	adapter, err := New(Config{
		WikiRoot: wiki, BundleRoot: filepath.Join(wiki, "knowledge"),
		ProjectDirectory: filepath.Join(wiki, ".qmd"), Collection: "buda-wiki",
	}, runner)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := adapter.Version(context.Background()); err != nil {
		t.Fatal(err)
	}
	if runner.requests[0].Directory != parent {
		t.Fatalf("compatibility directory = %q, want existing parent %q", runner.requests[0].Directory, parent)
	}
}

func TestValidateCollectionRejectsAdditionalProjectCollections(t *testing.T) {
	runner := &queuedRunner{}
	adapter := testAdapter(t, runner)
	configuration := "collections:\n  buda-wiki:\n    path: ../knowledge\n  another:\n    path: ../other\n"
	if err := os.MkdirAll(filepath.Dir(filepath.Join(adapter.config.ProjectDirectory, "index.yml")), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(adapter.config.ProjectDirectory, "index.yml"), []byte(configuration), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := adapter.ValidateCollection(context.Background()); err == nil {
		t.Fatal("expected exactly-one-collection validation error")
	}
}

func TestValidateCollectionAcceptsOneContainedRelativeMapping(t *testing.T) {
	helpOutput := "Collection: buda-wiki\n Path: knowledge\n Pattern: **/*.md\n"
	runner := &queuedRunner{results: []ProcessResult{{Stdout: []byte(helpOutput)}}}
	adapter := testAdapter(t, runner)
	configuration := "collections:\n  buda-wiki:\n    path: ../knowledge\n"
	if err := os.MkdirAll(adapter.config.ProjectDirectory, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(adapter.config.ProjectDirectory, "index.yml"), []byte(configuration), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := adapter.ValidateCollection(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestNormalizeDoctorAndStatusOutput(t *testing.T) {
	doctorOutput, err := os.ReadFile(filepath.Join("testdata", "doctor-2.5.3.txt"))
	if err != nil {
		t.Fatal(err)
	}
	checks, warnings, failures := parseDoctorChecks(string(doctorOutput))
	if checks != 10 || warnings != 1 || failures != 0 {
		t.Fatalf("doctor checks=%d warnings=%d failures=%d", checks, warnings, failures)
	}
	statusOutput, err := os.ReadFile(filepath.Join("testdata", "status-2.5.3.txt"))
	if err != nil {
		t.Fatal(err)
	}
	documents, vectors, pending := parseStatusCounts(string(statusOutput))
	if documents != 1234 || vectors != 100 || pending != 12 {
		t.Fatalf("status documents=%d vectors=%d pending=%d", documents, vectors, pending)
	}
}

func quote(value string) string {
	result := `"`
	for _, character := range value {
		if character == '"' || character == '\\' {
			result += `\`
		}
		result += string(character)
	}
	return result + `"`
}
