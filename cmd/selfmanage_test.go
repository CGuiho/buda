package cmd

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/CGuiho/buda/internal/selfmanage"
)

type releaseTransport func(*http.Request) (*http.Response, error)

func (transport releaseTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	return transport(request)
}

func catalogClient() *http.Client {
	return &http.Client{Transport: releaseTransport(func(request *http.Request) (*http.Response, error) {
		body := `[{
 "tag_name":"buda/v0.0.2","name":"Buda 0.0.2","published_at":"2026-07-27T20:00:00Z",
 "html_url":"https://github.com/CGuiho/buda/releases/tag/buda/v0.0.2",
 "assets":[
  {"name":"buda-linux-amd64","browser_download_url":"https://downloads.example/buda-linux-amd64"},
  {"name":"checksums.txt","browser_download_url":"https://downloads.example/checksums.txt"}
 ]
}]`
		return &http.Response{StatusCode: http.StatusOK, Status: "200 OK", Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}
}

func upgradeClient(binary []byte) *http.Client {
	digest := sha256.Sum256(binary)
	return &http.Client{Transport: releaseTransport(func(request *http.Request) (*http.Response, error) {
		body := ""
		switch request.URL.Host {
		case "api.github.com":
			body = `[{"tag_name":"buda/v0.0.2","assets":[{"name":"buda-linux-amd64","browser_download_url":"https://downloads.example/buda"},{"name":"checksums.txt","browser_download_url":"https://downloads.example/checksums.txt"}]}]`
		case "downloads.example":
			if strings.HasSuffix(request.URL.Path, "checksums.txt") {
				body = hex.EncodeToString(digest[:]) + "  buda-linux-amd64\n"
			} else {
				body = string(binary)
			}
		}
		return &http.Response{StatusCode: http.StatusOK, Status: "200 OK", Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header), ContentLength: int64(len(body))}, nil
	})}
}

func TestUpgradeDryRunUsesCanonicalReleaseWithoutWiki(t *testing.T) {
	var output bytes.Buffer
	deps := Dependencies{Out: &output, Err: &bytes.Buffer{}, Options: &Options{}, HTTPClient: catalogClient(), Executable: func() (string, error) { return "buda", nil }}
	root := NewRootCommand(deps, BuildInfo{Version: "0.0.1", Target: "buda-linux-amd64"})
	root.SetArgs([]string{"upgrade", "--dry-run"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"Current: 0.0.1", "Target: 0.0.2", "Asset: buda-linux-amd64", "Dry run: true"} {
		if !strings.Contains(output.String(), expected) {
			t.Fatalf("upgrade preview missing %q: %s", expected, output.String())
		}
	}
}

func TestUpgradeCheckAndListEmitOneJSONDocument(t *testing.T) {
	for _, arguments := range [][]string{{"--json", "upgrade", "check"}, {"--json", "upgrade", "list"}} {
		var output bytes.Buffer
		deps := Dependencies{Out: &output, Err: &bytes.Buffer{}, Options: &Options{}, HTTPClient: catalogClient()}
		root := NewRootCommand(deps, BuildInfo{Version: "0.0.1", Target: "buda-linux-amd64"})
		root.SetArgs(arguments)
		if err := root.Execute(); err != nil {
			t.Fatalf("%v: %v", arguments, err)
		}
		if strings.Count(strings.TrimSpace(output.String()), "\n{") != 0 || !strings.Contains(output.String(), `"command": "buda upgrade`) {
			t.Fatalf("%v did not emit one JSON document: %s", arguments, output.String())
		}
	}
}

func TestUninstallDryRunNeedsNoWikiAndDoesNotDelete(t *testing.T) {
	var output bytes.Buffer
	removed := false
	deps := Dependencies{
		Out: &output, Err: &bytes.Buffer{}, Options: &Options{},
		Executable:       func() (string, error) { return "buda", nil },
		RemoveExecutable: func(string) (bool, error) { removed = true; return false, nil },
	}
	root := NewRootCommand(deps, BuildInfo{Version: "0.0.1"})
	root.SetArgs([]string{"uninstall", "--dry-run"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if removed || !strings.Contains(output.String(), "Would remove executable:") || !strings.Contains(output.String(), "Would remove both global Buda skills") {
		t.Fatalf("uninstall preview = %q, removed=%t", output.String(), removed)
	}
}

func TestUpgradeAndUninstallNeverScheduleResourceMaintenance(t *testing.T) {
	for _, arguments := range [][]string{{"upgrade", "--dry-run"}, {"uninstall", "--dry-run"}} {
		scheduled := 0
		deps := Dependencies{
			Out: &bytes.Buffer{}, Err: &bytes.Buffer{}, Options: &Options{}, HTTPClient: catalogClient(),
			Executable:          func() (string, error) { return "buda", nil },
			ScheduleMaintenance: func(string, string) error { scheduled++; return nil },
		}
		root := NewRootCommand(deps, BuildInfo{Version: "0.0.1", Target: "buda-linux-amd64"})
		root.SetArgs(arguments)
		if err := root.Execute(); err != nil {
			t.Fatalf("%v: %v", arguments, err)
		}
		if scheduled != 0 {
			t.Fatalf("%v scheduled hidden maintenance", arguments)
		}
	}
}

func TestUpgradeEqualVersionIsNoOp(t *testing.T) {
	var output bytes.Buffer
	called := false
	deps := Dependencies{
		Out: &output, Err: &bytes.Buffer{}, Options: &Options{}, HTTPClient: catalogClient(),
		UpgradeBinary: func(context.Context, selfmanage.UpgradeOptions) (selfmanage.UpgradeResult, error) {
			called = true
			return selfmanage.UpgradeResult{}, nil
		},
	}
	root := NewRootCommand(deps, BuildInfo{Version: "0.0.2", Target: "buda-linux-amd64"})
	root.SetArgs([]string{"upgrade"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if called || !strings.Contains(output.String(), "already current") {
		t.Fatalf("equal upgrade output=%q called=%t", output.String(), called)
	}
}

func TestUpgradeTextReportsOperationalProgress(t *testing.T) {
	var output bytes.Buffer
	deps := Dependencies{
		Out: &output, Err: &bytes.Buffer{}, Options: &Options{}, HTTPClient: upgradeClient([]byte("binary")),
		Executable: func() (string, error) { return "C:/tools/buda.exe", nil },
		UpgradeBinary: func(_ context.Context, options selfmanage.UpgradeOptions) (selfmanage.UpgradeResult, error) {
			options.Progress(selfmanage.DownloadProgress{Bytes: 6, Total: 6, Percent: 100})
			return selfmanage.UpgradeResult{Executable: options.Executable, Backup: options.Executable + ".old", TargetVersion: options.TargetVersion}, nil
		},
		ReconcileInstalled: func(string, string) error { return nil },
	}
	root := NewRootCommand(deps, BuildInfo{Version: "0.0.1", Target: "buda-linux-amd64"})
	root.SetArgs([]string{"upgrade"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"Target: 0.0.2", "URL: https://downloads.example/buda", "Checksum verified: sha256:", "Installation path: C:/tools/buda.exe", "Download progress: 100.0%", "Agent resources: reconciled", "Final version verified: buda v0.0.2"} {
		if !strings.Contains(output.String(), expected) {
			t.Fatalf("upgrade text missing %q:\n%s", expected, output.String())
		}
	}
}

func TestReconciliationFailureAutomaticallyRollsBack(t *testing.T) {
	rolledBack := false
	deps := Dependencies{
		Out: &bytes.Buffer{}, Err: &bytes.Buffer{}, Options: &Options{}, HTTPClient: upgradeClient([]byte("binary")),
		Executable: func() (string, error) { return "C:/tools/buda.exe", nil },
		UpgradeBinary: func(_ context.Context, options selfmanage.UpgradeOptions) (selfmanage.UpgradeResult, error) {
			return selfmanage.UpgradeResult{Executable: options.Executable, Backup: options.Executable + ".old"}, nil
		},
		ReconcileInstalled: func(string, string) error { return errors.New("skill write failed") },
		RollbackExecutable: func(string) (bool, error) { rolledBack = true; return false, nil },
	}
	root := NewRootCommand(deps, BuildInfo{Version: "0.0.1", Target: "buda-linux-amd64"})
	root.SetArgs([]string{"upgrade"})
	err := root.Execute()
	if !rolledBack || err == nil || !strings.Contains(err.Error(), "automatically rolled back") {
		t.Fatalf("reconcile failure=%v rolledBack=%t", err, rolledBack)
	}
}

func TestUpgradeRollbackRouteUsesInjectedBoundary(t *testing.T) {
	called := false
	var output bytes.Buffer
	deps := Dependencies{
		Out: &output, Err: &bytes.Buffer{}, Options: &Options{},
		Executable:         func() (string, error) { return "C:/tools/buda.exe", nil },
		RollbackExecutable: func(string) (bool, error) { called = true; return false, nil },
	}
	root := NewRootCommand(deps, BuildInfo{Version: "0.0.2"})
	root.SetArgs([]string{"upgrade", "rollback"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if !called || !strings.Contains(output.String(), "Restored the previous") {
		t.Fatalf("rollback output=%q called=%t", output.String(), called)
	}
}
