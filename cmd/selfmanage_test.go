package cmd

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CGuiho/buda/internal/installlayout"
	"github.com/CGuiho/buda/internal/repository"
	"github.com/CGuiho/buda/internal/upgrade"
)

type releaseTransport func(*http.Request) (*http.Response, error)

func (transport releaseTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	return transport(request)
}

func completeReleaseAssetsJSON() string {
	assets := []string{
		`{"name":"buda-linux-amd64","browser_download_url":"https://downloads.example/buda-linux-amd64"}`,
		`{"name":"buda-linux-arm64","browser_download_url":"https://downloads.example/buda-linux-arm64"}`,
		`{"name":"buda-linux-armv7","browser_download_url":"https://downloads.example/buda-linux-armv7"}`,
		`{"name":"buda-linux-armv6","browser_download_url":"https://downloads.example/buda-linux-armv6"}`,
		`{"name":"buda-darwin-amd64","browser_download_url":"https://downloads.example/buda-darwin-amd64"}`,
		`{"name":"buda-darwin-arm64","browser_download_url":"https://downloads.example/buda-darwin-arm64"}`,
		`{"name":"buda-windows-amd64.exe","browser_download_url":"https://downloads.example/buda-windows-amd64.exe"}`,
		`{"name":"buda-windows-arm64.exe","browser_download_url":"https://downloads.example/buda-windows-arm64.exe"}`,
		`{"name":"buda-launcher-linux-amd64","browser_download_url":"https://downloads.example/buda-launcher-linux-amd64"}`,
		`{"name":"buda-launcher-linux-arm64","browser_download_url":"https://downloads.example/buda-launcher-linux-arm64"}`,
		`{"name":"buda-launcher-linux-armv7","browser_download_url":"https://downloads.example/buda-launcher-linux-armv7"}`,
		`{"name":"buda-launcher-linux-armv6","browser_download_url":"https://downloads.example/buda-launcher-linux-armv6"}`,
		`{"name":"buda-launcher-darwin-amd64","browser_download_url":"https://downloads.example/buda-launcher-darwin-amd64"}`,
		`{"name":"buda-launcher-darwin-arm64","browser_download_url":"https://downloads.example/buda-launcher-darwin-arm64"}`,
		`{"name":"buda-launcher-windows-amd64.exe","browser_download_url":"https://downloads.example/buda-launcher-windows-amd64.exe"}`,
		`{"name":"buda-launcher-windows-arm64.exe","browser_download_url":"https://downloads.example/buda-launcher-windows-arm64.exe"}`,
		`{"name":"buda.schema.json","browser_download_url":"https://downloads.example/buda.schema.json"}`,
		`{"name":"buda.global.schema.json","browser_download_url":"https://downloads.example/buda.global.schema.json"}`,
		`{"name":"buda.example.yaml","browser_download_url":"https://downloads.example/buda.example.yaml"}`,
		`{"name":"buda.global.example.yaml","browser_download_url":"https://downloads.example/buda.global.example.yaml"}`,
		`{"name":"guiho-i-buda.md","browser_download_url":"https://downloads.example/guiho-i-buda.md"}`,
		`{"name":"guiho-p-buda.md","browser_download_url":"https://downloads.example/guiho-p-buda.md"}`,
		`{"name":"guiho-s-0002-buda.zip","browser_download_url":"https://downloads.example/guiho-s-0002-buda.zip"}`,
		`{"name":"artifacts.json","browser_download_url":"https://downloads.example/artifacts.json"}`,
		`{"name":"checksums.txt","browser_download_url":"https://downloads.example/checksums.txt"}`,
	}
	return strings.Join(assets, ",\n")
}

