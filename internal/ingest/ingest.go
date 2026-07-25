// Package ingest registers explicit durable evidence and creates a bounded
// agent work item. It does not synthesize claims or run an LLM.
package ingest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/CGuiho/buda/internal/okf"
	"github.com/CGuiho/buda/internal/repository"
	"github.com/CGuiho/buda/internal/source"
	"go.yaml.in/yaml/v3"
)

const maxCandidates = 25

type Evidence struct {
	Path       string           `json:"path"`
	DocumentID string           `json:"document_id,omitempty"`
	Rank       int              `json:"rank,omitempty"`
	Score      float64          `json:"score,omitempty"`
	Snippet    string           `json:"snippet,omitempty"`
	Metadata   map[string]any   `json:"metadata,omitempty"`
	Sources    []EvidenceSource `json:"sources,omitempty"`
	Warnings   []string         `json:"warnings,omitempty"`
}

type EvidenceSource struct {
	ID       string `json:"id"`
	Resource string `json:"resource"`
	Title    string `json:"title,omitempty"`
	Joined   bool   `json:"claim_footnote_joined"`
}

type Request struct {
	Source     string
	Title      string
	Actor      string
	Now        time.Time
	HTTPClient source.Doer
	MaxBytes   int64
	Candidates []Evidence
}

type WorkItem struct {
	Schema             int        `json:"schema"`
	State              string     `json:"state"`
	WikiID             string     `json:"wiki_id"`
	SourceID           string     `json:"source_id"`
	SourceConcept      string     `json:"source_concept"`
	OriginalResource   string     `json:"original_resource"`
	CanonicalResource  string     `json:"canonical_resource"`
	Digest             string     `json:"digest"`
	MediaType          string     `json:"media_type,omitempty"`
	CreatedAt          string     `json:"created_at"`
	Actor              string     `json:"actor"`
	ExistingCandidates []Evidence `json:"existing_candidates"`
	Instructions       []string   `json:"instructions"`
}

type Result struct {
	SourceID        string `json:"source_id"`
	Digest          string `json:"digest"`
	Artifact        string `json:"artifact"`
	SourceConcept   string `json:"source_concept"`
	WorkItem        string `json:"work_item"`
	ArtifactCreated bool   `json:"artifact_created"`
	Unchanged       bool   `json:"unchanged"`
}

type sourceFrontmatter struct {
	Type        string           `yaml:"type"`
	Title       string           `yaml:"title"`
	Description string           `yaml:"description"`
	Resource    string           `yaml:"resource"`
	Status      string           `yaml:"status"`
	Generated   generated        `yaml:"generated"`
	Sources     []okf.Source     `yaml:"sources"`
	Buda        okf.BudaMetadata `yaml:"buda"`
}

type generated struct {
	By string `yaml:"by"`
	At string `yaml:"at"`
}

func Run(ctx context.Context, selected repository.Repository, request Request) (Result, error) {
	if strings.TrimSpace(request.Source) == "" {
		return Result{}, errors.New("--source is required")
	}
	if strings.TrimSpace(request.Actor) == "" {
		return Result{}, errors.New("ingest actor must not be empty")
	}
	if request.Now.IsZero() {
		request.Now = time.Now().UTC()
	}
	if len(request.Candidates) > maxCandidates {
		return Result{}, fmt.Errorf("ingest work item accepts at most %d existing candidates", maxCandidates)
	}
	if request.HTTPClient == nil {
		request.HTTPClient = &http.Client{Timeout: 30 * time.Second}
	}
	artifact, err := source.Acquire(ctx, request.Source, request.HTTPClient, request.MaxBytes)
	if err != nil {
		return Result{}, err
	}
	stored, err := source.Store(selected, artifact)
	if err != nil {
		return Result{}, err
	}
	if request.Title == "" {
		request.Title = deriveTitle(artifact.Original, stored.SourceID)
	}
	conceptRelative := filepath.ToSlash(filepath.Join("sources", stored.SourceID+".md"))
	conceptAbsolute := filepath.Join(selected.Bundle, filepath.FromSlash(conceptRelative))
	uid := source.StableUID("buda-source:"+selected.Config.WikiID, stored.Digest)
	metadata := sourceFrontmatter{
		Type: "Reference", Title: request.Title,
		Description: "Immutable source evidence registered by Buda ingest.",
		Resource:    stored.Resource, Status: "draft",
		Generated: generated{By: request.Actor, At: request.Now.UTC().Format(time.RFC3339)},
		Sources:   []okf.Source{{ID: "original", Resource: artifact.Original, Title: request.Title, Author: request.Actor}},
		Buda:      okf.BudaMetadata{SchemaVersion: "1", UID: uid, WikiID: selected.Config.WikiID, SourceDigests: map[string]string{"original": stored.Digest}},
	}
	front, err := yaml.Marshal(metadata)
	if err != nil {
		return Result{}, fmt.Errorf("encode source record: %w", err)
	}
	body := fmt.Sprintf("# Source\n\nBuda registered immutable evidence from `%s`.[^original]\n\n[^original]: Original explicit ingest source.\n", artifact.Original)
	conceptData := append([]byte("---\n"), front...)
	conceptData = append(conceptData, []byte("---\n\n"+body)...)
	conceptUnchanged, err := ensureSourceConcept(conceptAbsolute, conceptRelative, conceptData, selected.Config.WikiID, stored.Digest, artifact.Original)
	if err != nil {
		return Result{}, err
	}

	work := WorkItem{
		Schema: 1, State: "pending_agent_synthesis", WikiID: selected.Config.WikiID,
		SourceID: stored.SourceID, SourceConcept: conceptRelative,
		OriginalResource: artifact.Original, CanonicalResource: stored.Resource,
		Digest: stored.Digest, MediaType: artifact.MediaType,
		CreatedAt: request.Now.UTC().Format(time.RFC3339), Actor: request.Actor,
		ExistingCandidates: append([]Evidence(nil), request.Candidates...),
		Instructions: []string{
			"Use the Buda ingest skill; do not treat this work item as canonical knowledge.",
			"Read the immutable source and existing qmd candidates before proposing concept changes.",
			"Attach sources[].id footnotes, preserve contradictions, then run Buda lint and index.",
		},
	}
	workData, err := json.MarshalIndent(work, "", "  ")
	if err != nil {
		return Result{}, fmt.Errorf("encode ingest work item: %w", err)
	}
	workData = append(workData, '\n')
	digestHex := strings.TrimPrefix(stored.Digest, "sha256:")
	workAbsolute := filepath.Join(selected.Derived, "ingest", digestHex+".json")
	workUnchanged, err := ensureWorkItem(workAbsolute, workData, selected.Config.WikiID, stored.Digest, artifact.Original)
	if err != nil {
		return Result{}, err
	}
	if err := appendIndex(selected.Bundle, conceptRelative, request.Title); err != nil {
		return Result{}, err
	}
	if err := appendLog(selected.Bundle, conceptRelative, request.Title, request.Now); err != nil {
		return Result{}, err
	}
	return Result{
		SourceID: stored.SourceID, Digest: stored.Digest,
		Artifact:        filepath.ToSlash(filepath.Join("references", "raw", digestHex+".source")),
		SourceConcept:   conceptRelative,
		WorkItem:        filepath.ToSlash(filepath.Join(selected.Config.Derived, "ingest", digestHex+".json")),
		ArtifactCreated: stored.Created,
		Unchanged:       !stored.Created && conceptUnchanged && workUnchanged,
	}, nil
}

