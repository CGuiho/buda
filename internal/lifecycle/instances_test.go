package lifecycle

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRegisterInstanceOwnsTokenizedRecordAndCleansIt(t *testing.T) {
	directory := t.TempDir()
	cleanup, err := RegisterInstance(directory, os.Args[0], "0.2.0")
	if err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || filepath.Ext(entries[0].Name()) != ".json" {
		t.Fatalf("registry entries = %#v", entries)
	}
	if err := cleanup(); err != nil {
		t.Fatal(err)
	}
	entries, err = os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("registry was not cleaned: %#v", entries)
	}
}

func TestTerminateOtherInstancesNeverTargetsInvokingPID(t *testing.T) {
	if err := TerminateOtherInstances(t.TempDir(), os.Getpid(), os.Args[0]); err != nil {
		t.Fatal(err)
	}
}