func catalogClient() *http.Client {
	return &http.Client{Transport: releaseTransport(func(request *http.Request) (*http.Response, error) {
		body := `[{
 "tag_name":"buda/v0.0.2","name":"Buda 0.0.2","published_at":"2026-07-27T20:00:00Z",
 "html_url":"https://github.com/CGuiho/buda/releases/tag/buda/v0.0.2",
 "assets":[` + completeReleaseAssetsJSON() + `]
}]`
		return &http.Response{StatusCode: http.StatusOK, Status: "200 OK", Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}
}

func testWiki(t *testing.T) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "wiki")
	if _, err := repository.Initialize(root, repository.InitOptions{WikiID: "test-wiki"}); err != nil {
		t.Fatal(err)
	}
	return root
}

func testInstallLayout(t *testing.T) installlayout.Layout {
	t.Helper()
	layout, err := installlayout.ForHome(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return layout
}

func upgradeClient(binary []byte) *http.Client {
	digest := sha256.Sum256(binary)
	return &http.Client{Transport: releaseTransport(func(request *http.Request) (*http.Response, error) {
		body := ""
		switch request.URL.Host {
		case "api.github.com":
			body = `[{"tag_name":"buda/v0.0.2","html_url":"https://github.com/CGuiho/buda/releases/tag/buda/v0.0.2","assets":[` + completeReleaseAssetsJSON() + `]}]`
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

func TestUpgradeDryRunUsesCanonicalReleaseWithExplicitWiki(t *testing.T) {
	var output bytes.Buffer
	wiki := testWiki(t)
	deps := Dependencies{Out: &output, Err: &bytes.Buffer{}, Options: &Options{}, HTTPClient: catalogClient(), Executable: func() (string, error) { return "buda", nil }}
	root := NewRootCommand(deps, BuildInfo{Version: "0.0.1", Target: "buda-linux-amd64"})
	root.SetArgs([]string{"upgrade", "--dry-run", "--wiki", wiki})
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

func TestUninstallDryRunNeedsExplicitWikiAndDoesNotDelete(t *testing.T) {
	var output bytes.Buffer
	wiki := testWiki(t)
	removed := false
	layout, err := installlayout.ForHome(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	deps := Dependencies{
		Out: &output, Err: &bytes.Buffer{}, Options: &Options{},
		Executable:       func() (string, error) { return "buda", nil },
		RemoveExecutable: func(string) (bool, error) { removed = true; return false, nil },
		InstallLayout:    func() (installlayout.Layout, error) { return layout, nil },
	}
	root := NewRootCommand(deps, BuildInfo{Version: "0.0.1"})
	root.SetArgs([]string{"uninstall", "--dry-run", "--wiki", wiki})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if removed || !strings.Contains(output.String(), "REMOVE:") || !strings.Contains(output.String(), "PRESERVE:") || !strings.Contains(output.String(), "buda CLI home") {
		t.Fatalf("uninstall preview = %q, removed=%t", output.String(), removed)
	}
}

func TestUpgradeAndUninstallNeverScheduleResourceMaintenance(t *testing.T) {
	for _, arguments := range [][]string{{"upgrade", "--dry-run"}, {"uninstall", "--dry-run"}} {
		arguments = append(arguments, "--wiki", testWiki(t))
		scheduled := 0
		deps := Dependencies{
			Out: &bytes.Buffer{}, Err: &bytes.Buffer{}, Options: &Options{}, HTTPClient: catalogClient(),
			Executable:          func() (string, error) { return "buda", nil },
			ScheduleMaintenance: func(string, string) error { scheduled++; return nil },
		}
		if arguments[0] == "uninstall" {
			layout, layoutErr := installlayout.ForHome(t.TempDir())
			if layoutErr != nil {
				t.Fatal(layoutErr)
			}
			deps.InstallLayout = func() (installlayout.Layout, error) { return layout, nil }
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
	wiki := testWiki(t)
	layout := testInstallLayout(t)
	called := false
	deps := Dependencies{
		Out: &output, Err: &bytes.Buffer{}, Options: &Options{}, HTTPClient: catalogClient(),
		InstallLayout: func() (installlayout.Layout, error) { return layout, nil },
		UpgradeRelease: func(context.Context, upgrade.Options) (upgrade.Result, error) {
			called = true
			return upgrade.Result{}, nil
		},
	}
	root := NewRootCommand(deps, BuildInfo{Version: "0.0.2", Target: "buda-linux-amd64"})
	root.SetArgs([]string{"upgrade", "--wiki", wiki})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if called || !strings.Contains(output.String(), "already current") {
		t.Fatalf("equal upgrade output=%q called=%t", output.String(), called)
	}
}

func TestUpgradeExecutesManifestTransaction(t *testing.T) {
	var output bytes.Buffer
	wiki := testWiki(t)
	layout := testInstallLayout(t)
	called := false
	deps := Dependencies{
		Out: &output, Err: &bytes.Buffer{}, Options: &Options{}, HTTPClient: catalogClient(),
		InstallLayout: func() (installlayout.Layout, error) { return layout, nil },
		UpgradeRelease: func(_ context.Context, options upgrade.Options) (upgrade.Result, error) {
			called = true
			return upgrade.Result{
				PreviousVersion:  "0.0.1",
				InstalledVersion: "0.0.2",
				LauncherPath:     layout.Launcher,
				PayloadPath:      filepath.Join(layout.Versions, "0.0.2", "buda"),
			}, nil
		},
		ReconcileInstalled: func(string, string) error { return nil },
	}
	root := NewRootCommand(deps, BuildInfo{Version: "0.0.1", Target: "buda-linux-amd64"})
	root.SetArgs([]string{"upgrade", "--wiki", wiki})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("UpgradeRelease was not called")
	}
	for _, expected := range []string{"Buda upgraded from 0.0.1 to 0.0.2", "Launcher:", "Payload:", "Post-upgrade explicit-wiki init: verified"} {
		if !strings.Contains(output.String(), expected) {
			t.Fatalf("upgrade text missing %q:\n%s", expected, output.String())
		}
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
