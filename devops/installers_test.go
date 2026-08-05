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
			`& $QmdSource --version`,
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

func TestInstallersVerifyBinaryOnPath(t *testing.T) {
	for _, path := range []string{"install.sh", "install.ps1"} {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		text := string(content)
		// Both installers must warn (not fail) when the binary is not callable
		// on PATH after installation, and suggest buda doctor --wiki <path>.
		if !strings.Contains(text, "command -v") && !strings.Contains(text, "Get-Command $CliName") {
			t.Fatalf("%s missing binary-on-PATH verification (command -v or Get-Command)", path)
		}
		if !strings.Contains(text, "not on PATH") && !strings.Contains(text, "not on the session PATH") {
			t.Fatalf("%s missing PATH warning message", path)
		}
		if !strings.Contains(text, "buda doctor --wiki <path>") {
			t.Fatalf("%s missing buda doctor suggestion in PATH warning", path)
		}
	}
}

func TestInstallersProbeQmdOffPath(t *testing.T) {
	for _, path := range []string{"install.sh", "install.ps1"} {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		text := string(content)
		// Both installers must probe common off-PATH qmd locations before
		// claiming qmd is not installed, and print a PATH-fix hint when found.
		if !strings.Contains(text, ".hermes") {
			t.Fatalf("%s missing ~/.hermes probe for off-PATH qmd", path)
		}
		if !strings.Contains(text, "was not on PATH but was found") && !strings.Contains(text, "was not on PATH") {
			t.Fatalf("%s missing off-PATH qmd found hint in %s", path, path)
		}
		// The hard-fail message must still exist for the truly-absent case.
		if !strings.Contains(text, "qmd is required but not installed") {
			t.Fatalf("%s missing qmd not-installed hard-fail message", path)
		}
	}
}

func TestInstallersConfigurableSkillDirs(t *testing.T) {
	for _, path := range []string{"install.sh", "install.ps1"} {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		text := string(content)
		// BUDA_SKILL_DIRS env var for extra destinations.
		if !strings.Contains(text, "BUDA_SKILL_DIRS") {
			t.Fatalf("%s missing BUDA_SKILL_DIRS configurable destinations", path)
		}
		// HERMES_SKILLS_DIR env var / default for Hermes-aware registration.
		if !strings.Contains(text, "HERMES_SKILLS_DIR") {
			t.Fatalf("%s missing HERMES_SKILLS_DIR Hermes-aware registration", path)
		}
		if !strings.Contains(text, ".hermes/skills") && !strings.Contains(text, ".hermes\\skills") {
			t.Fatalf("%s missing default Hermes skills directory", path)
		}
		// The registration must copy the skill directory into destinations.
		if !strings.Contains(text, "register_skill_dir") && !strings.Contains(text, "Register-SkillDir") {
			t.Fatalf("%s missing skill registration function", path)
		}
	}
}
