package skills

import (
	"io/fs"
	"strings"
	"testing"
)

func TestBudaSkillRoutesSixIntentsThroughBuda(t *testing.T) {
	paths := []string{
		"guiho-s-0002-buda/SKILL.md",
		"guiho-s-0002-buda/references/capture.md",
		"guiho-s-0002-buda/references/ingest.md",
		"guiho-s-0002-buda/references/retrieval.md",
		"guiho-s-0002-buda/references/curation.md",
		"guiho-s-0002-buda/references/maintenance.md",
	}
	var corpus strings.Builder
	for _, path := range paths {
		data, err := fs.ReadFile(FS, path)
		if err != nil {
			t.Fatal(err)
		}
		corpus.Write(data)
		corpus.WriteByte('\n')
	}
	text := corpus.String()
	for _, intent := range []string{"save", "ingest", "find", "cite", "curate", "maintain"} {
		if !strings.Contains(strings.ToLower(text), intent) {
			t.Errorf("embedded skill omits %q intent", intent)
		}
	}
	for _, command := range []string{"buda capture", "buda ingest", "buda query", "buda get", "buda lint", "buda index", "buda status", "buda doctor"} {
		if !strings.Contains(text, command) {
			t.Errorf("embedded skill omits %q routing", command)
		}
	}
	if strings.Contains(text, "`qmd ") {
		t.Fatal("embedded skill contains a direct qmd command")
	}
}

func TestBudaSkillDocumentsRequiredWriteFlags(t *testing.T) {
	capture, _ := fs.ReadFile(FS, "guiho-s-0002-buda/references/capture.md")
	for _, flag := range []string{"--wiki", "--target", "--title", "--actor"} {
		if !strings.Contains(string(capture), flag) {
			t.Errorf("capture guidance omits %s", flag)
		}
	}
	ingest, _ := fs.ReadFile(FS, "guiho-s-0002-buda/references/ingest.md")
	for _, flag := range []string{"--wiki", "--source", "--actor"} {
		if !strings.Contains(string(ingest), flag) {
			t.Errorf("ingest guidance omits %s", flag)
		}
	}
}
