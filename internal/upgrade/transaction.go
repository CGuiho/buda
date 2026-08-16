// Package upgrade implements the synchronous manifest-based release
// transaction used by the installed stable launcher.
package upgrade

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/CGuiho/buda/internal/artifact"
	"github.com/CGuiho/buda/internal/installlayout"
	"github.com/CGuiho/buda/internal/launcher"
	"github.com/CGuiho/buda/internal/lifecycle"
	"github.com/CGuiho/buda/internal/releasecatalog"
	"github.com/CGuiho/buda/internal/selfmanage"
)

type Options struct {
	Layout         installlayout.Layout
	Selection      releasecatalog.Result
	Client         selfmanage.Doer
	CurrentVersion string
	Target         string
	OS             string
	Arch           string
	CurrentPID     int
}

type Result struct {
	PreviousVersion  string   `json:"previous_version"`
	InstalledVersion string   `json:"installed_version"`
	LauncherPath     string   `json:"launcher_path"`
	PayloadPath      string   `json:"payload_path"`
	Artifacts        []string `json:"artifacts"`
}

// Execute downloads, validates, stages, and activates one complete release.
// The running payload is never replaced. The stable launcher remains usable
// while the new immutable version directory and current pointer are changed.
func Execute(ctx context.Context, options Options) (result Result, err error) {
	if err = options.Layout.Validate(); err != nil {
		return Result{}, err
	}
	selection := options.Selection
	if selection.Release.Version == "" || selection.Binary.Name == "" || selection.Launcher.Name == "" {
		return Result{}, errors.New("complete release selection is required")
	}
	if !releasecatalog.IsSemver(selection.Release.Version) {
		return Result{}, fmt.Errorf("release version %q is not valid SemVer", selection.Release.Version)
	}
	expectedLauncher, launcherErr := selfmanage.LauncherAssetName(optionsTarget(options))
	if launcherErr != nil {
		return Result{}, launcherErr
	}
	if selection.Launcher.Name != expectedLauncher {
		return Result{}, fmt.Errorf("release launcher %q does not match native target %q", selection.Launcher.Name, expectedLauncher)
	}
	client := options.Client
	if client == nil {
		client = &http.Client{}
	}
	op, err := options.Layout.Operation("upgrade")
	if err != nil {
		return Result{}, err
	}
	defer os.RemoveAll(op)
	lock, err := lifecycle.AcquireLock(filepath.Join(options.Layout.State, "upgrade.lock"))
	if err != nil {
		return Result{}, err
	}
	defer func() {
		if releaseErr := lock.Release(); err == nil && releaseErr != nil {
			err = releaseErr
		}
	}()

	journalPath := filepath.Join(options.Layout.State, "transaction.json")
	if err = recoverInterrupted(options.Layout, journalPath); err != nil {
		return Result{}, err
	}
	journal := lifecycle.Journal{Operation: "upgrade", Phase: lifecycle.PhasePlanned, Token: lock.Token, Version: selection.Release.Version, PreviousVersion: options.CurrentVersion}
	if err = lifecycle.SaveJournal(journalPath, journal); err != nil {
		return Result{}, err
	}
	currentBytes, currentExists, currentErr := snapshotFile(options.Layout.Current)
	if currentErr != nil {
		return Result{}, currentErr
	}
	installedBytes, installedExists, installedErr := snapshotFile(options.Layout.InstalledManifest)
	if installedErr != nil {
		return Result{}, installedErr
	}
	launcherBytes, launcherExists, launcherErr := snapshotFile(options.Layout.Launcher)
	if launcherErr != nil {
		return Result{}, launcherErr
	}
	previous := readPrevious(options.Layout.Current)
	previousPath := filepath.Join(options.Layout.Versions, filepath.FromSlash(previous.Active))
	if previous.Active != "" && !withinPath(options.Layout.Versions, previousPath) {
		return Result{}, errors.New("existing active pointer escapes the Buda versions directory")
	}

	var versionDir string
	var versionBackup string
	var projectionSnapshots []projectionSnapshot
	versionChanged := false
	launcherChanged := false
	mutated := false
	defer func() {
		if err == nil || !mutated {
			return
		}
		_ = lifecycle.SaveJournal(journalPath, lifecycle.Journal{Operation: "upgrade", Phase: lifecycle.PhaseRollingBack, Token: lock.Token, Version: selection.Release.Version, PreviousVersion: previous.ActiveVersion})
		var failures []error
		if len(projectionSnapshots) > 0 {
			if projErr := rollbackProjections(projectionSnapshots); projErr != nil {
				failures = append(failures, projErr)
			}
		}
		if (versionChanged || versionBackup != "") && versionDir != "" {
			if versionChanged {
				if removeErr := lifecycle.SafeRemove(versionDir, options.Layout.Versions); removeErr != nil && !os.IsNotExist(removeErr) {
					failures = append(failures, removeErr)
				}
			}
			if versionBackup != "" {
				if restoreErr := os.Rename(versionBackup, versionDir); restoreErr != nil {
					failures = append(failures, restoreErr)
				}
			}
		}
		if launcherChanged {
			if restoreErr := restoreSnapshot(options.Layout.Launcher, launcherBytes, launcherExists); restoreErr != nil {
				failures = append(failures, restoreErr)
			}
		}
		if restoreErr := restoreSnapshot(options.Layout.Current, currentBytes, currentExists); restoreErr != nil {
			failures = append(failures, restoreErr)
		}
		if restoreErr := restoreSnapshot(options.Layout.InstalledManifest, installedBytes, installedExists); restoreErr != nil {
			failures = append(failures, restoreErr)
		}
		if len(failures) == 0 {
			_ = lifecycle.SaveJournal(journalPath, lifecycle.Journal{Operation: "upgrade", Phase: lifecycle.PhaseRolledBack, Token: lock.Token, Version: selection.Release.Version, PreviousVersion: previous.ActiveVersion})
		} else {
			err = errors.Join(err, fmt.Errorf("upgrade rollback failed: %v", failures))
		}
	}()

	files, err := fetchRelease(ctx, client, selection.Release, op)
	if err != nil {
		return Result{}, err
	}
	if err = verifyChecksums(files, filepath.Join(op, "checksums.txt")); err != nil {
		return Result{}, err
	}
	manifest, err := artifact.Load(filepath.Join(op, "artifacts.json"))
	if err != nil {
		return Result{}, fmt.Errorf("validate release artifacts.json: %w", err)
	}
	if manifest.Version != selection.Release.Version {
		return Result{}, fmt.Errorf("release manifest version %q does not match release %q", manifest.Version, selection.Release.Version)
	}
	if err = validateManifestAssets(manifest, files); err != nil {
		return Result{}, err
	}
	if err = validateArchive(filepath.Join(op, "guiho-s-0002-buda.zip"), manifest); err != nil {
		return Result{}, err
	}
	if err = lifecycle.SaveJournal(journalPath, lifecycle.Journal{Operation: "upgrade", Phase: lifecycle.PhaseStaged, Token: lock.Token, Version: selection.Release.Version, PreviousVersion: previous.ActiveVersion}); err != nil {
		return Result{}, err
	}

	payload := filepath.Join(op, selection.Binary.Name)
	if err = ensureExecutable(payload); err != nil {
		return Result{}, err
	}
	if observed, versionErr := versionOf(payload); versionErr != nil {
		return Result{}, versionErr
	} else if observed != selection.Release.Version {
		return Result{}, fmt.Errorf("candidate raw version %q does not match %q", observed, selection.Release.Version)
	}
	if selfTestErr := selfTest(payload); selfTestErr != nil {
		return Result{}, selfTestErr
	}
	if err = lifecycle.SaveJournal(journalPath, lifecycle.Journal{Operation: "upgrade", Phase: lifecycle.PhaseCandidateVerified, Token: lock.Token, Version: selection.Release.Version, PreviousVersion: previous.ActiveVersion}); err != nil {
		return Result{}, err
	}

	versionDir = filepath.Join(options.Layout.Versions, selection.Release.Version)
	if previous.ActiveVersion == selection.Release.Version {
		return Result{}, fmt.Errorf("release %s is already selected by the active pointer", selection.Release.Version)
	}
	if info, statErr := os.Lstat(versionDir); statErr == nil {
		if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			return Result{}, fmt.Errorf("installed version path is not a regular directory: %s", versionDir)
		}
		versionBackup = filepath.Join(op, "previous-version")
		if err = os.Rename(versionDir, versionBackup); err != nil {
			return Result{}, fmt.Errorf("stage existing immutable version directory: %w", err)
		}
		mutated = true
	} else if !os.IsNotExist(statErr) {
		return Result{}, statErr
	}
	candidateDir := filepath.Join(op, "version")
	if err = stageVersion(candidateDir, payload, files); err != nil {
		return Result{}, err
	}
	if err = os.MkdirAll(options.Layout.Versions, 0o755); err != nil {
		return Result{}, err
	}
	if err = os.Rename(candidateDir, versionDir); err != nil {
		return Result{}, fmt.Errorf("activate immutable version directory: %w", err)
	}
	versionChanged = true
	mutated = true

	var previousManifest artifact.Manifest
	if installedExists && len(installedBytes) > 0 {
		previousManifest, _ = artifact.Decode(bytes.NewReader(installedBytes))
	}

	projectionSnapshots, err = snapshotProjections(op, options.Layout.Home, manifest, previousManifest)
	if err != nil {
		return Result{}, fmt.Errorf("snapshot projections: %w", err)
	}
	if err = lifecycle.SaveJournal(journalPath, lifecycle.Journal{Operation: "upgrade", Phase: lifecycle.PhaseProjectionsSnapshotted, Token: lock.Token, Version: selection.Release.Version, PreviousVersion: previous.ActiveVersion}); err != nil {
		return Result{}, err
	}

	if err = applyProjections(op, options.Layout.Home, versionDir, manifest, previousManifest); err != nil {
		return Result{}, fmt.Errorf("apply manifest projections: %w", err)
	}
	if err = lifecycle.SaveJournal(journalPath, lifecycle.Journal{Operation: "upgrade", Phase: lifecycle.PhaseArtifactsReplaced, Token: lock.Token, Version: selection.Release.Version, PreviousVersion: previous.ActiveVersion}); err != nil {
		return Result{}, err
	}
	if previous.Active != "" {
		oldPayload := filepath.Join(options.Layout.Versions, filepath.FromSlash(previous.Active))
		if err = lifecycle.TerminateOtherInstances(options.Layout.Instances, options.CurrentPID, oldPayload); err != nil {
			return Result{}, fmt.Errorf("terminate other verified Buda payload instances: %w", err)
		}
	}
	if !launcherExists {
		launcherSource, ok := files[selection.Launcher.Name]
		if !ok {
			return Result{}, fmt.Errorf("selected launcher asset %q was not downloaded", selection.Launcher.Name)
		}
		launcherContent, readErr := os.ReadFile(launcherSource)
		if readErr != nil {
			return Result{}, fmt.Errorf("read selected stable launcher: %w", readErr)
		}
		if err = lifecycle.AtomicWrite(options.Layout.Launcher, launcherContent, 0o755); err != nil {
			return Result{}, fmt.Errorf("install stable launcher: %w", err)
		}
		launcherChanged = true
		mutated = true
	}

	pointer := launcher.Pointer{Schema: 1, Active: filepath.ToSlash(filepath.Join(selection.Release.Version, payloadNameForTarget())), Previous: previous.Active, ActiveVersion: selection.Release.Version, PreviousVersion: previous.ActiveVersion}
	data, err := json.Marshal(pointer)
	if err != nil {
		return Result{}, err
	}
	if err = lifecycle.AtomicWrite(options.Layout.Current, append(data, '\n'), 0o644); err != nil {
		return Result{}, err
	}
	mutated = true
	if err = lifecycle.AtomicWrite(options.Layout.InstalledManifest, mustEncodeManifest(manifest), 0o644); err != nil {
		return Result{}, err
	}
	if err = lifecycle.SaveJournal(journalPath, lifecycle.Journal{Operation: "upgrade", Phase: lifecycle.PhaseActivated, Token: lock.Token, Version: selection.Release.Version, PreviousVersion: previous.ActiveVersion}); err != nil {
		return Result{}, err
	}
	if verified, verifyErr := launcher.ReadPointer(options.Layout.Current); verifyErr != nil || verified.ActiveVersion != selection.Release.Version {
		if verifyErr != nil {
			return Result{}, fmt.Errorf("verify activated pointer: %w", verifyErr)
		}
		return Result{}, fmt.Errorf("activated pointer version %q does not match %q", verified.ActiveVersion, selection.Release.Version)
	}
	if err = verifyInstalledRelease(options.Layout, manifest, versionDir, selection.Release.Version); err != nil {
		return Result{}, err
	}
	if err = verifyLauncherVersion(options.Layout.Launcher, selection.Release.Version); err != nil {
		return Result{}, err
	}
	if err = lifecycle.SaveJournal(journalPath, lifecycle.Journal{Operation: "upgrade", Phase: lifecycle.PhaseVerified, Token: lock.Token, Version: selection.Release.Version, PreviousVersion: previous.ActiveVersion}); err != nil {
		return Result{}, err
	}
	if err = lifecycle.SaveJournal(journalPath, lifecycle.Journal{Operation: "upgrade", Phase: lifecycle.PhaseComplete, Token: lock.Token, Version: selection.Release.Version, PreviousVersion: previous.ActiveVersion}); err != nil {
		return Result{}, err
	}
	if err = os.Remove(journalPath); err != nil && !os.IsNotExist(err) {
		return Result{}, err
	}
	result = Result{PreviousVersion: previous.ActiveVersion, InstalledVersion: selection.Release.Version, LauncherPath: options.Layout.Launcher, PayloadPath: filepath.Join(versionDir, payloadNameForTarget())}
	for _, entry := range manifest.Artifacts {
		result.Artifacts = append(result.Artifacts, entry.ID)
	}
	sort.Strings(result.Artifacts)
	return result, nil
}

