package main

import (
	"os"
	"strings"
	"testing"
)

func TestLifecycleScriptsExposeConventionSelectorsAndOwnershipBoundaries(t *testing.T) {
	for _, path := range []string{"install.sh", "install.ps1", "uninstall.sh", "uninstall.ps1"} {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		text := string(content)
		required := []string{".guiho", ".temp"}
		if strings.HasPrefix(path, "uninstall") {
			if strings.HasSuffix(path, ".ps1") {
				required = append(required, "PreserveConfig", "PreserveData", "DryRun", "Yes")
			} else {
				required = append(required, "preserve-config", "preserve-data", "dry-run", "--yes")
			}
		} else {
			required = append(required, "artifacts.json", "checksums.txt")
		}
		for _, required := range required {
			if !strings.Contains(text, required) {
				t.Fatalf("%s missing %q", path, required)
			}
		}
		if strings.Contains(text, "agent skill update") || strings.Contains(text, "--keep-agent-resources") {
			t.Fatalf("%s contains an obsolete lifecycle route", path)
		}
	}
	installSh, _ := os.ReadFile("install.sh")
	for _, required := range []string{"--version", "--channel", "BUDA_RELEASE_ASSET_DIR", "buda-launcher-", "Release asset:", "Source:", "agent skill install"} {
		if !strings.Contains(string(installSh), required) {
			t.Fatalf("install.sh missing %q", required)
		}
	}
	for _, forbidden := range []string{"--wiki", "WIKI=", " init --wiki"} {
		if strings.Contains(string(installSh), forbidden) {
			t.Fatalf("install.sh performs project setup through %q", forbidden)
		}
	}
	installPS1, _ := os.ReadFile("install.ps1")
	for _, required := range []string{"-Version", "-Channel", "buda-launcher-", "Release asset:", "Source:", "agent skill install"} {
		if !strings.Contains(string(installPS1), required) {
			t.Fatalf("install.ps1 missing %q", required)
		}
	}
	for _, forbidden := range []string{"$Wiki", "$WikiId", "init --wiki"} {
		if strings.Contains(string(installPS1), forbidden) {
			t.Fatalf("install.ps1 performs project setup through %q", forbidden)
		}
	}
}

func TestReadmeLifecycleSectionsFollowConvention(t *testing.T) {
	content, err := os.ReadFile("../README.md")
	if err != nil {
		t.Fatal(err)
	}
	text := string(content)
	installStart := strings.Index(text, "## Install\n")
	uninstallStart := strings.Index(text, "## Uninstall\n")
	migrationStart := strings.Index(text, "## Migrate a 0.1.x direct-binary installation\n")
	if installStart < 0 || uninstallStart < 0 || migrationStart < 0 || !(installStart < uninstallStart && uninstallStart < migrationStart) {
		t.Fatal("README lifecycle sections are missing or out of order")
	}
	installSection := text[installStart:uninstallStart]
	if strings.Contains(strings.TrimPrefix(installSection, "## Install\n"), "\n## ") {
		t.Fatal("README Uninstall section is not directly after Install")
	}
	for _, required := range []string{
		"irm https://raw.githubusercontent.com/CGuiho/buda/main/devops/install.ps1 | iex",
		"curl -fsSL https://raw.githubusercontent.com/CGuiho/buda/main/devops/install.sh | sh",
		"Load the `guiho-p-buda-install` prompt and follow it in order to install the Buda CLI. Find it here: https://raw.githubusercontent.com/CGuiho/buda/main/prompts/guiho-p-buda-install.md",
		"buda --version",
		"does not select,\ninitialize, or modify a wiki",
	} {
		if !strings.Contains(installSection, required) {
			t.Fatalf("README Install section omits %q", required)
		}
	}
	if strings.Count(installSection, "devops/install.ps1") != 1 || strings.Count(installSection, "devops/install.sh") != 1 {
		t.Fatal("README Install section must contain one Windows and one macOS/Linux installer command")
	}
	uninstallSection := text[uninstallStart:migrationStart]
	for _, required := range []string{
		"devops/uninstall.ps1",
		"devops/uninstall.sh",
		"Load the `guiho-p-buda-uninstall` prompt and follow it in order to uninstall the Buda CLI. Find it here: https://raw.githubusercontent.com/CGuiho/buda/main/prompts/guiho-p-buda-uninstall.md",
		"buda uninstall --wiki <path> --dry-run",
		"buda uninstall --wiki <path> --yes",
		"--preserve-config --preserve-data --yes",
		"removes all Buda-owned installation data by default",
	} {
		if !strings.Contains(uninstallSection, required) {
			t.Fatalf("README Uninstall section omits %q", required)
		}
	}
	if strings.Count(uninstallSection, "devops/uninstall.ps1") != 1 || strings.Count(uninstallSection, "devops/uninstall.sh") != 1 {
		t.Fatal("README Uninstall section must contain one Windows and one macOS/Linux uninstaller command")
	}
}

