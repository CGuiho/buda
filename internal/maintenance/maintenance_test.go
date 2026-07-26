package maintenance

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/CGuiho/buda/internal/agent"
	"github.com/CGuiho/buda/internal/config"
	"github.com/spf13/cobra"
)

func TestShouldScheduleOnlyNormalRepositoryCommands(t *testing.T) {
	root := &cobra.Command{Use: "buda"}
	repository := &cobra.Command{Use: "lint"}
	agent := &cobra.Command{Use: "agent"}
	agentChild := &cobra.Command{Use: "list"}
	hidden := &cobra.Command{Use: WorkerCommand, Hidden: true}
	uninstall := &cobra.Command{Use: "uninstall"}
	wikiGroup := &cobra.Command{Use: "wiki"}
	nestedUninstall := &cobra.Command{Use: "uninstall"}
	root.AddCommand(repository, agent, hidden, uninstall)
	root.AddCommand(wikiGroup)
	wikiGroup.AddCommand(nestedUninstall)
	agent.AddCommand(agentChild)
	for name, test := range map[string]struct {
		command *cobra.Command
		want    bool
	}{
		"root": {root, true}, "repository": {repository, true}, "agent": {agent, false},
		"agent child": {agentChild, false}, "hidden": {hidden, false}, "uninstall": {uninstall, false},
		"nested uninstall": {nestedUninstall, false},
	} {
		t.Run(name, func(t *testing.T) {
			if got := ShouldSchedule(test.command); got != test.want {
				t.Fatalf("ShouldSchedule() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestRunWithoutWikiReconcilesOnlyBothGlobalSkillDestinations(t *testing.T) {
	home := t.TempDir()
	service := agent.NewService(agent.DefaultResources(), agent.WithHomeDir(func() (string, error) { return home, nil }))
	if err := Run(service, "", WithStateDirectory(t.TempDir())); err != nil {
		t.Fatal(err)
	}
	for _, tool := range []string{"agents", "claude"} {
		path := filepath.Join(home, "."+tool, "skills", agent.SkillID, "SKILL.md")
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("global skill missing at %s: %v", path, err)
		}
	}
	for _, name := range []string{"AGENTS.md", "CLAUDE.md"} {
		if _, err := os.Stat(filepath.Join(home, name)); !os.IsNotExist(err) {
			t.Fatalf("global-only maintenance unexpectedly wrote %s", name)
		}
	}
}

func TestRunWithExplicitWikiAlsoReconcilesInstructionsAndRejectsMalformedMarkers(t *testing.T) {
	home := t.TempDir()
	wiki := t.TempDir()
	configuration, err := config.Marshal(config.Default("maintenance-test"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wiki, config.FileName), configuration, 0o644); err != nil {
		t.Fatal(err)
	}
	service := agent.NewService(agent.DefaultResources(), agent.WithHomeDir(func() (string, error) { return home, nil }))
	if err := Run(service, wiki, WithStateDirectory(t.TempDir())); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(filepath.Join(wiki, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{agent.InstructionBegin, "Buda is the required tool", "--wiki", "evaluate the returned evidence", "do not bypass Buda"} {
		if !strings.Contains(string(content), expected) {
			t.Errorf("managed instruction missing %q: %s", expected, content)
		}
	}

	malformed := "owner policy\n" + agent.InstructionBegin + "\nunclosed\n"
	if err := os.WriteFile(filepath.Join(wiki, "AGENTS.md"), []byte(malformed), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Run(service, wiki, WithStateDirectory(t.TempDir())); err == nil || !strings.Contains(err.Error(), "malformed") {
		t.Fatalf("malformed marker error = %v", err)
	}
	unchanged, err := os.ReadFile(filepath.Join(wiki, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(unchanged) != malformed {
		t.Fatalf("malformed target changed: %q", unchanged)
	}
}

func TestMaintenanceLockSerializesBareAndExplicitWikiWorkers(t *testing.T) {
	state := t.TempDir()
	first, acquired, err := acquireLock("", state)
	if err != nil || !acquired {
		t.Fatalf("acquire bare lock = %v, %v", acquired, err)
	}
	defer first.Close()
	defer os.Remove(first.Name())
	second, acquired, err := acquireLock(`C:\wiki-a`, state)
	if err != nil || !acquired || second == nil {
		t.Fatalf("independent wiki lock while global worker active = %#v, %v, %v", second, acquired, err)
	}
	releaseLock(second)
}

func TestConcurrentDistinctWikiRunsReconcileBothWikisAndGlobalSkill(t *testing.T) {
	home := t.TempDir()
	state := t.TempDir()
	service := agent.NewService(agent.DefaultResources(), agent.WithHomeDir(func() (string, error) { return home, nil }))
	wikis := []string{t.TempDir(), t.TempDir()}
	for index, wiki := range wikis {
		configuration, err := config.Marshal(config.Default(fmt.Sprintf("concurrent-%d", index)))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(wiki, config.FileName), configuration, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	var wait sync.WaitGroup
	errorsByWiki := make([]error, len(wikis))
	for index, wiki := range wikis {
		wait.Add(1)
		go func(index int, wiki string) {
			defer wait.Done()
			errorsByWiki[index] = Run(service, wiki, WithStateDirectory(state))
		}(index, wiki)
	}
	wait.Wait()
	for index, err := range errorsByWiki {
		if err != nil {
			t.Fatalf("run[%d]: %v", index, err)
		}
		if _, err := os.Stat(filepath.Join(wikis[index], "AGENTS.md")); err != nil {
			t.Fatalf("wiki[%d] instructions: %v", index, err)
		}
	}
	for _, tool := range []string{"agents", "claude"} {
		if _, err := os.Stat(filepath.Join(home, "."+tool, "skills", agent.SkillID, "SKILL.md")); err != nil {
			t.Fatalf("global %s skill: %v", tool, err)
		}
	}
	entries, err := os.ReadDir(state)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("maintenance lock debris: %v", entries)
	}
}

func TestRunRejectsRelativeWikiBeforeMutation(t *testing.T) {
	home := t.TempDir()
	service := agent.NewService(agent.DefaultResources(), agent.WithHomeDir(func() (string, error) { return home, nil }))
	if err := Run(service, ".", WithStateDirectory(t.TempDir())); err == nil || !strings.Contains(err.Error(), "absolute canonical") {
		t.Fatalf("relative wiki error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".agents")); !os.IsNotExist(err) {
		t.Fatalf("relative worker mutated global skill: %v", err)
	}
}
