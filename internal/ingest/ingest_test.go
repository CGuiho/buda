package ingest

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/CGuiho/buda/internal/health"
	"github.com/CGuiho/buda/internal/repository"
)

func TestIngestCreatesImmutableSourceRecordAndBoundedWorkItem(t *testing.T) {
	wiki := filepath.Join(t.TempDir(), "wiki")
	if _, err := repository.Initialize(wiki, repository.InitOptions{WikiID: "wiki"}); err != nil {
		t.Fatal(err)
	}
	selected, _ := repository.Open(wiki)
	input := filepath.Join(t.TempDir(), "paper.txt")
	if err := os.WriteFile(input, []byte("evidence"), 0o644); err != nil {
		t.Fatal(err)
	}
	request := Request{Source: input, Title: "Paper", Actor: "human:owner", Now: time.Date(2026, 7, 26, 0, 0, 0, 0, time.UTC)}
	result, err := Run(context.Background(), selected, request)
	if err != nil {
		t.Fatal(err)
	}
	if !result.ArtifactCreated || result.SourceConcept == "" || result.WorkItem == "" {
		t.Fatalf("result = %+v", result)
	}
	second, err := Run(context.Background(), selected, request)
	if err != nil || !second.Unchanged {
		t.Fatalf("retry = %+v, %v", second, err)
	}
	report, err := health.Scan(selected.Bundle, "wiki", request.Now)
	if err != nil || !report.Conformant {
		t.Fatalf("health = %+v, %v", report, err)
	}
}

func TestIngestRejectsUnboundedCandidateSet(t *testing.T) {
	wiki := filepath.Join(t.TempDir(), "wiki")
	if _, err := repository.Initialize(wiki, repository.InitOptions{WikiID: "wiki"}); err != nil {
		t.Fatal(err)
	}
	selected, _ := repository.Open(wiki)
	candidates := make([]Evidence, maxCandidates+1)
	if _, err := Run(context.Background(), selected, Request{Source: "unused", Actor: "human:owner", Candidates: candidates}); err == nil {
		t.Fatal("unbounded work item accepted")
	}
}