func deriveTitle(original, fallback string) string {
	if parsed, err := url.Parse(original); err == nil && parsed.IsAbs() {
		if base := filepath.Base(parsed.Path); base != "." && base != "/" && base != "" {
			return base
		}
		return parsed.Host
	}
	if base := filepath.Base(original); base != "." && base != "" {
		return base
	}
	return fallback
}

func ensureSourceConcept(path, relative string, data []byte, wikiID, digest, original string) (bool, error) {
	existing, err := os.ReadFile(path)
	if err == nil {
		document, parseErr := okf.ParseConcept(relative, existing)
		if parseErr != nil {
			return false, fmt.Errorf("existing source record is invalid: %w", parseErr)
		}
		metadata, present, parseErr := document.Buda()
		sources, sourcesErr := document.Sources()
		if parseErr != nil || !present || metadata.WikiID != wikiID || metadata.SourceDigests["original"] != digest || sourcesErr != nil || len(sources) != 1 || sources[0].Resource != original {
			return false, fmt.Errorf("existing source record %q does not match selected wiki and digest", relative)
		}
		if document.String("type") == "Reference" {
			return true, nil
		}
		return false, fmt.Errorf("existing source record %q is not a Reference", relative)
	}
	if !errors.Is(err, os.ErrNotExist) {
		return false, err
	}
	if err := repository.AtomicWriteFile(path, data, 0o644); err != nil {
		return false, err
	}
	return false, nil
}

func ensureWorkItem(path string, data []byte, wikiID, digest, original string) (bool, error) {
	existing, err := os.ReadFile(path)
	if err == nil {
		var work WorkItem
		decoder := json.NewDecoder(bytes.NewReader(existing))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&work); err != nil {
			return false, fmt.Errorf("decode existing ingest work item: %w", err)
		}
		if work.Schema != 1 || work.WikiID != wikiID || work.Digest != digest || work.OriginalResource != original {
			return false, fmt.Errorf("existing ingest work item %q does not match selected wiki and digest", path)
		}
		return true, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return false, err
	}
	if err := repository.AtomicWriteFile(path, data, 0o600); err != nil {
		return false, err
	}
	return false, nil
}

func appendIndex(bundle, concept, title string) error {
	path := filepath.Join(bundle, "index.md")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if bytes.Contains(data, []byte("]("+concept+")")) {
		return nil
	}
	if len(data) > 0 && data[len(data)-1] != '\n' {
		data = append(data, '\n')
	}
	if !bytes.Contains(data, []byte("# Source Records\n")) {
		data = append(data, []byte("\n# Source Records\n")...)
	}
	data = append(data, []byte("\n* ["+title+"]("+concept+") - Immutable source record.\n")...)
	return repository.AtomicWriteFile(path, data, 0o644)
}

func appendLog(bundle, concept, title string, now time.Time) error {
	path := filepath.Join(bundle, "log.md")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	entry := fmt.Sprintf("* **Creation**: Registered source [%s](%s).", title, concept)
	if bytes.Contains(data, []byte(entry)) {
		return nil
	}
	heading := "## " + now.UTC().Format("2006-01-02")
	text := string(data)
	if index := strings.Index(text, heading); index >= 0 {
		insert := index + len(heading)
		text = text[:insert] + "\n" + entry + text[insert:]
	} else {
		text = strings.TrimRight(text, "\r\n") + "\n\n" + heading + "\n" + entry + "\n"
	}
	return repository.AtomicWriteFile(path, []byte(text), 0o644)
}
