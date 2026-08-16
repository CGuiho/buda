package selfmanage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const maxDownloadBytes int64 = 256 << 20

type ReplaceFunc func(executable, candidate, backup, targetVersion, checksum, wiki string, verify VerifyFunc) (bool, error)
type VerifyFunc func(path, version string) error

type UpgradeOptions struct {
	Executable       string
	TargetVersion    string
	DownloadURL      string
	ExpectedChecksum string
	Wiki             string
	Client           Doer
	Replace          ReplaceFunc
	Verify           VerifyFunc
	Progress         func(DownloadProgress)
}

type DownloadProgress struct {
	Bytes   int64   `json:"bytes"`
	Total   int64   `json:"total,omitempty"`
	Percent float64 `json:"percent,omitempty"`
}

type UpgradeResult struct {
	CurrentVersion string `json:"current_version"`
	TargetVersion  string `json:"target_version"`
	Asset          string `json:"asset"`
	Executable     string `json:"executable"`
	Backup         string `json:"backup,omitempty"`
	Recovery       string `json:"recovery,omitempty"`
}

func AssetName(buildTarget string) (string, error) {
	target := strings.TrimSuffix(strings.TrimSpace(buildTarget), ".exe")
	if target == "" || target == "development" {
		target = "buda-" + runtime.GOOS + "-" + runtime.GOARCH
		if runtime.GOARCH == "arm" {
			target += "v7"
		}
	}
	valid := map[string]bool{
		"buda-linux-amd64": true, "buda-linux-arm64": true,
		"buda-linux-armv7": true, "buda-linux-armv6": true,
		"buda-darwin-amd64": true, "buda-darwin-arm64": true,
		"buda-windows-amd64": true, "buda-windows-arm64": true,
	}
	if !valid[target] {
		return "", fmt.Errorf("unsupported embedded build target %q", buildTarget)
	}
	if strings.HasPrefix(target, "buda-windows-") {
		return target + ".exe", nil
	}
	return target, nil
}

// LauncherAssetName returns the stable launcher paired with a released
// payload. Keeping this derivation beside AssetName prevents an upgrade from
// accepting a payload-only release that cannot repair a fresh installation.
func LauncherAssetName(buildTarget string) (string, error) {
	payload, err := AssetName(buildTarget)
	if err != nil {
		return "", err
	}
	target := strings.TrimSuffix(payload, ".exe")
	if !strings.HasPrefix(target, "buda-") {
		return "", fmt.Errorf("unsupported embedded build target %q", buildTarget)
	}
	launcher := "buda-launcher-" + strings.TrimPrefix(target, "buda-")
	if strings.HasSuffix(payload, ".exe") {
		launcher += ".exe"
	}
	return launcher, nil
}

func FetchChecksum(ctx context.Context, client Doer, manifestURL, assetName string) (string, error) {
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, manifestURL, nil)
	if err != nil {
		return "", fmt.Errorf("create checksum request: %w", err)
	}
	response, err := client.Do(request)
	if err != nil {
		return "", fmt.Errorf("download checksums: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download checksums returned %s", response.Status)
	}
	content, err := io.ReadAll(io.LimitReader(response.Body, (1<<20)+1))
	if err != nil {
		return "", fmt.Errorf("read checksums: %w", err)
	}
	if len(content) > 1<<20 {
		return "", fmt.Errorf("checksums.txt exceeds 1048576 bytes")
	}
	for _, line := range strings.Split(string(content), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 || strings.TrimPrefix(fields[1], "*") != assetName {
			continue
		}
		checksum := strings.ToLower(fields[0])
		if len(checksum) == sha256.Size*2 {
			if _, err := hex.DecodeString(checksum); err == nil {
				return checksum, nil
			}
		}
	}
	return "", fmt.Errorf("checksums.txt does not contain a valid checksum for %s", assetName)
}

func Upgrade(ctx context.Context, options UpgradeOptions) (UpgradeResult, error) {
	result := UpgradeResult{TargetVersion: strings.TrimPrefix(options.TargetVersion, "v")}
	if !strictSemver(result.TargetVersion) {
		return result, fmt.Errorf("target version must be an exact semantic version")
	}
	parsed, err := url.ParseRequestURI(options.DownloadURL)
	if err != nil || parsed.Scheme != "https" && parsed.Scheme != "http" {
		return result, fmt.Errorf("valid HTTP(S) download URL is required")
	}
	expected := strings.ToLower(strings.TrimSpace(options.ExpectedChecksum))
	if len(expected) != sha256.Size*2 {
		return result, fmt.Errorf("expected SHA-256 checksum is required")
	}
	if _, err := hex.DecodeString(expected); err != nil {
		return result, fmt.Errorf("expected SHA-256 checksum is invalid")
	}
	executable, err := filepath.Abs(options.Executable)
	if err != nil {
		return result, fmt.Errorf("resolve executable: %w", err)
	}
	if resolved, resolveErr := filepath.EvalSymlinks(executable); resolveErr == nil {
		executable = resolved
	}
	result.Executable = executable
	result.Backup = executable + ".old"
	result.Recovery = fmt.Sprintf("restore %q to %q", result.Backup, executable)
	candidate, checksum, err := downloadCandidate(ctx, options.Client, options.DownloadURL, filepath.Dir(executable), options.Progress)
	if err != nil {
		return result, err
	}
	removeCandidate := true
	defer func() {
		if removeCandidate {
			_ = os.Remove(candidate)
		}
	}()
	if checksum != expected {
		return result, fmt.Errorf("checksum mismatch: expected %s, got %s", expected, checksum)
	}
	rotated, err := rotateBackup(result.Backup)
	if err != nil {
		return result, err
	}
	verify := options.Verify
	if verify == nil {
		verify = VerifyExecutable
	}
	replace := options.Replace
	if replace == nil {
		replace = replaceExecutable
	}
	deferred, err := replace(executable, candidate, result.Backup, result.TargetVersion, expected, options.Wiki, verify)
	if err != nil {
		if rotated {
			if restoreErr := restoreRotatedBackup(result.Backup); restoreErr != nil {
				return result, fmt.Errorf("%w; restoring prior rollback backup also failed: %v", err, restoreErr)
			}
		}
		return result, err
	}
	if deferred {
		if rotated {
			if restoreErr := restoreRotatedBackup(result.Backup); restoreErr != nil {
				return result, fmt.Errorf("replacement did not complete synchronously; restoring prior rollback backup also failed: %v", restoreErr)
			}
		}
		return result, errors.New("replacement did not complete synchronously; stable launcher installation is required")
	}
	removeCandidate = false
	result.Recovery = ""
	return result, nil
}

