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
			">=2.5.0,<3.0.0", "@tobilu/qmd@2.5.3", "agent skill update", "buda-backup", "buda v",
			"BUDA_RELEASE_ASSET_DIR", "buda/v0.0.2", "buda/v<semver>",
		} {
			if !strings.Contains(text, required) {
				t.Fatalf("%s missing %q", path, required)
			}
		}
		if strings.Contains(text, `"v${VERSION}"`) || strings.Contains(text, `"v$Version"`) {
			t.Fatalf("%s invents a release tag prefix", path)
		}
		if strings.Contains(text, `${TAG#v}`) || strings.Contains(text, `.TrimStart("v")`) {
			t.Fatalf("%s strips only a leading v and cannot verify canonical Buda tags", path)
		}
		if strings.Contains(text, "__") {
			t.Fatalf("%s contains an unreplaced template token", path)
		}
	}
}

func TestInstallersExtractVersionFromCanonicalBudaTag(t *testing.T) {
	tests := map[string][]string{
		"install.sh": {
			`EXPECTED_VERSION="${TAG#buda/v}"`,
			`"buda v${EXPECTED_VERSION}"`,
		},
		"install.ps1": {
			`^buda/v(?<Version>`,
			`$ExpectedVersion = $Matches.Version`,
			`"buda v$ExpectedVersion"`,
			`Get-Command qmd.cmd -CommandType Application`,
			`& $Qmd.Source --version`,
			`Buda is installed, but the user PATH could not be updated.`,
		},
	}
	for path, required := range tests {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, value := range required {
			if !strings.Contains(string(content), value) {
				t.Fatalf("%s missing canonical tag verification fragment %q", path, value)
			}
		}
	}
}
