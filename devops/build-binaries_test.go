package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestReleaseContractNamesAndTuning(t *testing.T) {
	want := []string{
		"buda-linux-amd64", "buda-linux-arm64", "buda-linux-armv7", "buda-linux-armv6",
		"buda-darwin-amd64", "buda-darwin-arm64", "buda-windows-amd64.exe", "buda-windows-arm64.exe",
	}
	got := make([]string, 0, len(targets))
	for _, target := range targets {
		got = append(got, target.name)
		if target.goarch == "amd64" && target.tuning != "GOAMD64=v1" {
			t.Fatalf("%s tuning = %s", target.name, target.tuning)
		}
		if target.goarch == "arm64" && target.tuning != "GOARM64=v8.0" {
			t.Fatalf("%s tuning = %s", target.name, target.tuning)
		}
		if strings.Contains(target.tuning, "v2") || strings.Contains(target.tuning, "v3") || strings.Contains(target.tuning, "v4") {
			t.Fatalf("unsupported performance tier in %s", target.tuning)
		}
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("targets = %#v, want %#v", got, want)
	}
	if skillAssetName != "guiho-s-0002-buda.zip" || instructionAssetName != "guiho-i-buda.md" || promptAssetName != "guiho-p-buda.md" {
		t.Fatalf("supporting assets = %s, %s, %s", skillAssetName, instructionAssetName, promptAssetName)
	}
}

func TestBuildEnvironmentIsPureGoAndTargetSpecific(t *testing.T) {
	environment := buildEnvironment(targets[2])
	joined := strings.Join(environment, "\n")
	for _, expected := range []string{"GOOS=linux", "GOARCH=arm", "GOARM=7", "CGO_ENABLED=0"} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("environment missing %s", expected)
		}
	}
	if filepath.Ext(targets[2].name) == ".exe" {
		t.Fatal("Linux target has Windows suffix")
	}
}

func TestVersionedResourcesUseTheExactReleaseVersion(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source.md")
	destination := filepath.Join(root, "destination.md")
	if err := os.WriteFile(source, []byte("version: \"0.2.0\"\nurl: /v0.2.0/example\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := copyVersionedFile(source, destination, "1.2.3-rc.1"); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(destination)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(data); got != "version: \"1.2.3-rc.1\"\nurl: /v1.2.3-rc.1/example\n" {
		t.Fatalf("versioned resource = %q", got)
	}
}
