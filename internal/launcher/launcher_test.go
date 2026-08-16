package launcher

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CGuiho/buda/internal/installlayout"
)

func TestPointerRejectsAbsoluteAndTraversal(t *testing.T) {
	for _, data := range []string{
		`{"schema":1,"active":"C:\\escape","active_version":"0.2.0"}`,
		`{"schema":1,"active":"C:escape","active_version":"0.2.0"}`,
		`{"schema":1,"active":"\\escape","active_version":"0.2.0"}`,
		`{"schema":1,"active":"../escape","active_version":"0.2.0"}`,
		`{"schema":1,"active":"0.2.0/../escape","active_version":"0.2.0"}`,
	} {
		path := filepath.Join(t.TempDir(), "current.json")
		if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := ReadPointer(path); err == nil {
			t.Fatalf("unsafe pointer accepted: %s", data)
		}
	}
}

func TestReadPointerRequiresVersionedPayload(t *testing.T) {
	path := filepath.Join(t.TempDir(), "current.json")
	data, err := json.Marshal(Pointer{Schema: 1, Active: "0.2.0/buda", ActiveVersion: "0.2.0"})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	pointer, err := ReadPointer(path)
	if err != nil {
		t.Fatal(err)
	}
	if pointer.ActiveVersion != "0.2.0" {
		t.Fatalf("pointer = %#v", pointer)
	}
}

func TestReadPointerRejectsVersionMismatch(t *testing.T) {
	path := filepath.Join(t.TempDir(), "current.json")
	if err := os.WriteFile(path, []byte(`{"schema":1,"active":"0.2.0/buda","active_version":"0.1.1"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadPointer(path); err == nil {
		t.Fatal("pointer version mismatch was accepted")
	}
}

func TestReadPointerRequiresStrictSemVerVersions(t *testing.T) {
	for _, version := range []string{"latest", "1.2", "1.02.3", "1.2.3-01"} {
		path := filepath.Join(t.TempDir(), "current.json")
		data := `{"schema":1,"active":"` + version + `/buda","active_version":"` + version + `"}`
		if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := ReadPointer(path); err == nil {
			t.Fatalf("invalid pointer version accepted: %s", version)
		}
	}
}

func TestRunRejectsSymlinkedPayload(t *testing.T) {
	root := t.TempDir()
	layout, err := installlayout.ForHome(filepath.Join(root, "home"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(layout.Versions, "0.2.0"), 0o755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(root, "outside")
	if err := os.WriteFile(target, []byte("not a payload"), 0o644); err != nil {
		t.Fatal(err)
	}
	symlink := filepath.Join(layout.Versions, "0.2.0", "buda")
	if err := os.Symlink(target, symlink); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	pointer := Pointer{Schema: 1, Active: "0.2.0/buda", ActiveVersion: "0.2.0"}
	data, err := json.Marshal(pointer)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(layout.CLIHome, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(layout.Current, data, 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errOut bytes.Buffer
	if code := Run(nil, strings.NewReader(""), &out, &errOut, layout); code == 0 || !strings.Contains(errOut.String(), "symlink") {
		t.Fatalf("symlinked payload result code=%d stderr=%q", code, errOut.String())
	}
}