func optionsTarget(options Options) string {
	if strings.TrimSpace(options.Target) != "" && options.Target != "development" {
		return options.Target
	}
	if strings.TrimSpace(options.OS) != "" && strings.TrimSpace(options.Arch) != "" {
		target := "buda-" + strings.ToLower(strings.TrimSpace(options.OS)) + "-" + strings.ToLower(strings.TrimSpace(options.Arch))
		if strings.EqualFold(options.Arch, "arm") {
			target += "v7"
		}
		return target
	}
	return "development"
}

func payloadNameForTarget() string {
	if runtime.GOOS == "windows" {
		return "buda.exe"
	}
	return "buda"
}

func stageVersion(destination, payload string, files map[string]string) error {
	if err := os.MkdirAll(filepath.Join(destination, "artifacts"), 0o755); err != nil {
		return err
	}
	payloadDestination := filepath.Join(destination, payloadNameForTarget())
	if err := copyFile(payload, payloadDestination, 0o755); err != nil {
		return err
	}
	for name, source := range files {
		if err := copyFile(source, filepath.Join(destination, "artifacts", filepath.Base(name)), 0o644); err != nil {
			return fmt.Errorf("copy release artifact %s: %w", name, err)
		}
	}
	return nil
}

func verifyInstalledRelease(layout installlayout.Layout, manifest artifact.Manifest, versionDir, version string) error {
	payload := filepath.Join(versionDir, payloadNameForTarget())
	if observed, err := versionOf(payload); err != nil {
		return fmt.Errorf("verify activated payload: %w", err)
	} else if observed != version {
		return fmt.Errorf("activated payload version %q does not match %q", observed, version)
	}
	if err := selfTest(payload); err != nil {
		return fmt.Errorf("verify activated payload self-test: %w", err)
	}
	for _, entry := range manifest.Artifacts {
		if strings.Trim(entry.SHA256, "0") == "" {
			continue
		}
		path := filepath.Join(layout.CLIHome, filepath.FromSlash(entry.InstalledPath))
		actual, err := digest(path)
		if err != nil {
			return fmt.Errorf("verify installed artifact %s: %w", entry.ID, err)
		}
		if !strings.EqualFold(actual, entry.SHA256) {
			return fmt.Errorf("installed artifact %s checksum mismatch", entry.ID)
		}
	}
	return nil
}

