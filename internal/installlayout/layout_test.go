package installlayout

import (
	"path/filepath"
	"testing"
)

func TestForHomeUsesNativeBudaOwnershipBoundaries(t *testing.T) {
	home := filepath.Join(t.TempDir(), "home")
	layout, err := ForHome(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := layout.Validate(); err != nil {
		t.Fatal(err)
	}
	if layout.CLIHome != filepath.Join(home, ".guiho", "buda") || layout.Bin != filepath.Join(home, ".guiho", "bin") {
		t.Fatalf("unexpected layout: %#v", layout)
	}
	if _, err := layout.Operation("upgrade"); err != nil {
		t.Fatal(err)
	}
	if _, err := layout.Operation("../escape"); err == nil {
		t.Fatal("unsafe operation prefix accepted")
	}
}
