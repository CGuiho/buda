package main

import (
	"archive/zip"
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/CGuiho/buda/internal/artifact"
)

const (
	cliName              = "buda"
	skillAssetName       = "guiho-s-0002-buda.zip"
	instructionAssetName = "guiho-i-buda.md"
	promptAssetName      = "guiho-p-buda.md"
	manifestAssetName    = "artifacts.json"
	outputDirectory      = "dist"
)

type target struct{ name, goos, goarch, tuning string }

var targets = []target{
	{name: "buda-linux-amd64", goos: "linux", goarch: "amd64", tuning: "GOAMD64=v1"},
	{name: "buda-linux-arm64", goos: "linux", goarch: "arm64", tuning: "GOARM64=v8.0"},
	{name: "buda-linux-armv7", goos: "linux", goarch: "arm", tuning: "GOARM=7"},
	{name: "buda-linux-armv6", goos: "linux", goarch: "arm", tuning: "GOARM=6"},
	{name: "buda-darwin-amd64", goos: "darwin", goarch: "amd64", tuning: "GOAMD64=v1"},
	{name: "buda-darwin-arm64", goos: "darwin", goarch: "arm64", tuning: "GOARM64=v8.0"},
	{name: "buda-windows-amd64.exe", goos: "windows", goarch: "amd64", tuning: "GOAMD64=v1"},
	{name: "buda-windows-arm64.exe", goos: "windows", goarch: "arm64", tuning: "GOARM64=v8.0"},
}

func main() {
	version := flag.String("version", "", "Semantic version embedded in every binary")
	commit := flag.String("commit", "unknown", "Source commit embedded in every binary")
	buildDate := flag.String("build-date", "", "Stable RFC3339 build timestamp")
	targetFilter := flag.String("target", "", "Build only a specific target binary")
	flag.Parse()
	if *version == "" {
		fatalf("--version is required")
	}
	if *targetFilter != "" {
		var selected *target
		for i, t := range targets {
			if t.name == *targetFilter || strings.TrimSuffix(t.name, ".exe") == strings.TrimSuffix(*targetFilter, ".exe") {
				selected = &targets[i]
				break
			}
		}
		if selected == nil {
			fatalf("unknown target %q", *targetFilter)
		}
		if *buildDate == "" {
			*buildDate = time.Now().UTC().Format(time.RFC3339)
		}
		if err := os.MkdirAll(outputDirectory, 0o755); err != nil {
			fatalf("create %s: %v", outputDirectory, err)
		}
		path := filepath.Join(outputDirectory, selected.name)
		if err := build(path, ".", *selected, *version, *commit, *buildDate); err != nil {
			fatalf("build %s: %v", selected.name, err)
		}
		fmt.Printf("built target %s\n", selected.name)
		return
	}
	if *buildDate == "" {
		fatalf("--build-date is required for a reproducible release archive")
	}
	stamp, err := time.Parse(time.RFC3339, *buildDate)
	if err != nil {
		fatalf("parse --build-date: %v", err)
	}
	if err := os.RemoveAll(outputDirectory); err != nil {
		fatalf("clean %s: %v", outputDirectory, err)
	}
	if err := os.MkdirAll(outputDirectory, 0o755); err != nil {
		fatalf("create %s: %v", outputDirectory, err)
	}
	resourceStage, err := os.MkdirTemp("", "buda-release-resources-")
	if err != nil {
		fatalf("create versioned resource staging directory: %v", err)
	}
	defer os.RemoveAll(resourceStage)
	versionedSkill := filepath.Join(resourceStage, "guiho-s-0002-buda")
	if err := copyVersionedTree(filepath.Join("skills", "guiho-s-0002-buda"), versionedSkill, *version); err != nil {
		fatalf("stage versioned skill resources: %v", err)
	}
	assets := []string{}
	for _, buildTarget := range targets {
		path := filepath.Join(outputDirectory, buildTarget.name)
		if err := build(path, ".", buildTarget, *version, *commit, *buildDate); err != nil {
			fatalf("build %s: %v", buildTarget.name, err)
		}
		assets = append(assets, path)
	}
	for _, buildTarget := range targets {
		launcherName := "buda-launcher-" + buildTarget.goos + "-" + buildTarget.goarch
		if buildTarget.goarch == "arm" {
			launcherName += "v" + strings.TrimPrefix(strings.TrimPrefix(buildTarget.tuning, "GOARM="), "v")
		}
		if buildTarget.goos == "windows" {
			launcherName += ".exe"
		}
		path := filepath.Join(outputDirectory, launcherName)
		if err := build(path, "./cmd/buda-launcher", buildTarget, *version, *commit, *buildDate); err != nil {
			fatalf("build launcher %s: %v", launcherName, err)
		}
		assets = append(assets, path)
	}
	supporting := []struct{ source, name string }{
		{filepath.Join("prompts", instructionAssetName), instructionAssetName},
		{filepath.Join("prompts", promptAssetName), promptAssetName},
		{filepath.Join("schemas", "buda.schema.json"), "buda.schema.json"},
		{filepath.Join("schemas", "buda.global.schema.json"), "buda.global.schema.json"},
		{filepath.Join("examples", "buda.example.yaml"), "buda.example.yaml"},
		{filepath.Join("examples", "buda.global.example.yaml"), "buda.global.example.yaml"},
	}
	for _, item := range supporting {
		destination := filepath.Join(outputDirectory, item.name)
		if err := copyVersionedFile(item.source, destination, *version); err != nil {
			fatalf("copy %s: %v", item.name, err)
		}
		assets = append(assets, destination)
	}
	skill := filepath.Join(outputDirectory, skillAssetName)
	if err := zipDirectory(versionedSkill, skill, stamp); err != nil {
		fatalf("create skill archive: %v", err)
	}
	assets = append(assets, skill)
	manifestPath := filepath.Join(outputDirectory, manifestAssetName)
	manifest := buildManifest(*version, assets)
	if err := manifest.Validate(); err != nil {
		fatalf("validate artifact manifest: %v", err)
	}
	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		fatalf("encode manifest: %v", err)
	}
	if err := os.WriteFile(manifestPath, append(manifestData, '\n'), 0o644); err != nil {
		fatalf("write manifest: %v", err)
	}
	assets = append(assets, manifestPath)
	if err := writeChecksums(filepath.Join(outputDirectory, "checksums.txt"), assets); err != nil {
		fatalf("write checksums: %v", err)
	}
	fmt.Printf("release matrix complete: %d payloads, %d launchers, %d managed resources, manifest, and checksums\n", len(targets), len(targets), len(supporting)+1)
}

