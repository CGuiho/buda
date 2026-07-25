package cmd

import (
	"bytes"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func executeTest(t *testing.T, arguments ...string) (string, string, error) {
	t.Helper()
	var output, diagnostics bytes.Buffer
	deps := Dependencies{In: strings.NewReader(""), Out: &output, Err: &diagnostics, Options: &Options{}}
	command := &cobra.Command{Use: "status", Short: "Report canonical and qmd state.", Args: NoArgs, RunE: func(*cobra.Command, []string) error { return nil }}
	root := NewRootCommand(deps, BuildInfo{Version: "1.2.3"}, command)
	root.SetArgs(arguments)
	err := root.Execute()
	if errors.Is(err, errHelpRendered) {
		err = nil
	}
	return output.String(), diagnostics.String(), err
}

func TestRootWelcomeVersionAndJSON(t *testing.T) {
	output, _, err := executeTest(t)
	if err != nil || output != "Hello Windows - buda v1.2.3\n" {
		t.Fatalf("welcome = %q, err = %v", output, err)
	}
	output, _, err = executeTest(t, "--version")
	if err != nil || output != "buda v1.2.3\n" {
		t.Fatalf("version = %q, err = %v", output, err)
	}
	output, _, err = executeTest(t, "--json")
	if err != nil || !strings.Contains(output, `"message": "Hello Windows - buda v1.2.3"`) || strings.Count(strings.TrimSpace(output), "\n{") != 0 {
		t.Fatalf("json = %q, err = %v", output, err)
	}
}

func TestRepositoryCommandsRequireAndResolveExplicitWiki(t *testing.T) {
	_, _, err := executeTest(t, "status")
	if ExitCode(err) != 3 || !strings.Contains(err.Error(), "never selects a wiki implicitly") {
		t.Fatalf("missing wiki err = %v, code = %d", err, ExitCode(err))
	}
	_, _, err = executeTest(t, "status", "--wiki", ".")
	if err != nil {
		t.Fatalf("explicit wiki: %v", err)
	}
}

func TestDeveloperHelpUsesInvokedSubtreeAndPositiveDepth(t *testing.T) {
	output, _, err := executeTest(t, "agent", "--help-tree-depth", "1")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(output, "COMMAND TREE\n\nbuda agent") || !strings.Contains(output, "├── instruction") || strings.Contains(output, "status") {
		t.Fatalf("scoped tree:\n%s", output)
	}
	if !strings.Contains(output, "--help-docs") || strings.Contains(output, "install  Install") {
		t.Fatalf("depth was not respected:\n%s", output)
	}
	_, _, err = executeTest(t, "agent", "--help-tree-depth", "0")
	if ExitCode(err) != 2 {
		t.Fatalf("depth error = %v, code = %d", err, ExitCode(err))
	}
	output, _, err = executeTest(t, "agent", "skill", "--help-docs", "--help-tree-depth", "1")
	if err != nil || !strings.HasPrefix(output, "## buda agent skill") || strings.Contains(output, "buda status") {
		t.Fatalf("scoped docs = %q, err = %v", output, err)
	}
}

func TestAliasPolicy(t *testing.T) {
	deps := Dependencies{Options: &Options{}}
	root := NewRootCommand(deps, BuildInfo{Version: "test"})
	for _, command := range append([]*cobra.Command{root}, flattenCommands(root.Commands())...) {
		command.LocalNonPersistentFlags().VisitAll(func(flag *pflag.Flag) {
			if flag.Shorthand == "" {
				return
			}
			if flag.Name == "help" && flag.Shorthand == "h" {
				return
			}
			if command == root && flag.Name == "version" && flag.Shorthand == "v" {
				return
			}
			t.Errorf("forbidden shorthand -%s on %s --%s", flag.Shorthand, command.CommandPath(), flag.Name)
		})
	}
}

func flattenCommands(commands []*cobra.Command) []*cobra.Command {
	var result []*cobra.Command
	for _, command := range commands {
		result = append(result, command)
		result = append(result, flattenCommands(command.Commands())...)
	}
	return result
}

func TestSuccessfulRepositoryCommandSchedulesFailureIsolatedMaintenance(t *testing.T) {
	var scheduled int
	deps := Dependencies{
		In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}, Options: &Options{},
		Executable: func() (string, error) { return "buda", nil },
		ScheduleMaintenance: func(_, wiki string) error {
			scheduled++
			if !filepath.IsAbs(wiki) {
				t.Errorf("worker wiki was not absolute: %q", wiki)
			}
			return errors.New("isolated")
		},
	}
	repository := &cobra.Command{Use: "lint", Args: NoArgs, RunE: func(*cobra.Command, []string) error { return nil }}
	root := NewRootCommand(deps, BuildInfo{Version: "test"}, repository)
	root.SetArgs([]string{"lint", "--wiki", "."})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if scheduled != 1 {
		t.Fatalf("scheduled = %d", scheduled)
	}
	root = NewRootCommand(deps, BuildInfo{Version: "test"}, repository)
	root.SetArgs([]string{"agent", "prompt", "list"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if scheduled != 1 {
		t.Fatalf("agent route scheduled maintenance: %d", scheduled)
	}
}

func TestAgentHumanOutputIsDistinctFromJSON(t *testing.T) {
	human, _, err := executeTest(t, "agent", "prompt", "list")
	if err != nil {
		t.Fatal(err)
	}
	if strings.HasPrefix(strings.TrimSpace(human), "[") || !strings.Contains(human, "description:") || !strings.Contains(human, "id: buda") {
		t.Fatalf("human output = %q", human)
	}
	jsonOutput, _, err := executeTest(t, "agent", "prompt", "list", "--json")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(strings.TrimSpace(jsonOutput), "[") || !strings.Contains(jsonOutput, `"id": "buda"`) {
		t.Fatalf("JSON output = %q", jsonOutput)
	}
}