func TestInstallPS1ReleaseDiscoverySurvivesWindowsPowerShell51(t *testing.T) {
	// Windows PowerShell 5.1 emits a top-level JSON array from Invoke-RestMethod
	// as a single non-enumerated pipeline object. Wrapping that call directly in
	// @(...) nests every release inside one element, so channel filtering always
	// fails with "No release found for channel stable." (issue #9). The response
	// must be captured first and flattened in a second step.
	installPS1, err := os.ReadFile("install.ps1")
	if err != nil {
		t.Fatal(err)
	}
	text := string(installPS1)
	if strings.Contains(text, "@(Invoke-RestMethod") {
		t.Fatal("install.ps1 wraps Invoke-RestMethod directly in @(...); Windows PowerShell 5.1 nests the release array and release discovery always fails")
	}
	for _, required := range []string{"$releasePage = Invoke-RestMethod", "$batch = @($releasePage)"} {
		if !strings.Contains(text, required) {
			t.Fatalf("install.ps1 missing %q", required)
		}
	}
}

func TestLifecycleScriptsRejectObsoleteMutableReleaseAssumptions(t *testing.T) {
	for _, path := range []string{"install.sh", "install.ps1", "uninstall.sh", "uninstall.ps1"} {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		text := string(content)
		for _, forbidden := range []string{"buda v", "exact 11", "exactly 11", "scheduled", "${TAG#v}"} {
			if strings.Contains(text, forbidden) {
				t.Fatalf("%s contains forbidden current contract text %q", path, forbidden)
			}
		}
	}
	installSh, _ := os.ReadFile("install.sh")
	if !strings.Contains(string(installSh), "armv6") || !strings.Contains(string(installSh), "armv7") {
		t.Fatal("install.sh lacks ARMv6 or ARMv7 platform support")
	}
	if !strings.Contains(string(installSh), `chmod 755 "$STAGE/$BINARY"`) {
		t.Fatal("install.sh does not chmod candidate binary before invocation")
	}
	uninstallSh, _ := os.ReadFile("uninstall.sh")
	if !strings.Contains(string(uninstallSh), "mv -f") {
		t.Fatal("uninstall.sh does not use atomic move for instruction edits")
	}
	if !strings.Contains(string(uninstallSh), "Buda Uninstall Plan:") || !strings.Contains(string(uninstallSh), "REMOVE:") || !strings.Contains(string(uninstallSh), "PRESERVE:") {
		t.Fatal("uninstall.sh lacks grouped plan output")
	}
	uninstallPs1, _ := os.ReadFile("uninstall.ps1")
	if !strings.Contains(string(uninstallPs1), "Move-Item") {
		t.Fatal("uninstall.ps1 does not use atomic move for instruction edits")
	}
	if !strings.Contains(string(uninstallPs1), "Buda Uninstall Plan:") || !strings.Contains(string(uninstallPs1), "REMOVE:") || !strings.Contains(string(uninstallPs1), "PRESERVE:") {
		t.Fatal("uninstall.ps1 lacks grouped plan output")
	}
}