func build(path, packagePath string, buildTarget target, version, commit, date string) error {
	ldflags := strings.Join([]string{"-s", "-w", "-X", "main.version=" + version, "-X", "main.commit=" + commit, "-X", "main.buildDate=" + date, "-X", "main.buildTarget=" + strings.TrimSuffix(buildTarget.name, ".exe")}, " ")
	command := exec.Command("go", "build", "-trimpath", "-buildvcs=true", "-ldflags", ldflags, "-o", path, packagePath)
	command.Stdout, command.Stderr = os.Stdout, os.Stderr
	command.Env = buildEnvironment(buildTarget)
	fmt.Printf("building %s\n", filepath.Base(path))
	return command.Run()
}

func buildManifest(version string, assets []string) artifact.Manifest {
	manifest := artifact.Manifest{Schema: artifact.CurrentSchema, CLI: cliName, Version: version, Artifacts: []artifact.Artifact{}}
	seen := map[string]bool{}
	for _, path := range assets {
		name := filepath.Base(path)
		if name == manifestAssetName {
			continue
		}
		if seen[name] {
			fatalf("duplicate release asset %q; the release unit must not publish two assets with one name", name)
		}
		seen[name] = true
		digest, err := digestFile(path)
		if err != nil {
			fatalf("digest %s: %v", name, err)
		}
		installed := filepath.ToSlash(filepath.Join("versions", version, "artifacts", name))
		id := strings.TrimSuffix(name, filepath.Ext(name))
		id = strings.ReplaceAll(id, ".", "-")
		entry := artifact.Artifact{ID: id, Version: version, Path: name, InstalledPath: installed, SHA256: digest, Ownership: artifact.OwnershipReplaceable, Replaceable: true}
		setTargetMetadata(&entry, name)
		if name == skillAssetName {
			entry.ArchiveMembers = archiveMembers(path)
			entry.ProjectionPaths = []string{".agents/skills/guiho-s-0002-buda", ".claude/skills/guiho-s-0002-buda"}
		}
		if name == instructionAssetName {
			entry.ProjectionPaths = []string{"AGENTS.md", "CLAUDE.md"}
		}
		manifest.Artifacts = append(manifest.Artifacts, entry)
	}
	manifest.Artifacts = append(manifest.Artifacts, artifact.Artifact{ID: "artifacts-json", Version: version, Path: manifestAssetName, InstalledPath: filepath.ToSlash(filepath.Join("versions", version, "artifacts", manifestAssetName)), SHA256: strings.Repeat("0", 64), Ownership: artifact.OwnershipReplaceable, Replaceable: true})
	sort.Slice(manifest.Artifacts, func(i, j int) bool { return manifest.Artifacts[i].Path < manifest.Artifacts[j].Path })
	return manifest
}

