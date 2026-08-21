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
	for _, required := range []string{"--version", "--channel", "--wiki", "BUDA_RELEASE_ASSET_DIR", "buda.global.yaml", "buda-launcher-"} {
		if !strings.Contains(string(installSh), required) {
			t.Fatalf("install.sh missing %q", required)
		}
	}
	installPS1, _ := os.ReadFile("install.ps1")
	for _, required := range []string{"-Version", "-Channel", "$Wiki", "buda.global.yaml", "buda-launcher-"} {
		if !strings.Contains(string(installPS1), required) {
			t.Fatalf("install.ps1 missing %q", required)
		}
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
