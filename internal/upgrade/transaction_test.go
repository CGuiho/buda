package upgrade

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CGuiho/buda/internal/artifact"
	"github.com/CGuiho/buda/internal/installlayout"
	"github.com/CGuiho/buda/internal/lifecycle"
)

func TestVerifyChecksumsRejectsMissingDuplicateAndExtraEntries(t *testing.T) {
	directory := t.TempDir()
	file := filepath.Join(directory, "payload")
	if err := os.WriteFile(file, []byte("payload"), 0o644); err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256([]byte("payload"))
	files := map[string]string{"payload": file}
	valid := hex.EncodeToString(digest[:]) + "  payload\n"
	for name, checksums := range map[string]string{
		"valid":     valid,
		"missing":   "",
		"duplicate": valid + valid,
		"extra":     valid + strings.Repeat("a", 64) + "  other\n",
	} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(directory, name+".txt")
			if err := os.WriteFile(path, []byte(checksums), 0o644); err != nil {
				t.Fatal(err)
			}
			err := verifyChecksums(files, path)
			if name == "valid" && err != nil {
				t.Fatal(err)
			}
			if name != "valid" && err == nil {
				t.Fatal("invalid checksum set was accepted")
			}
		})
	}
}

func TestValidateManifestAssetsRequiresEveryPublishedArtifact(t *testing.T) {
	directory := t.TempDir()
	payloadPath := filepath.Join(directory, "payload")
	if err := os.WriteFile(payloadPath, []byte("payload"), 0o644); err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256([]byte("payload"))
	manifest := artifact.Manifest{Schema: 1, CLI: "buda", Version: "0.2.0", Artifacts: []artifact.Artifact{{
		ID: "payload", Version: "0.2.0", Path: "payload", InstalledPath: "versions/0.2.0/artifacts/payload",
		SHA256: hex.EncodeToString(digest[:]), Ownership: artifact.OwnershipReplaceable, Replaceable: true,
	}}}
	if err := validateManifestAssets(manifest, map[string]string{"payload": payloadPath}); err != nil {
		t.Fatal(err)
	}
	if err := validateManifestAssets(manifest, map[string]string{"payload": payloadPath, "retired": "retired"}); err == nil {
		t.Fatal("undeclared published artifact was accepted")
	}
}