func verifyLauncherVersion(path, expected string) error {
	if strings.TrimSpace(path) == "" {
		return errors.New("stable launcher path is required for post-activation verification")
	}
	command := exec.Command(path, "--version")
	output, err := command.Output()
	if err != nil {
		return fmt.Errorf("run stable launcher after activation: %w", err)
	}
	if observed := strings.TrimSpace(string(output)); observed != expected {
		return fmt.Errorf("stable launcher reported %q after activation; expected %q", observed, expected)
	}
	return nil
}

func recoverInterrupted(layout installlayout.Layout, journalPath string) error {
	journal, err := lifecycle.LoadJournal(journalPath)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("inspect interrupted upgrade journal: %w", err)
	}
	if journal.Phase == lifecycle.PhaseComplete || journal.Phase == lifecycle.PhaseRolledBack {
		return os.Remove(journalPath)
	}
	pointer, pointerErr := launcher.ReadPointer(layout.Current)
	if pointerErr != nil {
		return fmt.Errorf("recover interrupted upgrade: active pointer is unavailable: %w", pointerErr)
	}
	if pointer.ActiveVersion == journal.Version {
		payload := filepath.Join(layout.Versions, filepath.FromSlash(pointer.Active))
		if _, statErr := os.Stat(payload); statErr != nil {
			return fmt.Errorf("recover interrupted upgrade: activated payload is missing: %w", statErr)
		}
		return os.Remove(journalPath)
	}
	if journal.PreviousVersion == "" || pointer.ActiveVersion == journal.PreviousVersion {
		return os.Remove(journalPath)
	}
	return fmt.Errorf("recover interrupted upgrade: journal target %s and active pointer %s disagree", journal.Version, pointer.ActiveVersion)
}