func downloadCandidate(ctx context.Context, client Doer, downloadURL, directory string, progress func(DownloadProgress)) (string, string, error) {
	if client == nil {
		client = &http.Client{Timeout: 2 * time.Minute}
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return "", "", fmt.Errorf("create binary request: %w", err)
	}
	response, err := client.Do(request)
	if err != nil {
		return "", "", fmt.Errorf("download update binary: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("download update binary returned %s", response.Status)
	}
	if response.ContentLength > maxDownloadBytes {
		return "", "", fmt.Errorf("update binary exceeds %d bytes", maxDownloadBytes)
	}
	file, err := os.CreateTemp(directory, ".buda-upgrade-*")
	if err != nil {
		return "", "", fmt.Errorf("create staged update: %w", err)
	}
	path := file.Name()
	remove := true
	defer func() {
		_ = file.Close()
		if remove {
			_ = os.Remove(path)
		}
	}()
	hash := sha256.New()
	limited := io.LimitReader(response.Body, maxDownloadBytes+1)
	buffer := make([]byte, 256<<10)
	var written int64
	var lastPercent int
	for {
		count, readErr := limited.Read(buffer)
		if count > 0 {
			if _, err := file.Write(buffer[:count]); err != nil {
				return "", "", fmt.Errorf("write staged update: %w", err)
			}
			_, _ = hash.Write(buffer[:count])
			written += int64(count)
			if progress != nil {
				event := DownloadProgress{Bytes: written, Total: response.ContentLength}
				if response.ContentLength > 0 {
					event.Percent = float64(written) * 100 / float64(response.ContentLength)
				}
				percent := int(event.Percent)
				if response.ContentLength <= 0 || percent >= lastPercent+5 || written == response.ContentLength {
					progress(event)
					lastPercent = percent
				}
			}
		}
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			return "", "", fmt.Errorf("read staged update: %w", readErr)
		}
	}
	if written > maxDownloadBytes {
		return "", "", fmt.Errorf("update binary exceeds %d bytes", maxDownloadBytes)
	}
	if err := file.Sync(); err != nil {
		return "", "", fmt.Errorf("sync staged update: %w", err)
	}
	if err := file.Close(); err != nil {
		return "", "", fmt.Errorf("close staged update: %w", err)
	}
	if runtime.GOOS != "windows" {
		if err := os.Chmod(path, 0o755); err != nil {
			return "", "", fmt.Errorf("make staged update executable: %w", err)
		}
	}
	remove = false
	return path, hex.EncodeToString(hash.Sum(nil)), nil
}

func rotateBackup(backup string) (bool, error) {
	if _, err := os.Stat(backup); errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("inspect prior upgrade backup: %w", err)
	}
	previous := backup + ".previous"
	if err := os.Remove(previous); err != nil && !errors.Is(err, os.ErrNotExist) {
		return false, fmt.Errorf("remove stale upgrade backup rotation: %w", err)
	}
	if err := os.Rename(backup, previous); err != nil {
		return false, fmt.Errorf("rotate prior upgrade backup: %w", err)
	}
	return true, nil
}

func restoreRotatedBackup(backup string) error {
	if _, err := os.Stat(backup); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return os.Rename(backup+".previous", backup)
}

func rollbackFiles(executable string) error {
	backup := executable + ".old"
	if info, err := os.Stat(backup); err != nil || info.IsDir() {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("no backup executable found at %s", backup)
		}
		if err != nil {
			return fmt.Errorf("inspect rollback backup: %w", err)
		}
		return fmt.Errorf("rollback backup is not a file: %s", backup)
	}
	failed := fmt.Sprintf("%s.failed-%d", executable, time.Now().UnixNano())
	if err := os.Rename(executable, failed); err != nil {
		return fmt.Errorf("stage current executable for rollback: %w", err)
	}
	if err := os.Rename(backup, executable); err != nil {
		if restoreErr := os.Rename(failed, executable); restoreErr != nil {
			return fmt.Errorf("restore backup: %w; restoring current executable also failed: %v", err, restoreErr)
		}
		return fmt.Errorf("restore backup: %w", err)
	}
	_ = os.Remove(failed)
	return nil
}

func VerifyExecutable(path, version string) error {
	command := exec.Command(path, "--version")
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("verify replacement executable: %w (%s)", err, strings.TrimSpace(string(output)))
	}
	expected := strings.TrimPrefix(version, "v")
	if observed := strings.TrimSpace(string(output)); observed != expected {
		return fmt.Errorf("verify replacement executable: expected %q, got %q", expected, observed)
	}
	return nil
}