func setTargetMetadata(entry *artifact.Artifact, name string) {
	target := strings.TrimSuffix(name, ".exe")
	if strings.HasPrefix(target, "buda-launcher-") {
		trimmed := strings.TrimPrefix(target, "buda-launcher-")
		parts := strings.Split(trimmed, "-")
		if len(parts) >= 2 {
			entry.OS = parts[0]
			entry.Arch = strings.Join(parts[1:], "-")
		}
	} else if strings.HasPrefix(target, "buda-") {
		trimmed := strings.TrimPrefix(target, "buda-")
		parts := strings.Split(trimmed, "-")
		if len(parts) >= 2 {
			entry.OS = parts[0]
			entry.Arch = strings.Join(parts[1:], "-")
		}
	}
}

func archiveMembers(path string) []string {
	reader, err := zip.OpenReader(path)
	if err != nil {
		fatalf("read skill archive members: %v", err)
	}
	defer reader.Close()
	members := make([]string, 0, len(reader.File))
	for _, entry := range reader.File {
		members = append(members, filepath.ToSlash(entry.Name))
	}
	sort.Strings(members)
	return members
}

func buildEnvironment(buildTarget target) []string {
	blocked := map[string]bool{"GOOS": true, "GOARCH": true, "GOAMD64": true, "GOARM64": true, "GOARM": true, "CGO_ENABLED": true}
	environment := make([]string, 0, len(os.Environ())+4)
	for _, item := range os.Environ() {
		key, _, found := strings.Cut(item, "=")
		if found && !blocked[key] {
			environment = append(environment, item)
		}
	}
	return append(environment, "GOOS="+buildTarget.goos, "GOARCH="+buildTarget.goarch, "CGO_ENABLED=0", buildTarget.tuning)
}

func zipDirectory(source, destination string, stamp time.Time) error {
	output, err := os.Create(destination)
	if err != nil {
		return err
	}
	archive := zip.NewWriter(output)
	var files []string
	err = filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		archive.Close()
		output.Close()
		return err
	}
	sort.Strings(files)
	for _, path := range files {
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(filepath.Dir(source), path)
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(relative)
		header.Method = zip.Deflate
		header.Modified = stamp.UTC()
		header.SetMode(0o644)
		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}
		input, err := os.Open(path)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(writer, input)
		closeErr := input.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
	}
	if err := archive.Close(); err != nil {
		output.Close()
		return err
	}
	return output.Close()
}

func copyFile(source, destination string) error {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	output, err := os.Create(destination)
	if err != nil {
		return err
	}
	if _, err := io.Copy(output, input); err != nil {
		output.Close()
		return err
	}
	return output.Close()
}

func copyVersionedFile(source, destination, version string) error {
	content, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	content = []byte(strings.ReplaceAll(string(content), "0.2.0", version))
	return os.WriteFile(destination, content, 0o644)
}

func copyVersionedTree(source, destination, version string) error {
	return filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		if relative == "." {
			return os.MkdirAll(destination, 0o755)
		}
		target := filepath.Join(destination, relative)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return copyVersionedFile(path, target, version)
	})
}

func digestFile(path string) (string, error) {
	input, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer input.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, input); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func writeChecksums(path string, assets []string) error {
	sort.Slice(assets, func(i, j int) bool { return filepath.Base(assets[i]) < filepath.Base(assets[j]) })
	output, err := os.Create(path)
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(output)
	for _, asset := range assets {
		digest, err := digestFile(asset)
		if err != nil {
			output.Close()
			return err
		}
		if _, err := fmt.Fprintf(writer, "%s  %s\n", digest, filepath.Base(asset)); err != nil {
			output.Close()
			return err
		}
	}
	if err := writer.Flush(); err != nil {
		output.Close()
		return err
	}
	return output.Close()
}

func fatalf(format string, values ...any) { fmt.Fprintf(os.Stderr, format+"\n", values...); os.Exit(1) }