func validateManifestAssets(manifest artifact.Manifest, files map[string]string) error {
	for _, entry := range manifest.Artifacts {
		path, ok := files[entry.Path]
		if !ok {
			return fmt.Errorf("release manifest artifact %s is not published", entry.Path)
		}
		if strings.Trim(entry.SHA256, "0") != "" {
			actual, err := digest(path)
			if err != nil {
				return fmt.Errorf("digest release artifact %s: %w", entry.ID, err)
			}
			if !strings.EqualFold(actual, entry.SHA256) {
				return fmt.Errorf("manifest checksum mismatch for %s", entry.ID)
			}
		}
	}
	for name := range files {
		if name == "checksums.txt" {
			continue
		}
		found := false
		for _, entry := range manifest.Artifacts {
			if entry.Path == name {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("published artifact %s is not declared by artifacts.json", name)
		}
	}
	return nil
}

func validateArchive(path string, manifest artifact.Manifest) error {
	archive, err := zip.OpenReader(path)
	if err != nil {
		return fmt.Errorf("open skill archive: %w", err)
	}
	defer archive.Close()
	if len(archive.File) == 0 {
		return errors.New("skill archive is empty")
	}
	actualMembers := make([]string, 0, len(archive.File))
	for _, entry := range archive.File {
		name := filepath.Clean(filepath.FromSlash(entry.Name))
		if filepath.IsAbs(name) || name == ".." || strings.HasPrefix(name, ".."+string(filepath.Separator)) {
			return fmt.Errorf("skill archive member escapes its root: %s", entry.Name)
		}
		actualMembers = append(actualMembers, filepath.ToSlash(name))
	}
	sort.Strings(actualMembers)
	for _, declared := range manifest.Artifacts {
		if declared.Path != "guiho-s-0002-buda.zip" {
			continue
		}
		members := append([]string(nil), declared.ArchiveMembers...)
		sort.Strings(members)
		if len(members) == 0 {
			return errors.New("skill archive manifest members are empty")
		}
		if len(members) != len(actualMembers) {
			return fmt.Errorf("skill archive member count %d does not match manifest %d", len(actualMembers), len(members))
		}
		for index := range members {
			if members[index] != actualMembers[index] {
				return fmt.Errorf("skill archive member %q is not declared", actualMembers[index])
			}
		}
	}
	return nil
}

func fetchRelease(ctx context.Context, client selfmanage.Doer, release selfmanage.Release, directory string) (map[string]string, error) {
	files := map[string]string{}
	for _, asset := range release.Assets {
		if strings.TrimSpace(asset.Name) == "" || filepath.Base(asset.Name) != asset.Name || strings.ContainsAny(asset.Name, `/\\`) {
			return nil, fmt.Errorf("release asset name is unsafe: %q", asset.Name)
		}
		if _, exists := files[asset.Name]; exists {
			return nil, fmt.Errorf("release contains duplicate asset %q", asset.Name)
		}
		parsed, err := url.Parse(asset.DownloadURL)
		if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") {
			return nil, fmt.Errorf("release asset %s has an invalid download URL", asset.Name)
		}
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, asset.DownloadURL, nil)
		if err != nil {
			return nil, err
		}
		response, err := client.Do(request)
		if err != nil {
			return nil, err
		}
		if response.StatusCode != http.StatusOK {
			response.Body.Close()
			return nil, fmt.Errorf("download %s returned %s", asset.Name, response.Status)
		}
		path := filepath.Join(directory, asset.Name)
		file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err != nil {
			response.Body.Close()
			return nil, err
		}
		_, copyErr := io.Copy(file, io.LimitReader(response.Body, 256<<20))
		response.Body.Close()
		closeErr := file.Close()
		if copyErr != nil {
			return nil, copyErr
		}
		if closeErr != nil {
			return nil, closeErr
		}
		files[asset.Name] = path
	}
	return files, nil
}

