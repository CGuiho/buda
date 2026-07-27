package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUnhealthyLintJSONIsRenderedExactlyOnce(t *testing.T) {
	wiki := initializedWiki(t)
	invalid := filepath.Join(wiki, "knowledge", "concepts", "invalid.md")
	if err := os.WriteFile(invalid, []byte("# Missing OKF frontmatter\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var output, diagnostics bytes.Buffer
	deps := Dependencies{
		In: strings.NewReader(""), Out: &output, Err: &diagnostics, Options: &Options{},
		Executable:          func() (string, error) { return "buda", nil },
		ScheduleMaintenance: func(string, string) error { return nil },
	}
	root := NewRootCommand(deps, BuildInfo{Version: "test"}, NewApplicationCommands(deps)...)
	root.SetArgs([]string{"lint", "--wiki", wiki, "--json"})
	err := root.Execute()
	if ExitCode(err) != 1 || !IsErrorRendered(err) {
		t.Fatalf("lint error = %v, code = %d, rendered = %t", err, ExitCode(err), IsErrorRendered(err))
	}
	if diagnostics.Len() != 0 || !strings.HasPrefix(strings.TrimSpace(output.String()), "{") || strings.Count(strings.TrimSpace(output.String()), "\n{") != 0 {
		t.Fatalf("stdout=%q stderr=%q", output.String(), diagnostics.String())
	}
}
