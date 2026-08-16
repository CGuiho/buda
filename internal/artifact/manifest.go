// Package artifact owns the release and installed ownership manifests.
package artifact

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const CurrentSchema = 1

var semverPattern = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?(?:\+[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?$`)

type Ownership string

const (
	OwnershipReplaceable Ownership = "replaceable"
	OwnershipPersistent  Ownership = "persistent"
	OwnershipDisposable  Ownership = "disposable"
)

type Artifact struct {
	ID              string    `json:"id"`
	Version         string    `json:"version"`
	Path            string    `json:"path"`
	InstalledPath   string    `json:"installed_path"`
	SHA256          string    `json:"sha256"`
	OS              string    `json:"os,omitempty"`
	Arch            string    `json:"arch,omitempty"`
	ProjectionPaths []string  `json:"projection_paths,omitempty"`
	Ownership       Ownership `json:"ownership"`
	Replaceable     bool      `json:"replaceable"`
	Persistent      bool      `json:"persistent"`
	Disposable      bool      `json:"disposable"`
	ArchiveMembers  []string  `json:"archive_members,omitempty"`
}

type Manifest struct {
	Schema      int        `json:"schema"`
	CLI         string     `json:"cli"`
	Version     string     `json:"version"`
	GeneratedAt string     `json:"generated_at,omitempty"`
	Artifacts   []Artifact `json:"artifacts"`
}

func (m Manifest) Validate() error {
	if m.Schema != CurrentSchema {
		return fmt.Errorf("artifact manifest schema must be %d", CurrentSchema)
	}
	if strings.TrimSpace(m.CLI) != "buda" {
		return errors.New("artifact manifest cli must be buda")
	}
	if !strictSemver(m.Version) {
		return fmt.Errorf("artifact manifest version %q is not SemVer", m.Version)
	}
	if len(m.Artifacts) == 0 {
		return errors.New("artifact manifest must declare at least one artifact")
	}
	ids, paths, releasePaths := map[string]bool{}, map[string]bool{}, map[string]bool{}
	for _, a := range m.Artifacts {
		if strings.TrimSpace(a.ID) == "" || ids[a.ID] {
			return fmt.Errorf("artifact IDs must be non-empty and unique: %q", a.ID)
		}
		ids[a.ID] = true
		if a.Version != m.Version {
			return fmt.Errorf("artifact %s version %q does not match manifest version %q", a.ID, a.Version, m.Version)
		}
		if err := relativeSafe(a.Path); err != nil {
			return fmt.Errorf("artifact %s path: %w", a.ID, err)
		}
		if releasePaths[a.Path] {
			return fmt.Errorf("artifact release paths must be unique: %q", a.Path)
		}
		releasePaths[a.Path] = true
		if err := relativeSafe(a.InstalledPath); err != nil {
			return fmt.Errorf("artifact %s installed path: %w", a.ID, err)
		}
		if paths[a.InstalledPath] {
			return fmt.Errorf("artifact installed paths must be unique: %q", a.InstalledPath)
		}
		paths[a.InstalledPath] = true
		if len(a.SHA256) != sha256.Size*2 {
			return fmt.Errorf("artifact %s SHA-256 is invalid", a.ID)
		}
		if _, err := hex.DecodeString(strings.ToLower(a.SHA256)); err != nil {
			return fmt.Errorf("artifact %s SHA-256 is invalid: %w", a.ID, err)
		}
		if a.Ownership != OwnershipReplaceable && a.Ownership != OwnershipPersistent && a.Ownership != OwnershipDisposable {
			return fmt.Errorf("artifact %s ownership is invalid", a.ID)
		}
		if a.Replaceable != (a.Ownership == OwnershipReplaceable) || a.Persistent != (a.Ownership == OwnershipPersistent) || a.Disposable != (a.Ownership == OwnershipDisposable) {
			return fmt.Errorf("artifact %s ownership flags do not match ownership class", a.ID)
		}
		for _, projection := range a.ProjectionPaths {
			if err := relativeSafe(projection); err != nil {
				return fmt.Errorf("artifact %s projection path: %w", a.ID, err)
			}
		}
		for _, member := range a.ArchiveMembers {
			if err := relativeSafe(member); err != nil {
				return fmt.Errorf("artifact %s archive member: %w", a.ID, err)
			}
		}
	}
	return nil
}

func strictSemver(version string) bool {
	match := semverPattern.FindStringSubmatch(strings.TrimPrefix(version, "v"))
	if match == nil {
		return false
	}
	for _, identifier := range strings.Split(match[4], ".") {
		if len(identifier) > 1 && identifier[0] == '0' {
			if _, err := strconv.ParseUint(identifier, 10, 64); err == nil {
				return false
			}
		}
	}
	return true
}

func relativeSafe(value string) error {
	normalized := strings.ReplaceAll(value, "\\", "/")
	if strings.TrimSpace(value) == "" || filepath.IsAbs(value) || filepath.VolumeName(value) != "" || strings.ContainsRune(value, ':') || strings.ContainsRune(value, '\x00') || strings.HasPrefix(normalized, "/") {
		return errors.New("must be a non-empty relative path")
	}
	for _, component := range strings.Split(normalized, "/") {
		if component == ".." {
			return errors.New("must not traverse a parent")
		}
	}
	clean := filepath.Clean(filepath.FromSlash(normalized))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return errors.New("must not traverse a parent")
	}
	return nil
}

func Decode(reader io.Reader) (Manifest, error) {
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	var manifest Manifest
	if err := decoder.Decode(&manifest); err != nil {
		return Manifest{}, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return Manifest{}, errors.New("manifest must contain one JSON document")
	}
	if err := manifest.Validate(); err != nil {
		return Manifest{}, err
	}
	return manifest, nil
}

func Load(path string) (Manifest, error) {
	file, err := os.Open(path)
	if err != nil {
		return Manifest{}, err
	}
	defer file.Close()
	return Decode(file)
}

func Encode(m Manifest) ([]byte, error) {
	if err := m.Validate(); err != nil {
		return nil, err
	}
	return json.MarshalIndent(m, "", "  ")
}

func WriteAtomic(path string, m Manifest) error {
	data, err := Encode(m)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	temp, err := os.CreateTemp(filepath.Dir(path), ".manifest-*")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)
	if _, err = temp.Write(append(data, '\n')); err != nil {
		temp.Close()
		return err
	}
	if err = temp.Sync(); err != nil {
		temp.Close()
		return err
	}
	if err = temp.Close(); err != nil {
		return err
	}
	return os.Rename(tempPath, path)
}

func DigestFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
