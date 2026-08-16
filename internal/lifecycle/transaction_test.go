package lifecycle

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestAcquireLockIsExclusiveAndSafeRemoveIsContained(t *testing.T) {
	root := t.TempDir()
	lockPath := filepath.Join(root, "state", "upgrade.lock")
	first, err := AcquireLock(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Release()
	if _, err := AcquireLock(lockPath); err == nil {
		t.Fatal("second lock acquisition succeeded")
	}
	owned := filepath.Join(root, "owned")
	if err := os.MkdirAll(owned, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(owned, "file"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := SafeRemove(filepath.Join(owned, "file"), owned); err != nil {
		t.Fatal(err)
	}
	if err := SafeRemove(root, owned); err == nil {
		t.Fatal("removal outside owned root was accepted")
	}
}

func TestAcquireLockRecoversDeadOwnerAndReleaseChecksToken(t *testing.T) {
	root := t.TempDir()
	lockPath := filepath.Join(root, "upgrade.lock")
	data, err := json.Marshal(Lock{PID: 2147483647, Token: "dead-owner"})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(lockPath, append(data, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
	lock, err := AcquireLock(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Release()
	if err := os.WriteFile(lockPath, []byte(`{"pid":1,"token":"another-owner"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := lock.Release(); err == nil {
		t.Fatal("release removed a lock owned by another token")
	}
}
