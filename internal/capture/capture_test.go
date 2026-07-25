package capture

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/CGuiho/buda/internal/health"
	"github.com/CGuiho/buda/internal/repository"
)

func TestCaptureCreatesCitedConceptAndRetryIsIdempotent(t *testing.T) {
	wiki := filepath.Join(t.TempDir(), "wiki")
	if _, err := repository.Initialize(wiki, repository.InitOptions{WikiID: "wiki"}); err != nil {
		t.Fatal(err)
	}
	selected, err := repository.Open(wiki)
	if err != nil {
		t.Fatal(err)
	}
	request := Request{
		Target: "concepts/decision.md", Title: "Decision", Description: "A recorded decision.",
		Type: "Decision", Text: []byte("Use stable public identifiers."), Actor: "human:owner",
		Now: time.Date(2026, 7, 26, 0, 0, 0, 0, time.UTC),
	}
	first, err := Run(selected, request)
	if err != nil || !first.Created {
		t.Fatalf("first Run = %+v, %v", first, err)
	}
	second, err := Run(selected, request)
	if err != nil || !second.Unchanged {
		t.Fatalf("second Run = %+v, %v", second, err)
	}
	data, err := os.ReadFile(filepath.Join(selected.Bundle, "concepts", "decision.md"))
	if err != nil || len(data) == 0 {
		t.Fatal(err)
	}
	report, err := health.Scan(selected.Bundle, "wiki", request.Now)
	if err != nil {
		t.Fatal(err)
	}
	if !report.Conformant {
		t.Fatalf("captured concept not conformant: %+v", report.Findings)
	}
}

func TestCaptureRequiresExplicitTargetAndReplacement(t *testing.T) {
	wiki := filepath.Join(t.TempDir(), "wiki")
	if _, err := repository.Initialize(wiki, repository.InitOptions{WikiID: "wiki"}); err != nil {
		t.Fatal(err)
	}
	selected, _ := repository.Open(wiki)
	base := Request{Target: "concepts/note.md", Title: "Note", Text: []byte("one"), Actor: "human:owner"}
	if _, err := Run(selected, base); err != nil {
		t.Fatal(err)
	}
	base.Text = []byte("two")
	if _, err := Run(selected, base); err == nil {
		t.Fatal("implicit replacement accepted")
	}
	base.Replace = true
	if result, err := Run(selected, base); err != nil || !result.Updated {
		t.Fatalf("approved update = %+v, %v", result, err)
	}
}

func TestCaptureRetryRepairsIndexAndLog(t *testing.T) {
	wiki := filepath.Join(t.TempDir(), "wiki")
	if _, err := repository.Initialize(wiki, repository.InitOptions{WikiID: "wiki"}); err != nil {
		t.Fatal(err)
	}
	selected, _ := repository.Open(wiki)
	request := Request{Target: "concepts/recover.md", Title: "Recover", Text: []byte("recoverable"), Actor: "human:owner", Now: time.Date(2026, 7, 26, 0, 0, 0, 0, time.UTC)}
	if _, err := Run(selected, request); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(selected.Bundle, "index.md"), []byte("---\nokf_version: \"0.2\"\n---\n# Knowledge\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(selected.Bundle, "log.md"), []byte("# Log\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	result, err := Run(selected, request)
	if err != nil || !result.Unchanged {
		t.Fatalf("retry = %+v, %v", result, err)
	}
	for _, name := range []string{"index.md", "log.md"} {
		data, readErr := os.ReadFile(filepath.Join(selected.Bundle, name))
		if readErr != nil || !bytes.Contains(data, []byte("concepts/recover.md")) {
			t.Fatalf("%s not repaired: %v\n%s", name, readErr, data)
		}
	}
}
