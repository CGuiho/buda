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

func TestCaptureReplaceRewritesWhenOnlyDigestDiffers(t *testing.T) {
	wiki := filepath.Join(t.TempDir(), "wiki")
	if _, err := repository.Initialize(wiki, repository.InitOptions{WikiID: "wiki"}); err != nil {
		t.Fatal(err)
	}
	selected, _ := repository.Open(wiki)
	plainText := []byte("Some content.")
	now := time.Date(2026, 7, 26, 0, 0, 0, 0, time.UTC)
	first, err := Run(selected, Request{
		Target: "concepts/digest.md", Title: "Digest", Description: "Digest drift fix.",
		Type: "Note", Text: plainText, Actor: "human:owner", Now: now,
	})
	if err != nil || !first.Created {
		t.Fatalf("first Run = %+v, %v", first, err)
	}
	// Input normalization always records the digest over the trimmed
	// body-before-marker, so a genuine digest drift can no longer be produced
	// through Run itself. Simulate a stale recorded digest (an artifact
	// captured before normalization) by corrupting the frontmatter digest, then
	// prove --replace rewrites the concept to repair it.
	path := filepath.Join(selected.Bundle, "concepts", "digest.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	staleDigest := append([]byte("sha256:"), bytes.Repeat([]byte("0"), 64)...)
	stale := bytes.Replace(data, []byte(first.Digest), staleDigest, 1)
	if bytes.Equal(stale, data) {
		t.Fatal("recorded digest not found in captured concept frontmatter")
	}
	if err := os.WriteFile(path, stale, 0o644); err != nil {
		t.Fatal(err)
	}
	// Identical content with a differing recorded digest must not be treated as
	// unchanged without explicit replacement approval.
	if _, err := Run(selected, Request{
		Target: "concepts/digest.md", Title: "Digest", Description: "Digest drift fix.",
		Type: "Note", Text: plainText, Actor: "human:owner", Now: now,
	}); err == nil {
		t.Fatal("digest-drift retry without --replace accepted as unchanged")
	}
	second, err := Run(selected, Request{
		Target: "concepts/digest.md", Title: "Digest", Description: "Digest drift fix.",
		Type: "Note", Text: plainText, Actor: "human:owner", Now: now, Replace: true,
	})
	if err != nil || !second.Updated {
		t.Fatalf("second Run = %+v, %v", second, err)
	}
	secondBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(bytes.TrimSpace(plainText))
	expectedDigest := "sha256:" + hex.EncodeToString(sum[:])
	if second.Digest != expectedDigest {
		t.Fatalf("digest = %q, want %q", second.Digest, expectedDigest)
	}
	if !bytes.Contains(secondBytes, []byte(expectedDigest)) {
		t.Fatalf("written frontmatter does not contain recorded digest %q", expectedDigest)
	}
	if bytes.Contains(secondBytes, staleDigest) {
		t.Fatal("stale digest still present after --replace rewrite")
	}
	report, err := health.Scan(selected.Bundle, "wiki", now)
	if err != nil {
		t.Fatal(err)
	}
	if !report.Conformant {
		t.Fatalf("rewritten concept not conformant: %+v", report.Findings)
	}
}

func TestCaptureIdenticalRetryUnchangedWithoutBudaMetadata(t *testing.T) {
	wiki := filepath.Join(t.TempDir(), "wiki")
	if _, err := repository.Initialize(wiki, repository.InitOptions{WikiID: "wiki"}); err != nil {
		t.Fatal(err)
	}
	selected, _ := repository.Open(wiki)
	text := []byte("Legacy content captured before Buda metadata.")
	legacy := []byte("---\ntype: Note\ntitle: Legacy\ndescription: Legacy concept.\nstatus: draft\ngenerated:\n  by: human:owner\n  at: 2026-07-26T00:00:00Z\nsources:\n  - id: capture-input\n    resource: buda:capture\n    title: Explicit user-directed capture input\n    author: human:owner\n---\n\n" + string(text) + "[^capture-input]\n\n[^capture-input]: Explicit user-directed capture input.\n")
	path := filepath.Join(selected.Bundle, "concepts", "legacy.md")
	if err := os.WriteFile(path, legacy, 0o644); err != nil {
		t.Fatal(err)
	}
	request := Request{
		Target: "concepts/legacy.md", Title: "Legacy", Description: "Legacy concept.",
		Type: "Note", Text: text, Actor: "human:owner",
		Now: time.Date(2026, 7, 26, 0, 0, 0, 0, time.UTC),
	}
	result, err := Run(selected, request)
	if err != nil || !result.Unchanged {
		t.Fatalf("identical retry without buda metadata = %+v, %v; want Unchanged with no error", result, err)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(after, legacy) {
		t.Fatal("unchanged retry without buda metadata rewrote the concept file")
	}
	request.Text = []byte("Different content.")
	if _, err := Run(selected, request); err == nil {
		t.Fatal("different content without buda metadata accepted replacement without --replace")
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

func TestCaptureReplacePreservesUnknownOKFMetadata(t *testing.T) {
	wiki := filepath.Join(t.TempDir(), "wiki")
	if _, err := repository.Initialize(wiki, repository.InitOptions{WikiID: "wiki"}); err != nil {
		t.Fatal(err)
	}
	selected, err := repository.Open(wiki)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 8, 16, 0, 0, 0, 0, time.UTC)
	initial := []byte(`---
type: Note
title: Custom Knowledge
description: Has custom producer metadata.
status: draft
custom_owner: human:alice
tags:
  - domain:research
  - priority:high
extra_field:
  nested_key: 42
generated:
  by: human:alice
  at: 2026-08-16T00:00:00Z
sources:
  - id: capture-input
    resource: buda:capture
    title: Explicit user-directed capture input
    author: human:alice
---

Initial text.[^capture-input]

[^capture-input]: Explicit user-directed capture input.
`)
	targetPath := filepath.Join(selected.Bundle, "concepts", "custom.md")
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(targetPath, initial, 0o644); err != nil {
		t.Fatal(err)
	}

	// Now run capture with --replace and updated text
	request := Request{
		Target:      "concepts/custom.md",
		Title:       "Updated Custom Knowledge",
		Description: "Updated description.",
		Type:        "Decision",
		Text:        []byte("Updated text content with new insights."),
		Actor:       "human:bob",
		Now:         now,
		Replace:     true,
	}
	result, err := Run(selected, request)
	if err != nil || !result.Updated {
		t.Fatalf("Run = %+v, %v", result, err)
	}

	updatedBytes, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatal(err)
	}
	document, err := okf.ParseConcept("concepts/custom.md", updatedBytes)
	if err != nil {
		t.Fatalf("parse updated concept: %v", err)
	}

	if document.String("type") != "Decision" {
		t.Fatalf("type = %q, want Decision", document.String("type"))
	}
	if document.String("title") != "Updated Custom Knowledge" {
		t.Fatalf("title = %q, want Updated Custom Knowledge", document.String("title"))
	}
	// Verify unknown OKF metadata keys were preserved
	if document.String("custom_owner") != "human:alice" {
		t.Fatalf("custom_owner = %q, want human:alice", document.String("custom_owner"))
	}
	tags := document.Strings("tags")
	if len(tags) != 2 || tags[0] != "domain:research" || tags[1] != "priority:high" {
		t.Fatalf("tags = %+v, want [domain:research priority:high]", tags)
	}
	metaCopy, err := document.MetadataCopy()
	if err != nil {
		t.Fatalf("MetadataCopy() err = %v", err)
	}
	extra, ok := metaCopy["extra_field"].(map[string]any)
	if !ok || extra["nested_key"] != 42 {
		t.Fatalf("extra_field = %+v, want map with nested_key: 42", metaCopy["extra_field"])
	}
}
