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
	FileName       = "buda.yaml"
	GlobalFileName = "buda.global.yaml"
	CurrentSchema  = 1
)

// Config is the effective configuration after the global baseline and the
// explicitly selected project's overrides have been merged.
type Config struct {
	Schema  int         `yaml:"schema" json:"schema"`
	WikiID  string      `yaml:"wiki_id" json:"wiki_id"`
	Bundle  string      `yaml:"bundle" json:"bundle"`
	QMD     QMDConfig   `yaml:"qmd" json:"qmd"`
	Derived string      `yaml:"derived" json:"derived"`
	Agent   AgentConfig `yaml:"agent" json:"agent"`
}

type QMDConfig struct {
	Executable       string `yaml:"executable" json:"executable"`
	Collection       string `yaml:"collection" json:"collection"`
	ProjectDirectory string `yaml:"project_directory" json:"project_directory"`
}

// Policy is deliberately a string enum rather than a Boolean. This keeps
// unattended agent authority explicit and makes unknown values fail closed.
type Policy string

const (
	PolicyDisabled      Policy = "disabled"
	PolicyAlwaysAsk     Policy = "always-ask"
	PolicyAlwaysProceed Policy = "always-proceed"
)

type AgentConfig struct {
	Evolution EvolutionConfig `yaml:"evolution" json:"evolution"`
}

type EvolutionConfig struct {
	Upgrade Policy        `yaml:"upgrade" json:"upgrade"`
	Issues  IssuePolicies `yaml:"issues" json:"issues"`
}

type IssuePolicies struct {
	Bugs         Policy `yaml:"bugs" json:"bugs"`
	Improvements Policy `yaml:"improvements" json:"improvements"`
	Reviews      Policy `yaml:"reviews" json:"reviews"`
}

// GlobalConfig is the user-wide baseline stored in the CLI home. Pointers are
// intentional: an omitted value is distinguishable from an explicit value
// when the effective project configuration is merged.
type GlobalConfig struct {
	Schema  int          `yaml:"schema" json:"schema"`
	Bundle  *string      `yaml:"bundle,omitempty" json:"bundle,omitempty"`
	QMD     *QMDConfig   `yaml:"qmd,omitempty" json:"qmd,omitempty"`
	Derived *string      `yaml:"derived,omitempty" json:"derived,omitempty"`
	Agent   *AgentConfig `yaml:"agent,omitempty" json:"agent,omitempty"`
}

// ProjectConfig is the explicit-wiki configuration. WikiID is required;
// other fields override the global baseline only when present.
type ProjectConfig struct {
	Schema  int          `yaml:"schema" json:"schema"`
	WikiID  string       `yaml:"wiki_id" json:"wiki_id"`
	Bundle  *string      `yaml:"bundle,omitempty" json:"bundle,omitempty"`
	QMD     *QMDConfig   `yaml:"qmd,omitempty" json:"qmd,omitempty"`
	Derived *string      `yaml:"derived,omitempty" json:"derived,omitempty"`
	Agent   *AgentConfig `yaml:"agent,omitempty" json:"agent,omitempty"`
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
		Agent: DefaultAgent(),
	}
}

func DefaultAgent() AgentConfig {
	return AgentConfig{Evolution: EvolutionConfig{
		Upgrade: PolicyAlwaysAsk,
		Issues: IssuePolicies{
			Bugs: PolicyAlwaysAsk, Improvements: PolicyAlwaysAsk, Reviews: PolicyAlwaysAsk,
		},
	}}
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
			candidates = append(candidates, filepath.Join(home, ".guiho", "buda", GlobalFileName))
			// Legacy 0.1.x used buda.yaml in the global directory. Keep this
			// read-only discovery path for migration; new writes always use the
			// distinct global filename.
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

// LoadEffective loads the explicit project configuration and the optional
// user-wide baseline without discovering a wiki. A missing global file uses
// the embedded defaults; callers that own initialization should create it.
func LoadEffective(projectPath, home string) (Config, error) {
	project, err := LoadProject(projectPath)
	if err != nil {
		return Config{}, fmt.Errorf("load project configuration: %w", err)
	}
	global := DefaultGlobal()
	if strings.TrimSpace(home) != "" {
		path := GlobalPath(home)
		if _, statErr := os.Stat(path); statErr == nil {
			global, err = LoadGlobal(path)
			if err != nil {
				return Config{}, err
			}
		} else if !errors.Is(statErr, os.ErrNotExist) {
			return Config{}, fmt.Errorf("inspect global configuration: %w", statErr)
		}
	}
	return Merge(global, project)
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
	if err := validateAgent(configuration.Agent); err != nil {
		return err
	}
	return nil
}

func validateRepositoryRelative(name, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s must not be empty", name)
	}
	if strings.ContainsAny(value, "\r\n\x00") {
		return fmt.Errorf("%s contains an invalid control character", name)
	}
	normalized := strings.ReplaceAll(value, "\\", "/")
	if strings.ContainsRune(value, ':') || filepath.IsAbs(value) || filepath.VolumeName(value) != "" || strings.HasPrefix(normalized, "/") {
		return fmt.Errorf("%s must be a repository-relative path", name)
	}
	for _, component := range strings.Split(normalized, "/") {
		if component == ".." {
			return fmt.Errorf("%s must resolve within the selected wiki", name)
		}
	}
	clean := filepath.Clean(filepath.FromSlash(normalized))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return fmt.Errorf("%s must resolve within the selected wiki", name)
	}
	return nil
}

func Marshal(configuration Config) ([]byte, error) {
	if err := configuration.Validate(); err != nil {
		return nil, err
	}
	bundle, derived := configuration.Bundle, configuration.Derived
	qmd := configuration.QMD
	agent := configuration.Agent
	return MarshalProject(ProjectConfig{Schema: configuration.Schema, WikiID: configuration.WikiID, Bundle: &bundle, QMD: &qmd, Derived: &derived, Agent: &agent}, "0.2.0")
}
