package pack

import (
	"archive/zip"
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const archiveSchema = 1

var archiveTime = time.Date(1980, time.January, 1, 0, 0, 0, 0, time.UTC)

type Options struct {
	WikiRoot   string
	BundleRoot string
	BundleName string
	WikiID     string
	Output     string
}

type File struct {
	Path   string `json:"path"`
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256"`
}

type Manifest struct {
	Schema int    `json:"schema"`
	WikiID string `json:"wiki_id"`
	Bundle string `json:"bundle"`
	Files  []File `json:"files"`
}

type Result struct {
	Output        string `json:"output"`
	ArchiveSHA256 string `json:"archive_sha256"`
	Files         int    `json:"files"`
	Bytes         int64  `json:"bytes"`
}

type sourceFile struct {
	absolute string
	archive  string
	size     int64
	digest   string
}

func Create(options Options) (Result, error) {
	if options.WikiID == "" {
		return Result{}, errors.New("pack wiki id is required")
	}
	if options.WikiRoot == "" || options.BundleRoot == "" || options.Output == "" {
		return Result{}, errors.New("pack wiki, bundle, and output paths are required")
	}
	for label, value := range map[string]string{"wiki": options.WikiRoot, "bundle": options.BundleRoot, "output": options.Output} {
		if !filepath.IsAbs(value) {
			return Result{}, fmt.Errorf("pack %s path must be absolute: %q", label, value)
		}
	}
	options.WikiRoot = filepath.Clean(options.WikiRoot)
	options.BundleRoot = filepath.Clean(options.BundleRoot)
	options.Output = filepath.Clean(options.Output)
	if _, err := contained(options.WikiRoot, options.BundleRoot); err != nil {
		return Result{}, fmt.Errorf("pack bundle: %w", err)
	}
	resolvedBundle, err := filepath.EvalSymlinks(options.BundleRoot)
	if err != nil {
		return Result{}, fmt.Errorf("resolve canonical bundle: %w", err)
	}
	resolvedOutput, err := resolveOutputPath(options.Output)
	if err != nil {
		return Result{}, err
	}
	if _, err := contained(resolvedBundle, resolvedOutput); err == nil {
		return Result{}, errors.New("pack output may not be inside the canonical bundle")
	}
	if options.BundleName == "" {
		options.BundleName = filepath.Base(options.BundleRoot)
	}
	if strings.Contains(options.BundleName, "/") || strings.Contains(options.BundleName, "\\") || options.BundleName == "." || options.BundleName == ".." {
		return Result{}, fmt.Errorf("invalid pack bundle name %q", options.BundleName)
	}
	files, err := collect(options.BundleRoot, options.BundleName)
	if err != nil {
		return Result{}, err
	}
	manifestFiles := make([]File, 0, len(files))
	for _, file := range files {
		manifestFiles = append(manifestFiles, File{Path: file.archive, Size: file.size, SHA256: file.digest})
	}
	manifestBytes, err := json.MarshalIndent(Manifest{Schema: archiveSchema, WikiID: options.WikiID, Bundle: options.BundleName, Files: manifestFiles}, "", "  ")
	if err != nil {
		return Result{}, err
	}
	manifestBytes = append(manifestBytes, '\n')
	checksums := checksums(files)
	if err := os.MkdirAll(filepath.Dir(options.Output), 0o755); err != nil {
		return Result{}, fmt.Errorf("create pack output directory: %w", err)
	}
	temporary, err := os.CreateTemp(filepath.Dir(options.Output), ".buda-pack-*.tmp")
	if err != nil {
		return Result{}, fmt.Errorf("create temporary pack: %w", err)
	}
	temporaryPath := temporary.Name()
	succeeded := false
	defer func() {
		if !succeeded {
			_ = os.Remove(temporaryPath)
		}
	}()
	if err := writeArchive(temporary, files, manifestBytes, checksums); err != nil {
		_ = temporary.Close()
		return Result{}, err
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return Result{}, fmt.Errorf("sync pack: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return Result{}, fmt.Errorf("close pack: %w", err)
	}
	if err := replace(temporaryPath, options.Output); err != nil {
		return Result{}, fmt.Errorf("replace pack output: %w", err)
	}
	succeeded = true
	info, err := os.Stat(options.Output)
	if err != nil {
		return Result{}, err
	}
	digest, err := fileDigest(options.Output)
	if err != nil {
		return Result{}, err
	}
	return Result{Output: options.Output, ArchiveSHA256: digest, Files: len(files), Bytes: info.Size()}, nil
}

func collect(bundleRoot, bundleName string) ([]sourceFile, error) {
	var files []sourceFile
	err := filepath.WalkDir(bundleRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == bundleRoot {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("pack rejects symbolic link %q", path)
		}
		if entry.IsDir() {
			return nil
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("pack supports only regular files: %q", path)
		}
		relative, err := filepath.Rel(bundleRoot, path)
		if err != nil {
			return err
		}
		digest, err := fileDigest(path)
		if err != nil {
			return err
		}
		files = append(files, sourceFile{absolute: path, archive: filepath.ToSlash(filepath.Join(bundleName, relative)), size: info.Size(), digest: digest})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("collect canonical bundle: %w", err)
	}
	sort.Slice(files, func(i, j int) bool { return files[i].archive < files[j].archive })
	return files, nil
}

func writeArchive(output io.Writer, files []sourceFile, manifest, checksumData []byte) error {
	archive := zip.NewWriter(output)
	for _, file := range files {
		if err := addFile(archive, file.archive, file.absolute); err != nil {
			_ = archive.Close()
			return err
		}
	}
	for _, generated := range []struct {
		name string
		data []byte
	}{{"manifest.json", manifest}, {"checksums.txt", checksumData}} {
		header := &zip.FileHeader{Name: generated.name, Method: zip.Deflate, Modified: archiveTime}
		header.SetMode(0o644)
		writer, err := archive.CreateHeader(header)
		if err != nil {
			_ = archive.Close()
			return err
		}
		if _, err := writer.Write(generated.data); err != nil {
			_ = archive.Close()
			return err
		}
	}
	if err := archive.Close(); err != nil {
		return fmt.Errorf("close zip archive: %w", err)
	}
	return nil
}

func addFile(archive *zip.Writer, name, source string) error {
	header := &zip.FileHeader{Name: name, Method: zip.Deflate, Modified: archiveTime}
	header.SetMode(0o644)
	writer, err := archive.CreateHeader(header)
	if err != nil {
		return err
	}
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(writer, input)
	closeErr := input.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

func checksums(files []sourceFile) []byte {
	var builder strings.Builder
	writer := bufio.NewWriter(&builder)
	for _, file := range files {
		fmt.Fprintf(writer, "%s  %s\n", file.digest, file.archive)
	}
	_ = writer.Flush()
	return []byte(builder.String())
}

func fileDigest(path string) (string, error) {
	input, err := os.Open(path)
	if err != nil {
		return "", err
	}
	hash := sha256.New()
	_, copyErr := io.Copy(hash, input)
	closeErr := input.Close()
	if copyErr != nil {
		return "", copyErr
	}
	if closeErr != nil {
		return "", closeErr
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func contained(root, candidate string) (string, error) {
	relative, err := filepath.Rel(filepath.Clean(root), filepath.Clean(candidate))
	if err != nil {
		return "", err
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) || filepath.IsAbs(relative) {
		return "", fmt.Errorf("path %q escapes %q", candidate, root)
	}
	return candidate, nil
}

func resolveOutputPath(target string) (string, error) {
	missing := make([]string, 0)
	cursor := filepath.Clean(target)
	for {
		if _, err := os.Lstat(cursor); err == nil {
			resolved, resolveErr := filepath.EvalSymlinks(cursor)
			if resolveErr != nil {
				return "", fmt.Errorf("resolve pack output symlinks: %w", resolveErr)
			}
			for index := len(missing) - 1; index >= 0; index-- {
				resolved = filepath.Join(resolved, missing[index])
			}
			return resolved, nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("inspect pack output: %w", err)
		}
		parent := filepath.Dir(cursor)
		if parent == cursor {
			return "", fmt.Errorf("no existing parent for pack output %q", target)
		}
		missing = append(missing, filepath.Base(cursor))
		cursor = parent
	}
}

func replace(source, destination string) error {
	if err := os.Rename(source, destination); err == nil {
		return nil
	}
	if err := os.Remove(destination); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return os.Rename(source, destination)
}
