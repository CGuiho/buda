package cmd

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/CGuiho/buda/internal/ingest"
	"github.com/CGuiho/buda/internal/qmd"
	"github.com/CGuiho/buda/internal/repository"
)

type ingestQMD struct {
	domainFakeQMD
	queries []qmd.SearchOptions
	err     error
}

func (fake *ingestQMD) Search(_ context.Context, options qmd.SearchOptions) ([]qmd.Match, error) {
	fake.queries = append(fake.queries, options)
	return append([]qmd.Match(nil), fake.search...), fake.err
}

type ingestOutput struct {
	Mode       qmd.SearchMode    `json:"mode"`
	Ingest     ingest.Result     `json:"ingest"`
	Candidates []ingest.Evidence `json:"existing_candidates"`
}

func runIngestTest(t *testing.T, wiki string, client *ingestQMD, args ...string) (ingestOutput, error) {
	t.Helper()
	var output bytes.Buffer
	deps := domainDeps(&output, t.TempDir())
	deps.Now = func() time.Time { return time.Date(2026, 10, 3, 23, 0, 0, 0, time.UTC) }
	root := NewRootCommand(deps, BuildInfo{Version: "test"}, NewIngestCommand(deps, func(repository.Repository) (QMDClient, error) { return client, nil }))
	root.SetArgs(append([]string{"ingest", "--wiki", wiki, "--json"}, args...))
	err := root.Execute()
	var result ingestOutput
	if err == nil {
		if err := json.Unmarshal(output.Bytes(), &result); err != nil {
			t.Fatalf("decode ingest output: %v\n%s", err, output.String())
		}
	}
	return result, err
}

func TestIngestSelectsExistingQMDModesAndKeepsHybridDefault(t *testing.T) {
	for _, test := range []struct {
		name, flag, title string
		mode              qmd.SearchMode
	}{
		{name: "default", mode: qmd.ModeHybrid, title: "Repository overview"},
		{name: "hybrid", flag: "--mode=hybrid", mode: qmd.ModeHybrid, title: "Repository overview"},
		{name: "lexical", flag: "--mode=lexical", mode: qmd.ModeLexical, title: "Repository overview"},
		{name: "semantic", flag: "--mode=semantic", mode: qmd.ModeSemantic, title: "Repository overview"},
		{name: "source-fallback", flag: "--mode=lexical", mode: qmd.ModeLexical},
	} {
		t.Run(test.name, func(t *testing.T) {
			wiki := initializedWiki(t)
			source := filepath.Join(t.TempDir(), "overview.txt")
			writeTestFile(t, source, "Exact source evidence.\n")
			args := []string{"--source", source, "--actor", "human:owner"}
			if test.flag != "" {
				args = append(args, test.flag)
			}
			if test.title != "" {
				args = append(args, "--title", test.title)
			}
			client := &ingestQMD{}
			result, err := runIngestTest(t, wiki, client, args...)
			if err != nil {
				t.Fatal(err)
			}
			query := test.title
			if query == "" {
				query = source
			}
			want := []qmd.SearchOptions{{Mode: test.mode, Text: query, Limit: 10}}
			if !reflect.DeepEqual(client.queries, want) || result.Mode != test.mode || client.ensure != 1 || client.update != 1 {
				t.Fatalf("queries=%+v mode=%s ensure=%d update=%d", client.queries, result.Mode, client.ensure, client.update)
			}
			if !result.Ingest.ArtifactCreated || result.Ingest.Unchanged {
				t.Fatalf("initial ingest = %+v", result.Ingest)
			}
		})
	}
}

