package prompts

import (
	"io/fs"
	"strings"
	"testing"
)

func TestInstructionPromptKeepsExplicitWikiAndBudaBoundary(t *testing.T) {
	data, err := fs.ReadFile(FS, "guiho-i-buda.md")
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, expected := range []string{"{{WIKI_ID}}", "{{BUNDLE}}", "--wiki", "guiho-s-0002-buda", "Never infer another wiki"} {
		if !strings.Contains(text, expected) {
			t.Errorf("instruction prompt omits %q", expected)
		}
	}
	if strings.Contains(text, "`qmd ") {
		t.Fatal("instruction prompt contains a direct qmd command")
	}
}
