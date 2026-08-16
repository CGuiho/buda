package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

func TestGitHubWorkflowsAreValidYAMLAndReleaseOnlyToGitHub(t *testing.T) {
	paths := []string{
		filepath.Join("..", ".github", "workflows", "ci.yml"),
		filepath.Join("..", ".github", "workflows", "publish.yml"),
	}
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var document yaml.Node
		if err := yaml.Unmarshal(data, &document); err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		if document.Kind != yaml.DocumentNode || len(document.Content) != 1 {
			t.Fatalf("%s is not one YAML document", path)
		}

		text := string(data)
		forbidden := []*regexp.Regexp{
			regexp.MustCompile(`(?mi)^\s*(?:npm|pnpm|bun)\s+(?:publish|deploy)\b`),
			regexp.MustCompile(`(?mi)^\s*id-token:\s*write\s*$`),
			regexp.MustCompile(`(?mi)^\s*environment:\s*production\s*$`),
		}
		for _, pattern := range forbidden {
			if pattern.MatchString(text) {
				t.Fatalf("%s contains forbidden package or production publication %q", path, pattern)
			}
		}
	}

	publish, err := os.ReadFile(paths[1])
	if err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{
		"tags: ['buda/v*']",
		"contents: write",
		"Build complete release unit",
		"Verify public asset set and checksums",
		"gh release create",
		"buda/v",
	} {
		if !strings.Contains(string(publish), required) {
			t.Fatalf("publish workflow missing %q", required)
		}
	}

	ci, err := os.ReadFile(paths[0])
	if err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{
		"go test -count=1 ./...",
		"go vet ./...",
		"Build manifest-derived release unit",
		"xdocs doctor . --warnings-as-errors",
	} {
		if !strings.Contains(string(ci), required) {
			t.Fatalf("CI workflow missing %q", required)
		}
	}
}
