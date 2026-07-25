// Package config owns Buda's strict, typed YAML configuration contract.
package config

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v3"
)

const (
	FileName      = "buda.yaml"
	CurrentSchema = 1
)

// Config is the complete supported buda.yaml shape. Configuration is never
// merged: callers load exactly one resolved file.
type Config struct {
	Schema  int       `yaml:"schema" json:"schema"`
	WikiID  string    `yaml:"wiki_id" json:"wiki_id"`
	Bundle  string    `yaml:"bundle" json:"bundle"`
	QMD     QMDConfig `yaml:"qmd" json:"qmd"`
	Derived string    `yaml:"derived" json:"derived"`
}

type QMDConfig struct {
	Executable       string `yaml:"executable" json:"executable"`
	Collection       string `yaml:"collection" json:"collection"`
	ProjectDirectory string `yaml:"project_directory" json:"project_directory"`
}

func Default(wikiID string) Config {
	return Config{
		Schema:  CurrentSchema,
		WikiID:  wikiID,
		Bundle:  "knowledge",
		Derived: ".buda",
		QMD: QMDConfig{
			Executable:       "qmd",
			Collection:       "buda-wiki",
			ProjectDirectory: ".qmd",
		},
	}
}

// Resolve implements the shared GUIHO precedence contract without merging:
// explicit path, cwd/buda.yaml, then ~/.guiho/buda/buda.yaml.
func Resolve(explicit, cwd, home string) (string, error) {
	candidates := make([]string, 0, 3)
	if explicit != "" {
		candidates = append(candidates, explicit)
	} else {
		if cwd != "" {
			candidates = append(candidates, filepath.Join(cwd, FileName))
		}
		if home != "" {
			candidates = append(candidates, filepath.Join(home, ".guiho", "buda", FileName))
		}
	}
	for _, candidate := range candidates {
		absolute, err := filepath.Abs(candidate)
		if err != nil {
			return "", fmt.Errorf("resolve configuration path %q: %w", candidate, err)
		}
		info, err := os.Stat(absolute)
		if err == nil {
			if info.IsDir() {
				return "", fmt.Errorf("configuration path %q is a directory", absolute)
			}
			return absolute, nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("inspect configuration %q: %w", absolute, err)
		}
	}
	if explicit != "" {
		return "", fmt.Errorf("configuration file not found: %s", explicit)
	}
	return "", fmt.Errorf("configuration file not found (checked project and global Buda locations)")
}

func Load(path string) (Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("open configuration %q: %w", path, err)
	}
	defer file.Close()
	configuration, err := Decode(file)
	if err != nil {
		return Config{}, fmt.Errorf("decode configuration %q: %w", path, err)
	}
	return configuration, nil
}

func Decode(reader io.Reader) (Config, error) {
	decoder := yaml.NewDecoder(reader)
	decoder.KnownFields(true)
	var configuration Config
	if err := decoder.Decode(&configuration); err != nil {
		return Config{}, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return Config{}, errors.New("multiple YAML documents are not supported")
		}
		return Config{}, fmt.Errorf("decode trailing YAML: %w", err)
	}
	if err := configuration.Validate(); err != nil {
		return Config{}, err
	}
	return configuration, nil
}

func (configuration Config) Validate() error {
	if configuration.Schema != CurrentSchema {
		return fmt.Errorf("schema must be %d", CurrentSchema)
	}
	if strings.TrimSpace(configuration.WikiID) == "" {
		return errors.New("wiki_id must not be empty")
	}
	if strings.ContainsAny(configuration.WikiID, "\r\n\x00") {
		return errors.New("wiki_id contains an invalid control character")
	}
	for name, value := range map[string]string{
		"bundle":                configuration.Bundle,
		"qmd.project_directory": configuration.QMD.ProjectDirectory,
		"derived":               configuration.Derived,
	} {
		if err := validateRepositoryRelative(name, value); err != nil {
			return err
		}
	}
	if strings.TrimSpace(configuration.QMD.Executable) == "" {
		return errors.New("qmd.executable must not be empty")
	}
	if strings.ContainsAny(configuration.QMD.Executable, "\r\n\x00") {
		return errors.New("qmd.executable contains an invalid control character")
	}
	if strings.TrimSpace(configuration.QMD.Collection) == "" {
		return errors.New("qmd.collection must not be empty")
	}
	if strings.ContainsAny(configuration.QMD.Collection, "\r\n\x00") {
		return errors.New("qmd.collection contains an invalid control character")
	}
	cleanBundle := filepath.Clean(configuration.Bundle)
	cleanQMD := filepath.Clean(configuration.QMD.ProjectDirectory)
	cleanDerived := filepath.Clean(configuration.Derived)
	if cleanBundle == cleanQMD || cleanBundle == cleanDerived || cleanQMD == cleanDerived {
		return errors.New("bundle, qmd.project_directory, and derived must be distinct paths")
	}
	return nil
}

func validateRepositoryRelative(name, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s must not be empty", name)
	}
	if strings.ContainsRune(value, '\x00') || filepath.IsAbs(value) {
		return fmt.Errorf("%s must be a repository-relative path", name)
	}
	clean := filepath.Clean(value)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return fmt.Errorf("%s must resolve within the selected wiki", name)
	}
	return nil
}

func Marshal(configuration Config) ([]byte, error) {
	if err := configuration.Validate(); err != nil {
		return nil, err
	}
	data, err := yaml.Marshal(configuration)
	if err != nil {
		return nil, fmt.Errorf("encode configuration: %w", err)
	}
	return data, nil
}