func TestManifestDrivenProjectionsApplyAndRollback(t *testing.T) {
	home := t.TempDir()
	op := t.TempDir()
	versionDir := filepath.Join(op, "version")
	artifactsDir := filepath.Join(versionDir, "artifacts")
	if err := os.MkdirAll(artifactsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create a dummy text projection artifact
	artifactFile := filepath.Join(artifactsDir, "instruction.md")
	if err := os.WriteFile(artifactFile, []byte("v2 instruction"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Set up prior projection in home
	priorTarget := filepath.Join(home, ".agents", "instructions", "test.md")
	if err := os.MkdirAll(filepath.Dir(priorTarget), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(priorTarget, []byte("v1 instruction"), 0o644); err != nil {
		t.Fatal(err)
	}

	manifest := artifact.Manifest{
		Schema:  1,
		CLI:     "buda",
		Version: "0.2.0",
		Artifacts: []artifact.Artifact{
			{
				ID:              "instruction",
				Version:         "0.2.0",
				Path:            "instruction.md",
				InstalledPath:   "versions/0.2.0/artifacts/instruction.md",
				Ownership:       artifact.OwnershipReplaceable,
				Replaceable:     true,
				ProjectionPaths: []string{".agents/instructions/test.md"},
			},
		},
	}
	prevManifest := artifact.Manifest{
		Schema:  1,
		CLI:     "buda",
		Version: "0.1.0",
		Artifacts: []artifact.Artifact{
			{
				ID:              "instruction",
				Version:         "0.1.0",
				Path:            "instruction.md",
				InstalledPath:   "versions/0.1.0/artifacts/instruction.md",
				Ownership:       artifact.OwnershipReplaceable,
				Replaceable:     true,
				ProjectionPaths: []string{".agents/instructions/test.md"},
			},
		},
	}

	// 1. Snapshot projections
	snapshots, err := snapshotProjections(op, home, manifest, prevManifest)
	if err != nil {
		t.Fatalf("snapshotProjections error: %v", err)
	}
	if len(snapshots) != 1 || !snapshots[0].existed {
		t.Fatalf("expected 1 existing snapshot, got: %+v", snapshots)
	}

	// 2. Apply projections
	if err := applyProjections(op, home, versionDir, manifest, prevManifest); err != nil {
		t.Fatalf("applyProjections error: %v", err)
	}
	content, err := os.ReadFile(priorTarget)
	if err != nil || string(content) != "v2 instruction" {
		t.Fatalf("applied projection content = %q, want 'v2 instruction'", string(content))
	}

	// 3. Rollback projections
	if err := rollbackProjections(snapshots); err != nil {
		t.Fatalf("rollbackProjections error: %v", err)
	}
	rolledBackContent, err := os.ReadFile(priorTarget)
	if err != nil || string(rolledBackContent) != "v1 instruction" {
		t.Fatalf("rolled back content = %q, want 'v1 instruction'", string(rolledBackContent))
	}
}

func TestRecoverInterruptedJournalStates(t *testing.T) {
	setup := func(t *testing.T, activeVersion string, journal lifecycle.Journal) (installlayout.Layout, string) {
		t.Helper()
		layout, err := installlayout.ForHome(t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(layout.Versions, 0o755); err != nil {
			t.Fatal(err)
		}
		payloadName := "buda"
		if strings.Contains(activeVersion, ".exe") {
			payloadName = "buda.exe"
		}
		pointer := map[string]any{
			"schema": 1, "active": activeVersion + "/" + payloadName, "previous": "",
			"active_version": activeVersion, "previous_version": "",
		}
		data, _ := json.Marshal(pointer)
		if err := os.WriteFile(layout.Current, append(data, '\n'), 0o644); err != nil {
			t.Fatal(err)
		}
		payload := filepath.Join(layout.Versions, activeVersion, payloadName)
		if err := os.MkdirAll(filepath.Dir(payload), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(payload, []byte("payload"), 0o755); err != nil {
			t.Fatal(err)
		}
		journalPath := filepath.Join(layout.State, "transaction.json")
		if err := os.MkdirAll(layout.State, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := lifecycle.SaveJournal(journalPath, journal); err != nil {
			t.Fatal(err)
		}
		return layout, journalPath
	}

	t.Run("complete journal is removed", func(t *testing.T) {
		layout, journalPath := setup(t, "0.2.0", lifecycle.Journal{Operation: "upgrade", Phase: lifecycle.PhaseComplete, Token: "t", Version: "0.2.0", PreviousVersion: "0.1.1"})
		if err := recoverInterrupted(layout, journalPath); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(journalPath); !os.IsNotExist(err) {
			t.Fatalf("complete journal still exists: %v", err)
		}
	})
	t.Run("rolled-back journal is removed", func(t *testing.T) {
		layout, journalPath := setup(t, "0.1.1", lifecycle.Journal{Operation: "upgrade", Phase: lifecycle.PhaseRolledBack, Token: "t", Version: "0.2.0", PreviousVersion: "0.1.1"})
		if err := recoverInterrupted(layout, journalPath); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(journalPath); !os.IsNotExist(err) {
			t.Fatalf("rolled-back journal still exists: %v", err)
		}
	})
	t.Run("activated target wins and journal is removed", func(t *testing.T) {
		layout, journalPath := setup(t, "0.2.0", lifecycle.Journal{Operation: "upgrade", Phase: lifecycle.PhaseActivated, Token: "t", Version: "0.2.0", PreviousVersion: "0.1.1"})
		if err := recoverInterrupted(layout, journalPath); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(journalPath); !os.IsNotExist(err) {
			t.Fatalf("activated journal still exists: %v", err)
		}
	})
	t.Run("activated target with missing payload fails", func(t *testing.T) {
		layout, journalPath := setup(t, "0.2.0", lifecycle.Journal{Operation: "upgrade", Phase: lifecycle.PhaseActivated, Token: "t", Version: "0.2.0", PreviousVersion: "0.1.1"})
		payload := filepath.Join(layout.Versions, "0.2.0", "buda")
		if err := os.Remove(payload); err != nil {
			t.Fatal(err)
		}
		if err := recoverInterrupted(layout, journalPath); err == nil {
			t.Fatal("interrupted upgrade with missing activated payload was accepted")
		}
	})
	t.Run("previous active pointer is treated as aborted", func(t *testing.T) {
		layout, journalPath := setup(t, "0.1.1", lifecycle.Journal{Operation: "upgrade", Phase: lifecycle.PhaseArtifactsReplaced, Token: "t", Version: "0.2.0", PreviousVersion: "0.1.1"})
		if err := recoverInterrupted(layout, journalPath); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(journalPath); !os.IsNotExist(err) {
			t.Fatalf("aborted journal still exists: %v", err)
		}
	})
	t.Run("pointer journal disagreement fails closed", func(t *testing.T) {
		layout, journalPath := setup(t, "0.1.1", lifecycle.Journal{Operation: "upgrade", Phase: lifecycle.PhaseActivated, Token: "t", Version: "0.2.0", PreviousVersion: "0.1.0"})
		if err := recoverInterrupted(layout, journalPath); err == nil {
			t.Fatal("journal/pointer disagreement was accepted")
		}
	})
	t.Run("missing pointer fails closed", func(t *testing.T) {
		layout, journalPath := setup(t, "0.1.1", lifecycle.Journal{Operation: "upgrade", Phase: lifecycle.PhaseStaged, Token: "t", Version: "0.2.0", PreviousVersion: "0.1.1"})
		if err := os.Remove(layout.Current); err != nil {
			t.Fatal(err)
		}
		if err := recoverInterrupted(layout, journalPath); err == nil {
			t.Fatal("interrupted upgrade without a readable pointer was accepted")
		}
	})
}
