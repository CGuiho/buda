package agent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/CGuiho/buda/internal/config"
)

func testService(t *testing.T, home string) *Service {
	t.Helper()
	return NewService(DefaultResources(), WithHomeDir(func() (string, error) { return home, nil }))
}

func writeWikiConfig(t *testing.T, wiki, wikiID string) {
	t.Helper()
	data, err := config.Marshal(config.Default(wikiID))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wiki, config.FileName), data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestEmbeddedResourcesCarryTheRequiredIdentityAndReferences(t *testing.T) {
	service := testService(t, t.TempDir())
	skill, err := service.Skill()
	if err != nil {
		t.Fatal(err)
	}
	if skill.ID != SkillID || skill.Version == "" || !strings.HasPrefix(skill.Digest, "sha256:") {
		t.Fatalf("skill = %#v", skill)
	}
	prompt, err := service.Prompt()
	if err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{"{{WIKI_ID}}", "{{BUNDLE}}", "not access", "--wiki"} {
		if !strings.Contains(prompt.Body, required) {
			t.Errorf("prompt missing %q", required)
		}
	}
	content, err := os.ReadFile(filepath.Join("..", "..", "skills", SkillID, "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, reference := range []string{"knowledge-catalog/blob/main/okf/SPEC.md", "open-knowledge-format-can-improve-data-sharing", "442a6bf555914893e9891c11519de94f"} {
		if !strings.Contains(string(content), reference) {
			t.Errorf("skill missing governing reference %q", reference)
		}
	}
}

func TestSkillInstallationIsDualAtomicAndIdempotent(t *testing.T) {
	home := t.TempDir()
	service := testService(t, home)
	results, err := service.InstallSkill(false, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("results = %#v", results)
	}
	for _, tool := range []string{"agents", "claude"} {
		root := filepath.Join(home, "."+tool, "skills", SkillID)
		for _, path := range []string{"SKILL.md", filepath.Join("references", "capture.md"), filepath.Join("references", "retrieval.md")} {
			if _, err := os.Stat(filepath.Join(root, path)); err != nil {
				t.Fatalf("missing %s: %v", filepath.Join(root, path), err)
			}
		}
	}
	results, err = service.InstallSkill(false, "")
	if err != nil {
		t.Fatal(err)
	}
	for _, result := range results {
		if result.Changed || !result.Current {
			t.Errorf("idempotent result = %#v", result)
		}
	}
	removed, err := service.UninstallSkill(false, "")
	if err != nil {
		t.Fatal(err)
	}
	for _, result := range removed {
		if !result.Changed {
			t.Errorf("removal result = %#v", result)
		}
		if _, err := os.Stat(result.Path); !os.IsNotExist(err) {
			t.Errorf("skill still exists at %s", result.Path)
		}
	}
}

func TestCurrentSkillDirectoriesAreLeftUntouched(t *testing.T) {
	home := t.TempDir()
	service := testService(t, home)
	if _, err := service.InstallSkill(false, ""); err != nil {
		t.Fatal(err)
	}
	stamp := time.Date(2001, 2, 3, 4, 5, 6, 0, time.UTC)
	paths := []string{
		filepath.Join(home, ".agents", "skills", SkillID, "SKILL.md"),
		filepath.Join(home, ".claude", "skills", SkillID, "SKILL.md"),
	}
	for _, path := range paths {
		if err := os.Chtimes(path, stamp, stamp); err != nil {
			t.Fatal(err)
		}
	}
	results, err := service.InstallSkill(false, "")
	if err != nil {
		t.Fatal(err)
	}
	for index, path := range paths {
		if results[index].Changed {
			t.Fatalf("current destination reported changed: %#v", results[index])
		}
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if !info.ModTime().Equal(stamp) {
			t.Errorf("current destination was replaced: modtime = %s", info.ModTime())
		}
	}
}

func TestLocalSkillRequiresExplicitWikiAndUsesBothLocalTargets(t *testing.T) {
	service := testService(t, t.TempDir())
	if _, err := service.InstallSkill(true, ""); err == nil || err.(*Error).ExitCode() != 2 {
		t.Fatalf("missing local wiki error = %v", err)
	}
	wiki := t.TempDir()
	if _, err := service.InstallSkill(true, wiki); err == nil || err.(*Error).ExitCode() != 3 {
		t.Fatalf("non-wiki local target error = %v", err)
	}
	writeWikiConfig(t, wiki, "wiki-local")
	results, err := service.InstallSkill(true, wiki)
	if err != nil {
		t.Fatal(err)
	}
	for _, result := range results {
		if result.Scope != "local" || !strings.HasPrefix(result.Path, wiki) {
			t.Errorf("local result = %#v", result)
		}
	}
}

func TestInstructionsRespectTargetsLineEndingsOutsideContentAndIdempotency(t *testing.T) {
	wiki := t.TempDir()
	writeWikiConfig(t, wiki, "wiki-test")
	agentsPath := filepath.Join(wiki, "AGENTS.md")
	claudePath := filepath.Join(wiki, "CLAUDE.md")
	if err := os.WriteFile(agentsPath, []byte("owner policy\r\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(claudePath, []byte("claude policy\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	service := testService(t, t.TempDir())
	results, err := service.ApplyInstructions(wiki)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("targets = %#v", results)
	}
	for _, path := range []string{agentsPath, claudePath} {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		text := string(content)
		if strings.Count(text, InstructionBegin) != 1 || !strings.Contains(text, "wiki-test") || !strings.Contains(text, "knowledge") {
			t.Errorf("managed content in %s:\n%s", path, text)
		}
	}
	agentsContent, _ := os.ReadFile(agentsPath)
	if strings.Contains(strings.ReplaceAll(string(agentsContent), "\r\n", ""), "\n") {
		t.Fatalf("CRLF target gained bare LF: %q", agentsContent)
	}
	results, err = service.ApplyInstructions(wiki)
	if err != nil {
		t.Fatal(err)
	}
	for _, result := range results {
		if result.Changed {
			t.Errorf("idempotent result = %#v", result)
		}
	}
	if _, err := service.RemoveInstructions(wiki); err != nil {
		t.Fatal(err)
	}
	content, _ := os.ReadFile(agentsPath)
	if strings.Contains(string(content), InstructionBegin) || !strings.Contains(string(content), "owner policy") {
		t.Fatalf("remove changed outside content: %q", content)
	}
}

func TestInstructionUpdateConvergesDuplicatesAndRejectsMalformedMarkers(t *testing.T) {
	wiki := t.TempDir()
	writeWikiConfig(t, wiki, "wiki-test")
	path := filepath.Join(wiki, "AGENTS.md")
	duplicate := "before\n" + InstructionBegin + "\nstale\n" + InstructionEnd + "\nmiddle\n" + InstructionBegin + "\nold\n" + InstructionEnd + "\nafter\n"
	if err := os.WriteFile(path, []byte(duplicate), 0o644); err != nil {
		t.Fatal(err)
	}
	service := testService(t, t.TempDir())
	if _, err := service.ApplyInstructions(wiki); err != nil {
		t.Fatal(err)
	}
	content, _ := os.ReadFile(path)
	if strings.Count(string(content), InstructionBegin) != 1 || !strings.Contains(string(content), "before") || !strings.Contains(string(content), "middle") || !strings.Contains(string(content), "after") {
		t.Fatalf("duplicate convergence lost content: %s", content)
	}
	malformed := "owner\n" + InstructionBegin + "\nunclosed\n"
	if err := os.WriteFile(path, []byte(malformed), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := service.ApplyInstructions(wiki); err == nil || !strings.Contains(err.Error(), "refuse malformed") {
		t.Fatalf("malformed error = %v", err)
	}
	unchanged, _ := os.ReadFile(path)
	if string(unchanged) != malformed {
		t.Fatalf("malformed target was changed: %q", unchanged)
	}
}

func TestInstalledInstructionUsesCanonicalFilenameOnCaseSensitiveSystems(t *testing.T) {
	wiki := t.TempDir()
	writeWikiConfig(t, wiki, "wiki-test")
	service := testService(t, t.TempDir())
	if _, err := service.ApplyInstructions(wiki); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"agents.md", "AGENTS.MD", "AGENTS.md"} {
		body, err := service.InstalledInstruction(wiki, name)
		if err != nil || !strings.Contains(body, "wiki-test") {
			t.Fatalf("show %q = %q, %v", name, body, err)
		}
	}
}

func TestInstructionContextUsesStrictConfig(t *testing.T) {
	wiki := t.TempDir()
	bad := "schema: 1\nwiki_id: test\nbundle: knowledge\nderived: .buda\nqmd:\n  executable: qmd\n  collection: test\n  project_directory: .qmd\nunknown: true\n"
	if err := os.WriteFile(filepath.Join(wiki, config.FileName), []byte(bad), 0o644); err != nil {
		t.Fatal(err)
	}
	service := testService(t, t.TempDir())
	if _, err := service.ReadInstructionContext(wiki); err == nil || err.(*Error).ExitCode() != 3 {
		t.Fatalf("strict config error = %v", err)
	}
}