func verifyChecksums(files map[string]string, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) == 0 || (len(lines) == 1 && strings.TrimSpace(lines[0]) == "") {
		return errors.New("checksums.txt is empty")
	}
	seen := map[string]bool{}
	for _, line := range lines {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) != 2 {
			return errors.New("malformed checksum entry")
		}
		name := strings.TrimPrefix(fields[1], "*")
		if name == "checksums.txt" || filepath.Base(name) != name || strings.ContainsAny(name, `/\\`) {
			return fmt.Errorf("unsafe checksum filename %q", name)
		}
		if seen[name] {
			return fmt.Errorf("duplicate checksum entry for %s", name)
		}
		seen[name] = true
		if len(fields[0]) != sha256.Size*2 {
			return fmt.Errorf("invalid checksum for %s", name)
		}
		if _, err := hex.DecodeString(fields[0]); err != nil {
			return fmt.Errorf("invalid checksum for %s", name)
		}
		filePath, ok := files[name]
		if !ok {
			return fmt.Errorf("checksum names unpublished asset %s", name)
		}
		actual, err := digest(filePath)
		if err != nil {
			return err
		}
		if !strings.EqualFold(actual, fields[0]) {
			return fmt.Errorf("checksum mismatch for %s", name)
		}
	}
	for name := range files {
		if name != "checksums.txt" && !seen[name] {
			return fmt.Errorf("checksum missing for %s", name)
		}
	}
	return nil
}

