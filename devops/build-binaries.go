package main

import (
	"archive/zip"
	"bufio"
	"crypto/sha256"
	"encoding/hex"
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
)

const (
	cliName         = "buda"
	skillAssetName  = "guiho-s-0002-buda.zip"
	promptAssetName = "guiho-i-buda.md"
	outputDirectory = "dist"
)

type target struct {
	name   string
	goos   string
	goarch string
	tuning string
}

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
	flag.Parse()
	if *version == "" {
		fatalf("--version is required")
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
	assets := make([]string, 0, 10)
	for _, buildTarget := range targets {
		outputPath := filepath.Join(outputDirectory, buildTarget.name)
		ldflags := strings.Join([]string{
			"-s", "-w", "-X", "main.version=" + *version, "-X", "main.commit=" + *commit,
			"-X", "main.buildDate=" + *buildDate,
			"-X", "main.buildTarget=" + strings.TrimSuffix(buildTarget.name, ".exe"),
		}, " ")
		command := exec.Command("go", "build", "-trimpath", "-buildvcs=true", "-ldflags", ldflags, "-o", outputPath, ".")
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
		command.Env = buildEnvironment(buildTarget)
		fmt.Printf("building %s\n", buildTarget.name)
		if err := command.Run(); err != nil {
			fatalf("build %s: %v", buildTarget.name, err)
		}
		assets = append(assets, outputPath)
	}
	skillAsset := filepath.Join(outputDirectory, skillAssetName)
	if err := zipDirectory(filepath.Join("skills", "guiho-s-0002-buda"), skillAsset, stamp); err != nil {
		fatalf("create skill archive: %v", err)
	}
	assets = append(assets, skillAsset)
	promptAsset := filepath.Join(outputDirectory, promptAssetName)
	if err := copyFile(filepath.Join("prompts", promptAssetName), promptAsset); err != nil {
		fatalf("copy prompt asset: %v", err)
	}
	assets = append(assets, promptAsset)
	sort.Strings(assets)
	if err := writeChecksums(filepath.Join(outputDirectory, "checksums.txt"), assets); err != nil {
		fatalf("write checksums: %v", err)
	}
	if len(assets)+1 != 11 {
		fatalf("release must contain exactly 11 artifacts, got %d", len(assets)+1)
	}
	fmt.Println("release matrix complete: 8 binaries and 3 supporting artifacts")
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
		_ = archive.Close()
		_ = output.Close()
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
		_ = output.Close()
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
		_ = output.Close()
		return err
	}
	return output.Close()
}

func writeChecksums(path string, assets []string) error {
	output, err := os.Create(path)
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(output)
	for _, asset := range assets {
		input, err := os.Open(asset)
		if err != nil {
			_ = output.Close()
			return err
		}
		hash := sha256.New()
		_, copyErr := io.Copy(hash, input)
		closeErr := input.Close()
		if copyErr != nil || closeErr != nil {
			_ = output.Close()
			if copyErr != nil {
				return copyErr
			}
			return closeErr
		}
		if _, err := fmt.Fprintf(writer, "%s  %s\n", hex.EncodeToString(hash.Sum(nil)), filepath.Base(asset)); err != nil {
			_ = output.Close()
			return err
		}
	}
	if err := writer.Flush(); err != nil {
		_ = output.Close()
		return err
	}
	return output.Close()
}

func fatalf(format string, values ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", values...)
	os.Exit(1)
}
