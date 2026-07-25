package main

import (
	"os"
	"strings"
	"testing"
)

func TestInstallersUseExactTagsAndRequiredArtifacts(t *testing.T) {
	for _, path := range []string{"install.sh", "install.ps1"} {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		text := string(content)
		for _, required := range []string{
			"guiho-s-0002-buda.zip", "checksums.txt", "qmd", "--wiki <path>",
			">=2.5.0,<3.0.0", "agent skill update", "buda-backup", "buda v",
		} {
			if !strings.Contains(text, required) {
				t.Fatalf("%s missing %q", path, required)
			}
		}
		if strings.Contains(text, `"v${VERSION}"`) || strings.Contains(text, `"v$Version"`) {
			t.Fatalf("%s invents a release tag prefix", path)
		}
		if strings.Contains(text, "__") {
			t.Fatalf("%s contains an unreplaced template token", path)
		}
	}
}
