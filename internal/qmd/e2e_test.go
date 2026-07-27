package qmd

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestOfficialQMDEndToEnd(t *testing.T) {
	if os.Getenv("BUDA_QMD_E2E") != "1" {
		t.Skip("set BUDA_QMD_E2E=1 to exercise an installed official qmd runtime")
	}
	executable := os.Getenv("BUDA_QMD_EXECUTABLE")
	if executable == "" {
		var err error
		executable, err = exec.LookPath("qmd")
		if err != nil {
			t.Fatal("BUDA_QMD_EXECUTABLE is unset and qmd is not on PATH")
		}
	}

	root := t.TempDir()
	if base := os.Getenv("BUDA_QMD_E2E_ROOT"); base != "" {
		createdRoot, err := os.MkdirTemp(base, "buda-qmd-e2e-")
		if err != nil {
			t.Fatal(err)
		}
		root = createdRoot
		t.Cleanup(func() { _ = os.RemoveAll(root) })
	}
	bundle := filepath.Join(root, "knowledge")
	concepts := filepath.Join(bundle, "concepts")
	if err := os.MkdirAll(concepts, 0o755); err != nil {
		t.Fatal(err)
	}
	concept := []byte("---\ntitle: API Policy\n---\n\n# API Policy\n\nBuda preserves durable external identifiers and cited evidence.\n")
	if err := os.WriteFile(filepath.Join(concepts, "api-policy.md"), concept, 0o644); err != nil {
		t.Fatal(err)
	}

	adapter, err := New(Config{
		Executable:       executable,
		WikiRoot:         root,
		BundleRoot:       bundle,
		ProjectDirectory: filepath.Join(root, ".qmd"),
		Collection:       "buda-wiki",
	}, OSRunner{})
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("QMD_DOCTOR_DEVICE_PROBE", "0")

	compatibility, err := adapter.CheckCompatibility(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if compatibility.Version.Raw == "" {
		t.Fatal("qmd compatibility omitted the runtime version")
	}
	if err := adapter.EnsureProject(context.Background()); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"index.yml", "index.sqlite"} {
		if _, err := os.Stat(filepath.Join(root, ".qmd", name)); err != nil {
			t.Fatalf("project-local qmd artifact %s: %v", name, err)
		}
	}
	if _, err := adapter.Update(context.Background()); err != nil {
		t.Fatal(err)
	}

	matches, err := adapter.Search(context.Background(), SearchOptions{
		Mode: ModeLexical, Text: "durable external identifiers", Limit: 5,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 1 || matches[0].Path != "concepts/api-policy.md" || matches[0].DocumentID == "" {
		t.Fatalf("lexical matches = %#v", matches)
	}
	document, err := adapter.Get(context.Background(), matches[0].DocumentID)
	if err != nil {
		t.Fatal(err)
	}
	if document.Path != "concepts/api-policy.md" || document.Body == "" {
		t.Fatalf("document = %#v", document)
	}
	status, err := adapter.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if status.Documents != 1 || status.Pending != 1 {
		t.Fatalf("status = %#v", status)
	}
	doctor, err := adapter.Doctor(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if doctor.Checks == 0 {
		t.Fatalf("doctor = %#v", doctor)
	}
}
