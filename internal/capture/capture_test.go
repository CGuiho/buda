package capture

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/CGuiho/buda/internal/health"
	"github.com/CGuiho/buda/internal/okf"
	"github.com/CGuiho/buda/internal/repository"
)

var footnoteDefLine = regexp.MustCompile(`(?m)^\[\^([^]]+)\]:.*$`)

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

// TestCaptureTrailingNewlineDigestMatches proves the recorded source digest is
// computed over the trimmed body, not the raw input: `cat file | buda capture`
// yields text ending in a trailing newline that must not cause a digest
// mismatch when the health check re-hashes the trimmed body-before-marker.
func TestCaptureTrailingNewlineDigestMatches(t *testing.T) {
	wiki := filepath.Join(t.TempDir(), "wiki")
	if _, err := repository.Initialize(wiki, repository.InitOptions{WikiID: "wiki"}); err != nil {
		t.Fatal(err)
	}
	selected, err := repository.Open(wiki)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 8, 5, 0, 0, 0, 0, time.UTC)
	request := Request{
		Target: "concepts/notes.md", Title: "Notes", Description: "Trailing newline capture.",
		Type: "Note", Text: []byte("Stable public identifiers.\n"), Actor: "human:owner", Now: now,
	}
	result, err := Run(selected, request)
	if err != nil || !result.Created {
		t.Fatalf("Run = %+v, %v", result, err)
	}
	data, err := os.ReadFile(filepath.Join(selected.Bundle, "concepts", "notes.md"))
	if err != nil {
		t.Fatal(err)
	}
	document, err := okf.ParseConcept("concepts/notes.md", data)
	if err != nil {
		t.Fatalf("parse captured concept: %v", err)
	}
	metadata, present, err := document.Buda()
	if err != nil || !present {
		t.Fatalf("buda metadata: present=%v err=%v", present, err)
	}
	expectedSum := sha256.Sum256([]byte("Stable public identifiers."))
	expected := "sha256:" + hex.EncodeToString(expectedSum[:])
	recorded := metadata.SourceDigests[result.SourceID]
	if recorded != expected {
		t.Fatalf("source_digest mismatch: recorded %q want %q", recorded, expected)
	}
	// The health check hashes bytes.TrimSpace(body-before-marker), which is
	// exactly the trimmed input. The captured concept must be conformant and
	// healthy so no source_digest_mismatch finding is raised.
	report, err := health.Scan(selected.Bundle, "wiki", now)
	if err != nil {
		t.Fatal(err)
	}
	if !report.Conformant || !report.Healthy {
		t.Fatalf("captured concept not healthy: conformant=%v healthy=%v findings=%+v", report.Conformant, report.Healthy, report.Findings)
	}
}

// TestCaptureAlreadyMarkedDoesNotDuplicateMarker proves that when the input
// text already ends with the capture footnote marker, the marker is not
// appended a second time. The body must contain the marker exactly once and
// the digest must match the trimmed body-before-marker.
func TestCaptureAlreadyMarkedDoesNotDuplicateMarker(t *testing.T) {
	wiki := filepath.Join(t.TempDir(), "wiki")
	if _, err := repository.Initialize(wiki, repository.InitOptions{WikiID: "wiki"}); err != nil {
		t.Fatal(err)
	}
	selected, err := repository.Open(wiki)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 8, 5, 0, 0, 0, 0, time.UTC)
	marker := "[^capture-input]"
	request := Request{
		Target: "concepts/marked.md", Title: "Marked", Description: "Pre-marked capture.",
		Type: "Note", Text: []byte("Explicit user-directed capture input." + marker), Actor: "human:owner", Now: now,
	}
	result, err := Run(selected, request)
	if err != nil || !result.Created {
		t.Fatalf("Run = %+v, %v", result, err)
	}
	data, err := os.ReadFile(filepath.Join(selected.Bundle, "concepts", "marked.md"))
	if err != nil {
		t.Fatal(err)
	}
	document, err := okf.ParseConcept("concepts/marked.md", data)
	if err != nil {
		t.Fatalf("parse captured concept: %v", err)
	}
	if bytes.Contains(document.Body, []byte(marker+marker)) {
		t.Fatalf("doubled marker %q present in body\n%s", marker+marker, document.Body)
	}
	// Count citation markers excluding footnote definition lines, mirroring the
	// health check's footnoteDefLine.ReplaceAll approach. The citation marker
	// must appear exactly once.
	claimBody := footnoteDefLine.ReplaceAll(document.Body, nil)
	if count := bytes.Count(claimBody, []byte(marker)); count != 1 {
		t.Fatalf("citation marker %q appears %d times, want exactly 1\n%s", marker, count, document.Body)
	}
	// The recorded digest must equal the hash of the trimmed body-before-marker,
	// which is the plain text without the marker.
	beforeMarker := bytes.TrimSpace(bytes.Split(document.Body, []byte(marker))[0])
	document2, _ := okf.ParseConcept("concepts/marked.md", data)
	metadata, present, err := document2.Buda()
	if err != nil || !present {
		t.Fatalf("buda metadata: present=%v err=%v", present, err)
	}
	expected := fmt.Sprintf("sha256:%x", sha256.Sum256(beforeMarker))
	if recorded := metadata.SourceDigests[result.SourceID]; recorded != expected {
		t.Fatalf("source_digest mismatch: recorded %q want %q", recorded, expected)
	}
	report, err := health.Scan(selected.Bundle, "wiki", now)
	if err != nil {
		t.Fatal(err)
	}
	if !report.Conformant || !report.Healthy {
		t.Fatalf("captured concept not healthy: conformant=%v healthy=%v findings=%+v", report.Conformant, report.Healthy, report.Findings)
	}
}
