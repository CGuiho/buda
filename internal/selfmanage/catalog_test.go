package selfmanage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCatalogRetainsCanonicalSortedReleases(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Query().Get("page") != "1" {
			t.Fatalf("unexpected page: %s", request.URL.String())
		}
		fmt.Fprint(writer, `[
 {"tag_name":"buda/v0.0.2-rc.1","prerelease":true,"assets":[]},
 {"tag_name":"other/v9.0.0","assets":[]},
 {"tag_name":"buda/v0.0.1","assets":[]},
 {"tag_name":"buda/v0.0.2","assets":[]},
 {"tag_name":"buda/v01.2.3","assets":[]},
 {"tag_name":"buda/v9.9.9","draft":true,"assets":[]}
]`)
	}))
	defer server.Close()
	releases, err := (Catalog{Client: server.Client(), BaseURL: server.URL, Repository: "owner/repo"}).Releases(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"0.0.2", "0.0.2-rc.1", "0.0.1"}
	if len(releases) != len(want) {
		t.Fatalf("releases = %#v", releases)
	}
	for index, version := range want {
		if releases[index].Version != version {
			t.Fatalf("release %d = %q, want %q", index, releases[index].Version, version)
		}
	}
}

func TestVersionComparisonAndTargetSelection(t *testing.T) {
	for _, test := range []struct {
		left, right string
		sign        int
	}{
		{"0.0.2", "0.0.1", 1}, {"1.0.0-rc.1", "1.0.0", -1}, {"1.0.0-rc.10", "1.0.0-rc.2", 1},
	} {
		value := CompareVersions(test.left, test.right)
		if value*test.sign <= 0 {
			t.Fatalf("CompareVersions(%q, %q) = %d", test.left, test.right, value)
		}
	}
	for target, want := range map[string]string{
		"buda-linux-armv6": "buda-linux-armv6", "buda-windows-arm64": "buda-windows-arm64.exe",
	} {
		got, err := AssetName(target)
		if err != nil || got != want {
			t.Fatalf("AssetName(%q) = %q, %v", target, got, err)
		}
	}
	if _, err := AssetName("buda-linux-ppc64"); err == nil {
		t.Fatal("unsupported embedded target was accepted")
	}
}

func TestUpgradeVerifiesChecksumBeforeInjectedReplacement(t *testing.T) {
	content := []byte("new-buda-binary")
	digest := sha256.Sum256(content)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) { _, _ = writer.Write(content) }))
	defer server.Close()
	directory := t.TempDir()
	executable := filepath.Join(directory, "buda")
	if err := os.WriteFile(executable, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	called := false
	result, err := Upgrade(context.Background(), UpgradeOptions{
		Executable: executable, TargetVersion: "0.0.2", DownloadURL: server.URL,
		ExpectedChecksum: hex.EncodeToString(digest[:]), Client: server.Client(),
		Replace: func(gotExecutable, candidate, backup, version, checksum, wiki string, _ VerifyFunc) (bool, error) {
			called = true
			if gotExecutable != executable || backup != executable+".old" || version != "0.0.2" || wiki != "" {
				t.Fatalf("replacement inputs: %q %q %q %q", gotExecutable, backup, version, wiki)
			}
			staged, readErr := os.ReadFile(candidate)
			if readErr != nil || string(staged) != string(content) || checksum != hex.EncodeToString(digest[:]) {
				t.Fatalf("staged candidate was not checksum verified: %q %v", staged, readErr)
			}
			return false, nil
		},
	})
	if err != nil || !called || result.TargetVersion != "0.0.2" {
		t.Fatalf("upgrade = %#v, %v, called=%t", result, err, called)
	}
}

