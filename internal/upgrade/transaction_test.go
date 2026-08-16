package upgrade

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CGuiho/buda/internal/artifact"
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
