package prompts

import (
	"io/fs"
	"regexp"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

func TestInstructionKeepsExplicitWikiAndBudaBoundary(t *testing.T) {
	data, err := fs.ReadFile(FS, "guiho-i-buda.md")
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, expected := range []string{
		"{{WIKI_ID}}", "{{BUNDLE}}", "--wiki", "guiho-s-buda",
		"Buda is the required tool", "evaluate the returned evidence", "do not bypass Buda",
		"Never infer another wiki",
	} {
		if !strings.Contains(text, expected) {
			t.Errorf("instruction prompt omits %q", expected)
		}
	}
	if strings.Contains(text, "`qmd ") {
		t.Fatal("instruction contains a direct qmd command")
	}
}

func TestPromptArtifactsUseCanonicalNamesAndMetadataVersions(t *testing.T) {
	namePattern := regexp.MustCompile(`^guiho-p-[a-z0-9]+(?:-[a-z0-9]+)*$`)
	for _, path := range []string{"guiho-p-buda.md", "guiho-p-buda-install.md", "guiho-p-buda-uninstall.md"} {
		data, err := fs.ReadFile(FS, path)
		if err != nil {
			t.Fatal(err)
		}
		parts := strings.SplitN(string(data), "---", 3)
		if len(parts) != 3 {
			t.Fatalf("%s is missing YAML frontmatter", path)
		}
		var frontmatter struct {
			Name        string `yaml:"name"`
			Description string `yaml:"description"`
			Metadata    struct {
				Version string `yaml:"version"`
			} `yaml:"metadata"`
		}
		if err := yaml.Unmarshal([]byte(parts[1]), &frontmatter); err != nil {
			t.Fatalf("decode %s: %v", path, err)
		}
		if !namePattern.MatchString(frontmatter.Name) || strings.TrimSuffix(path, ".md") != frontmatter.Name {
			t.Fatalf("%s name = %q", path, frontmatter.Name)
		}
		if frontmatter.Description == "" || frontmatter.Metadata.Version == "" {
			t.Fatalf("%s lacks required artifact metadata", path)
		}
	}
}

func TestInstallPromptStopsBeforeWikiInitialization(t *testing.T) {
	data, err := fs.ReadFile(FS, "guiho-p-buda-install.md")
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, required := range []string{"irm https://raw.githubusercontent.com/CGuiho/buda/main/devops/install.ps1 | iex", "curl -fsSL https://raw.githubusercontent.com/CGuiho/buda/main/devops/install.sh | sh", "buda --version", "Never install Buda through", "manual fallback", "Do not run `buda init`"} {
		if !strings.Contains(text, required) {
			t.Fatalf("install prompt omits %q", required)
		}
	}
}

func TestUninstallPromptRequiresPlanAndPreservesCanonicalKnowledge(t *testing.T) {
	data, err := fs.ReadFile(FS, "guiho-p-buda-uninstall.md")
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, required := range []string{"REMOVE", "PRESERVE", "--dry-run", "--preserve-config --preserve-data", "canonical OKF knowledge", "Ask for confirmation"} {
		if !strings.Contains(text, required) {
			t.Fatalf("uninstall prompt omits %q", required)
		}
	}
}