func TestUpgradeRejectsChecksumMismatchWithoutReplacement(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) { fmt.Fprint(writer, "tampered") }))
	defer server.Close()
	directory := t.TempDir()
	executable := filepath.Join(directory, "buda")
	if err := os.WriteFile(executable, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	_, err := Upgrade(context.Background(), UpgradeOptions{
		Executable: executable, TargetVersion: "0.0.2", DownloadURL: server.URL,
		ExpectedChecksum: strings.Repeat("0", 64), Client: server.Client(),
		Replace: func(string, string, string, string, string, string, VerifyFunc) (bool, error) {
			t.Fatal("replacement ran after a checksum mismatch")
			return false, nil
		},
	})
	if err == nil || !strings.Contains(err.Error(), "checksum mismatch") {
		t.Fatalf("checksum error = %v", err)
	}
}

func TestFetchChecksumRequiresExactAssetEntry(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		fmt.Fprintln(writer, strings.Repeat("a", 64), "*buda-linux-amd64")
	}))
	defer server.Close()
	checksum, err := FetchChecksum(context.Background(), server.Client(), server.URL, "buda-linux-amd64")
	if err != nil || checksum != strings.Repeat("a", 64) {
		t.Fatalf("checksum = %q, %v", checksum, err)
	}
}

func TestUpgradeRotatesPriorBackupAndKeepsScheduledCandidate(t *testing.T) {
	content := []byte("next")
	digest := sha256.Sum256(content)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) { _, _ = writer.Write(content) }))
	defer server.Close()
	directory := t.TempDir()
	executable := filepath.Join(directory, "buda")
	if err := os.WriteFile(executable, []byte("current"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(executable+".old", []byte("previous"), 0o755); err != nil {
		t.Fatal(err)
	}
	var candidate string
	_, err := Upgrade(context.Background(), UpgradeOptions{
		Executable: executable, TargetVersion: "0.0.3", DownloadURL: server.URL,
		ExpectedChecksum: hex.EncodeToString(digest[:]), Client: server.Client(),
		Replace: func(_, gotCandidate, _, _, _, _ string, _ VerifyFunc) (bool, error) {
			candidate = gotCandidate
			return true, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(candidate)
	if _, err := os.Stat(candidate); err != nil {
		t.Fatalf("scheduled candidate was removed before helper takeover: %v", err)
	}
	rotated, err := os.ReadFile(executable + ".old.previous")
	if err != nil || string(rotated) != "previous" {
		t.Fatalf("prior backup was not safely rotated: %q %v", rotated, err)
	}
}

func TestRollbackFilesRestoresBackup(t *testing.T) {
	directory := t.TempDir()
	executable := filepath.Join(directory, "buda")
	if err := os.WriteFile(executable, []byte("new"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(executable+".old", []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := rollbackFiles(executable); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(executable)
	if err != nil || string(content) != "old" {
		t.Fatalf("rollback content=%q err=%v", content, err)
	}
	if _, err := os.Stat(executable + ".old"); !os.IsNotExist(err) {
		t.Fatalf("rollback backup was not consumed: %v", err)
	}
}

func TestUpgradeProgressIsInjectedAndBounded(t *testing.T) {
	content := []byte(strings.Repeat("x", 1024))
	digest := sha256.Sum256(content)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Length", "1024")
		_, _ = writer.Write(content)
	}))
	defer server.Close()
	directory := t.TempDir()
	executable := filepath.Join(directory, "buda")
	if err := os.WriteFile(executable, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	var events []DownloadProgress
	_, err := Upgrade(context.Background(), UpgradeOptions{
		Executable: executable, TargetVersion: "0.0.2", DownloadURL: server.URL,
		ExpectedChecksum: hex.EncodeToString(digest[:]), Client: server.Client(),
		Progress: func(event DownloadProgress) { events = append(events, event) },
		Replace:  func(string, string, string, string, string, string, VerifyFunc) (bool, error) { return false, nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) == 0 || events[len(events)-1].Percent != 100 || events[len(events)-1].Bytes != 1024 {
		t.Fatalf("progress events=%#v", events)
	}
}