func versionOf(path string) (string, error) {
	command := exec.Command(path, "--version")
	output, err := command.Output()
	if err != nil {
		return "", fmt.Errorf("run candidate --version: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// ensureExecutable makes a staged candidate runnable on POSIX before the
// transaction executes it. Windows relies on the executable extension instead.
func ensureExecutable(path string) error {
	if runtime.GOOS == "windows" {
		return nil
	}
	if err := os.Chmod(path, 0o755); err != nil {
		return fmt.Errorf("make candidate executable: %w", err)
	}
	return nil
}

func selfTest(path string) error {
	command := exec.Command(path, "__self-test")
	output, err := command.Output()
	if err != nil {
		return fmt.Errorf("candidate self-test failed: %w", err)
	}
	if strings.TrimSpace(string(output)) != "ok" {
		return fmt.Errorf("candidate self-test returned %q", strings.TrimSpace(string(output)))
	}
	return nil
}

func copyFile(source, destination string, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return err
	}
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	output, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	if _, err := io.Copy(output, input); err != nil {
		output.Close()
		return err
	}
	return output.Close()
}

func digest(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	_, err = io.Copy(hash, file)
	return hex.EncodeToString(hash.Sum(nil)), err
}

func snapshotFile(path string) ([]byte, bool, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return data, true, nil
}

func restoreSnapshot(path string, data []byte, existed bool) error {
	if !existed {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	return lifecycle.AtomicWrite(path, data, 0o644)
}

func withinPath(root, target string) bool {
	rel, err := filepath.Rel(root, target)
	return err == nil && rel != "." && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}

func readPrevious(path string) launcher.Pointer {
	value, err := launcher.ReadPointer(path)
	if err != nil {
		return launcher.Pointer{Schema: 1}
	}
	return value
}

func mustEncodeManifest(manifest artifact.Manifest) []byte {
	data, err := artifact.Encode(manifest)
	if err != nil {
		panic(err)
	}
	return append(data, '\n')
}

func copyDirectory(source, destination string) error {
	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(destination, rel)
		if info.IsDir() {
			return os.MkdirAll(destPath, 0o755)
		}
		return copyFile(path, destPath, info.Mode().Perm())
	})
}

func unzipArchive(archivePath, targetDir string) error {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer reader.Close()
	for _, file := range reader.File {
		cleanName := filepath.Clean(filepath.FromSlash(file.Name))
		if cleanName == "." || cleanName == ".." || strings.HasPrefix(cleanName, ".."+string(filepath.Separator)) {
			return errors.New("unsafe archive member path")
		}
		destPath := filepath.Join(targetDir, cleanName)
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(destPath, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
			return err
		}
		rc, err := file.Open()
		if err != nil {
			return err
		}
		out, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
		if err != nil {
			rc.Close()
			return err
		}
		_, copyErr := io.Copy(out, rc)
		closeErr := out.Close()
		rc.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return nil
}

type projectionSnapshot struct {
	relPath    string
	targetPath string
	backupPath string
	existed    bool
	isDir      bool
}

func snapshotProjections(op, home string, manifest, previousManifest artifact.Manifest) ([]projectionSnapshot, error) {
	seen := make(map[string]bool)
	var relPaths []string
	for _, a := range manifest.Artifacts {
		for _, p := range a.ProjectionPaths {
			if !seen[p] {
				seen[p] = true
				relPaths = append(relPaths, p)
			}
		}
	}
	for _, a := range previousManifest.Artifacts {
		for _, p := range a.ProjectionPaths {
			if !seen[p] {
				seen[p] = true
				relPaths = append(relPaths, p)
			}
		}
	}

	backupBase := filepath.Join(op, "projection-backups")
	var snapshots []projectionSnapshot
	for i, rel := range relPaths {
		targetPath := filepath.Join(home, filepath.FromSlash(rel))
		if err := lifecycle.VerifyNoLinkedAncestors(targetPath, home); err != nil {
			return nil, fmt.Errorf("verify projection target %s: %w", targetPath, err)
		}
		info, err := os.Lstat(targetPath)
		if errors.Is(err, os.ErrNotExist) {
			snapshots = append(snapshots, projectionSnapshot{
				relPath:    rel,
				targetPath: targetPath,
				existed:    false,
			})
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("stat projection target %s: %w", targetPath, err)
		}
		backupPath := filepath.Join(backupBase, fmt.Sprintf("proj-%d", i))
		if info.IsDir() {
			if err := copyDirectory(targetPath, backupPath); err != nil {
				return nil, fmt.Errorf("snapshot projection dir %s: %w", targetPath, err)
			}
			snapshots = append(snapshots, projectionSnapshot{
				relPath:    rel,
				targetPath: targetPath,
				backupPath: backupPath,
				existed:    true,
				isDir:      true,
			})
		} else {
			if err := copyFile(targetPath, backupPath, info.Mode().Perm()); err != nil {
				return nil, fmt.Errorf("snapshot projection file %s: %w", targetPath, err)
			}
			snapshots = append(snapshots, projectionSnapshot{
				relPath:    rel,
				targetPath: targetPath,
				backupPath: backupPath,
				existed:    true,
				isDir:      false,
			})
		}
	}
	return snapshots, nil
}

func applyProjections(op, home, versionDir string, manifest, previousManifest artifact.Manifest) error {
	for _, a := range manifest.Artifacts {
		if len(a.ProjectionPaths) == 0 {
			continue
		}
		if strings.HasSuffix(a.Path, ".zip") {
			zipPath := filepath.Join(versionDir, "artifacts", filepath.Base(a.Path))
			unpackedDir := filepath.Join(op, "unpacked-"+a.ID)
			if err := unzipArchive(zipPath, unpackedDir); err != nil {
				return fmt.Errorf("unpack projection archive %s: %w", a.Path, err)
			}
			sourceDir := unpackedDir
			entries, _ := os.ReadDir(unpackedDir)
			if len(entries) == 1 && entries[0].IsDir() {
				sourceDir = filepath.Join(unpackedDir, entries[0].Name())
			}
			for _, proj := range a.ProjectionPaths {
				target := filepath.Join(home, filepath.FromSlash(proj))
				_ = os.RemoveAll(target)
				if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
					return err
				}
				if err := copyDirectory(sourceDir, target); err != nil {
					return fmt.Errorf("apply projection to %s: %w", target, err)
				}
			}
		} else {
			filePath := filepath.Join(versionDir, "artifacts", filepath.Base(a.Path))
			content, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("read artifact %s: %w", a.Path, err)
			}
			for _, proj := range a.ProjectionPaths {
				target := filepath.Join(home, filepath.FromSlash(proj))
				if err := lifecycle.AtomicWrite(target, content, 0o644); err != nil {
					return fmt.Errorf("apply projection to %s: %w", target, err)
				}
			}
		}
	}

	newProjections := make(map[string]bool)
	for _, a := range manifest.Artifacts {
		for _, p := range a.ProjectionPaths {
			newProjections[p] = true
		}
	}
	for _, a := range previousManifest.Artifacts {
		for _, p := range a.ProjectionPaths {
			if !newProjections[p] {
				target := filepath.Join(home, filepath.FromSlash(p))
				_ = os.RemoveAll(target)
			}
		}
	}
	return nil
}

func rollbackProjections(snapshots []projectionSnapshot) error {
	var failures []error
	for _, s := range snapshots {
		if !s.existed {
			if err := os.RemoveAll(s.targetPath); err != nil && !os.IsNotExist(err) {
				failures = append(failures, err)
			}
		} else {
			_ = os.RemoveAll(s.targetPath)
			if s.isDir {
				if err := copyDirectory(s.backupPath, s.targetPath); err != nil {
					failures = append(failures, err)
				}
			} else {
				if info, statErr := os.Stat(s.backupPath); statErr == nil {
					if err := copyFile(s.backupPath, s.targetPath, info.Mode().Perm()); err != nil {
						failures = append(failures, err)
					}
				}
			}
		}
	}
	if len(failures) > 0 {
		return fmt.Errorf("rollback projections: %v", failures)
	}
	return nil
}
