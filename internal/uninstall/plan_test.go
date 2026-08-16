package uninstall

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CGuiho/buda/internal/installlayout"
)

func TestBuildPlanPreservesSharedParentsAndSelectedDataClasses(t *testing.T) {
	home := t.TempDir()
	layout, err := installlayout.ForHome(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(layout.Launcher), 0o755); err != nil {
		t.Fatal(err)
	}
	plan, err := BuildPlan(layout, filepath.Join(home, "wiki"), true, true)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range plan.Items {
		if item.Path == layout.GUIHO || item.Path == layout.Bin || item.Path == layout.Temp {
			t.Fatalf("plan targets shared parent %s", item.Path)
		}
		if item.Path == layout.GlobalConfig && !item.Preserved {
			t.Fatal("global config was not preserved")
		}
	}
	if len(plan.Items) == 0 {
		t.Fatal("empty uninstall plan")
	}
}

func TestBuildPlanRejectsCorruptInstalledManifest(t *testing.T) {
	home := t.TempDir()
	layout, err := installlayout.ForHome(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(layout.InstalledManifest), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(layout.InstalledManifest, []byte(`{"schema":1,"cli":"buda","version":"0.2.0","artifacts":[`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := BuildPlan(layout, "", false, false); err == nil {
		t.Fatal("corrupt installed manifest was accepted")
	}
}

func TestBuildPlanKeepsCLIHomeWhenOnlyConfigurationIsPreserved(t *testing.T) {
	home := t.TempDir()
	layout, err := installlayout.ForHome(home)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := BuildPlan(layout, "", true, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range plan.Items {
		if item.Path == layout.CLIHome && item.Action != "PRESERVE" {
			t.Fatalf("CLI home action = %s, want PRESERVE", item.Action)
		}
		if item.Path == layout.GlobalConfig && !item.Preserved {
			t.Fatal("global configuration was not preserved")
		}
	}
}

func TestBuildPlanRefusesExistingCLIHomeWithoutOwnershipManifest(t *testing.T) {
	home := t.TempDir()
	layout, err := installlayout.ForHome(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(layout.CLIHome, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := BuildPlan(layout, "", false, false); err == nil {
		t.Fatal("uninstall plan accepted an existing CLI home without its ownership manifest")
	}
}

func TestBuildPlanNeverRecursivelyRemovesCLIHomeContainer(t *testing.T) {
	home := t.TempDir()
	layout, err := installlayout.ForHome(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(layout.CLIHome, 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := `{"schema":1,"cli":"buda","version":"0.2.0","artifacts":[{"id":"example","version":"0.2.0","path":"buda.example.yaml","installed_path":"versions/0.2.0/artifacts/buda.example.yaml","sha256":"` + strings.Repeat("a", 64) + `","ownership":"replaceable","replaceable":true,"persistent":false,"disposable":false}]}`
	if err := os.WriteFile(layout.InstalledManifest, []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	sentinel := filepath.Join(layout.CLIHome, "user-created.txt")
	if err := os.WriteFile(sentinel, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	plan, err := BuildPlan(layout, "", false, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range plan.Items {
		if item.Path == layout.CLIHome && item.Action != "PRESERVE" {
			t.Fatalf("CLI home container is destructive: %#v", item)
		}
	}
	if err := Apply(plan, layout.Temp); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(sentinel); err != nil {
		t.Fatalf("unmanaged CLI-home content was removed: %v", err)
	}
}

func TestApplyQuarantineRollbackOnFailure(t *testing.T) {
	home := t.TempDir()
	layout, err := installlayout.ForHome(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(layout.CLIHome, 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := `{"schema":1,"cli":"buda","version":"0.2.0","artifacts":[{"id":"example","version":"0.2.0","path":"buda.example.yaml","installed_path":"versions/0.2.0/artifacts/buda.example.yaml","sha256":"` + strings.Repeat("a", 64) + `","ownership":"replaceable","replaceable":true,"persistent":false,"disposable":false}]}`
	if err := os.WriteFile(layout.InstalledManifest, []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	target1 := filepath.Join(layout.CLIHome, "versions", "0.2.0", "artifacts", "buda.example.yaml")
	if err := os.MkdirAll(filepath.Dir(target1), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target1, []byte("content1"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Add an unremovable/failing item to simulate failure
	plan, err := BuildPlan(layout, "", false, false)
	if err != nil {
		t.Fatal(err)
	}
	// Inject an item with an invalid path that cannot be accessed or moved
	plan.Items = append(plan.Items, Item{
		Path:   filepath.Join(layout.CLIHome, "nonexistent-dir", "invalid\x00file"),
		Root:   layout.CLIHome,
		Action: "REMOVE",
		Owner:  "test-fail",
	})
	if err := Apply(plan, layout.Temp); err == nil {
		t.Fatal("Apply unexpectedly succeeded with invalid item")
	}
	// Verify target1 was restored by rollback
	if content, err := os.ReadFile(target1); err != nil || string(content) != "content1" {
		t.Fatalf("target1 was not restored after rollback: %q, %v", content, err)
	}
}
