package artifact

import (
	"strings"
	"testing"
)

func validManifest() Manifest {
	return Manifest{
		Schema: 1, CLI: "buda", Version: "0.2.0",
		Artifacts: []Artifact{{
			ID: "payload", Version: "0.2.0", Path: "buda-linux-amd64",
			InstalledPath: "versions/0.2.0/buda", SHA256: strings.Repeat("a", 64),
			Ownership: OwnershipReplaceable, Replaceable: true,
		}},
	}
}

func TestManifestRejectsUnsafeAndInconsistentOwnership(t *testing.T) {
	cases := []struct {
		name string
		edit func(*Manifest)
	}{
		{"absolute path", func(value *Manifest) { value.Artifacts[0].Path = `C:\escape` }},
		{"volume-relative path", func(value *Manifest) { value.Artifacts[0].Path = `C:escape` }},
		{"rooted backslash path", func(value *Manifest) { value.Artifacts[0].Path = `\escape` }},
		{"traversal", func(value *Manifest) { value.Artifacts[0].InstalledPath = "versions/../escape" }},
		{"version drift", func(value *Manifest) { value.Artifacts[0].Version = "0.1.1" }},
		{"ownership flags", func(value *Manifest) { value.Artifacts[0].Persistent = true }},
		{"duplicate release path", func(value *Manifest) { value.Artifacts = append(value.Artifacts, validManifest().Artifacts[0]) }},
		{"all-zero digest on regular artifact", func(value *Manifest) { value.Artifacts[0].SHA256 = strings.Repeat("0", 64) }},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			value := validManifest()
			test.edit(&value)
			if err := value.Validate(); err == nil {
				t.Fatal("Validate unexpectedly succeeded")
			}
		})
	}
}

func TestManifestRoundTripUsesStrictJSON(t *testing.T) {
	value := validManifest()
	data, err := Encode(value)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := Decode(strings.NewReader(string(data)))
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Artifacts[0].ID != "payload" {
		t.Fatalf("decoded artifact = %#v", decoded.Artifacts[0])
	}
	if _, err := Decode(strings.NewReader(`{"schema":1,"cli":"buda","version":"0.2.0","artifacts":[],"unknown":true}`)); err == nil {
		t.Fatal("unknown manifest field was accepted")
	}
}

func TestManifestAllowsAllZeroForArtifactsJSON(t *testing.T) {
	manifest := Manifest{
		Schema: 1, CLI: "buda", Version: "0.2.0",
		Artifacts: []Artifact{
			{
				ID: "payload", Version: "0.2.0", Path: "buda-linux-amd64",
				InstalledPath: "versions/0.2.0/buda", SHA256: strings.Repeat("a", 64),
				Ownership: OwnershipReplaceable, Replaceable: true,
			},
			{
				ID: "artifacts-json", Version: "0.2.0", Path: "artifacts.json",
				InstalledPath: "versions/0.2.0/artifacts/artifacts.json", SHA256: strings.Repeat("0", 64),
				Ownership: OwnershipReplaceable, Replaceable: true,
			},
		},
	}
	if err := manifest.Validate(); err != nil {
		t.Fatalf("Validate() failed for manifest with all-zero artifacts.json: %v", err)
	}
}