func TestIngestLexicalPreservesNormalizedCandidatesAndSealedRepeat(t *testing.T) {
	wiki := initializedWiki(t)
	source := filepath.Join(t.TempDir(), "overview.txt")
	firstBytes := []byte("Exact first source evidence.\n")
	writeTestFile(t, source, string(firstBytes))
	client := &ingestQMD{}
	args := []string{"--source", source, "--actor", "human:owner", "--title", "Repository overview", "--mode", "lexical"}
	first, err := runIngestTest(t, wiki, client, args...)
	if err != nil {
		t.Fatal(err)
	}
	client.search = []qmd.Match{{Path: first.Ingest.SourceConcept, Rank: 1, DocumentID: "#abc123", Score: 0.73, Snippet: "Registered immutable evidence"}}
	before := ingestTree(t, wiki)
	second, err := runIngestTest(t, wiki, client, args...)
	if err != nil || !second.Ingest.Unchanged || second.Ingest.ArtifactCreated || second.Ingest.Digest != first.Ingest.Digest || second.Ingest.SourceID != first.Ingest.SourceID {
		t.Fatalf("repeat = %+v, %v", second, err)
	}
	if after := ingestTree(t, wiki); !reflect.DeepEqual(before, after) {
		t.Fatal("unchanged repeat changed wiki bytes or file set")
	}
	if len(second.Candidates) != 1 || second.Candidates[0].DocumentID != "#abc123" || second.Candidates[0].Score != 0.73 || len(second.Candidates[0].Sources) != 1 || !second.Candidates[0].Sources[0].Joined || second.Candidates[0].Sources[0].Resource != source {
		t.Fatalf("normalized candidates = %+v", second.Candidates)
	}
	if second.Candidates[0].Metadata["title"] != "Repository overview" {
		t.Fatalf("candidate metadata = %+v", second.Candidates[0].Metadata)
	}
	changedBytes := []byte("Changed source evidence with a new seal.\n")
	writeTestFile(t, source, string(changedBytes))
	changed, err := runIngestTest(t, wiki, client, args...)
	if err != nil || changed.Ingest.Unchanged || !changed.Ingest.ArtifactCreated || changed.Ingest.Digest == first.Ingest.Digest {
		t.Fatalf("changed ingest = %+v, %v", changed, err)
	}
	for _, sealed := range []struct {
		result ingest.Result
		bytes  []byte
	}{{first.Ingest, firstBytes}, {changed.Ingest, changedBytes}} {
		data, err := os.ReadFile(filepath.Join(wiki, "knowledge", sealed.result.Artifact))
		if err != nil || !bytes.Equal(data, sealed.bytes) {
			t.Fatalf("raw evidence differs: %s, %v", sealed.result.Artifact, err)
		}
	}
	var work ingest.WorkItem
	data, err := os.ReadFile(filepath.Join(wiki, changed.Ingest.WorkItem))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &work); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(work.ExistingCandidates, changed.Candidates) || work.OriginalResource != source || work.Digest != changed.Ingest.Digest {
		t.Fatalf("work item did not preserve provenance/candidates: %+v", work)
	}
}

func TestIngestRejectsInvalidModeBeforeRepositoryAndQMDSideEffects(t *testing.T) {
	for _, mode := range []string{"", "Lexical", "lex", " lexical", "lexical ", "unknown"} {
		t.Run(mode, func(t *testing.T) {
			wiki := initializedWiki(t)
			before := ingestTree(t, wiki)
			client := &ingestQMD{}
			_, err := runIngestTest(t, wiki, client, "--source", "unused", "--actor", "human:owner", "--mode", mode)
			if ExitCode(err) != 2 || !strings.Contains(err.Error(), "--mode must be") {
				t.Fatalf("error = %v, code=%d", err, ExitCode(err))
			}
			if client.ensure != 0 || client.update != 0 || len(client.queries) != 0 || !reflect.DeepEqual(before, ingestTree(t, wiki)) {
				t.Fatal("invalid mode caused qmd calls or wiki mutation")
			}
		})
	}
	var output bytes.Buffer
	deps := domainDeps(&output, t.TempDir())
	deps.Options.Wiki = filepath.Join(t.TempDir(), "absent-wiki")
	command := NewIngestCommand(deps, func(repository.Repository) (QMDClient, error) {
		t.Fatal("invalid mode created qmd client")
		return nil, nil
	})
	command.SetArgs([]string{"--source", "unused", "--actor", "human:owner", "--mode", "unknown"})
	if err := command.Execute(); ExitCode(err) != 2 {
		t.Fatalf("invalid mode must precede absent-repository error: %v", err)
	}
}

func TestIngestSearchFailureKeepsCanonicalEvidenceUntouched(t *testing.T) {
	wiki := initializedWiki(t)
	before := ingestTree(t, wiki)
	client := &ingestQMD{err: errors.New("qmd unavailable")}
	_, err := runIngestTest(t, wiki, client, "--source", "unused", "--actor", "human:owner", "--mode", "lexical")
	if err == nil || !strings.Contains(err.Error(), "retrieve existing qmd evidence") || client.update != 0 || !reflect.DeepEqual(before, ingestTree(t, wiki)) {
		t.Fatalf("search failure mutated canonical evidence: %v", err)
	}
}

func ingestTree(t *testing.T, root string) map[string][32]byte {
	t.Helper()
	files := make(map[string][32]byte)
	if err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		files[relative] = sha256.Sum256(data)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	return files
}
