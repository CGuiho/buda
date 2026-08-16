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

const schemaReleaseBase = "https://github.com/CGuiho/buda/releases/download/buda/v"

func GlobalPath(home string) string { return filepath.Join(home, ".guiho", "buda", GlobalFileName) }

func ProjectPath(wiki string) string { return filepath.Join(wiki, FileName) }

// MigrateLegacyGlobal performs a one-time, read-only discovery of the 0.1.x
// global buda.yaml. The legacy wiki_id is intentionally discarded because a
// global file can never select a repository. The legacy file remains in place
// until the new global file has been atomically written by the caller.
func MigrateLegacyGlobal(home string, version string) (GlobalConfig, bool, error) {
	legacy := filepath.Join(home, ".guiho", "buda", FileName)
	if _, err := os.Stat(legacy); errors.Is(err, os.ErrNotExist) {
		return GlobalConfig{}, false, nil
	} else if err != nil {
		return GlobalConfig{}, false, err
	}
	legacyConfig, err := Load(legacy)
	if err != nil {
		return GlobalConfig{}, false, fmt.Errorf("validate legacy global configuration: %w", err)
	}
	bundle, derived, qmd, agent := legacyConfig.Bundle, legacyConfig.Derived, legacyConfig.QMD, legacyConfig.Agent
	global := GlobalConfig{Schema: CurrentSchema, Bundle: &bundle, Derived: &derived, QMD: &qmd, Agent: &agent}
	if err := global.Validate(); err != nil {
		return GlobalConfig{}, false, err
	}
	data, err := MarshalGlobal(global, version)
	if err != nil {
		return GlobalConfig{}, false, err
	}
	path := GlobalPath(home)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return GlobalConfig{}, false, err
	}
	temp, err := os.CreateTemp(filepath.Dir(path), ".buda-global-migrate-*")
	if err != nil {
		return GlobalConfig{}, false, err
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)
	if _, err := temp.Write(data); err != nil {
		temp.Close()
		return GlobalConfig{}, false, err
	}
	if err := temp.Sync(); err != nil {
		temp.Close()
		return GlobalConfig{}, false, err
	}
	if err := temp.Close(); err != nil {
		return GlobalConfig{}, false, err
	}
	if err := os.Rename(tempPath, path); err != nil {
		return GlobalConfig{}, false, err
	}
	return global, true, nil
}

func DefaultGlobal() GlobalConfig {
	bundle, derived := "knowledge", ".buda"
	qmd := Default("placeholder").QMD
	agent := DefaultAgent()
	return GlobalConfig{Schema: CurrentSchema, Bundle: &bundle, QMD: &qmd, Derived: &derived, Agent: &agent}
}

func DefaultProject(wikiID string) ProjectConfig {
	return ProjectConfig{Schema: CurrentSchema, WikiID: wikiID}
}

func LoadGlobal(path string) (GlobalConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return GlobalConfig{}, fmt.Errorf("open global configuration %q: %w", path, err)
	}
	defer file.Close()
	return DecodeGlobal(file)
}

func LoadProject(path string) (ProjectConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return ProjectConfig{}, fmt.Errorf("open project configuration %q: %w", path, err)
	}
	defer file.Close()
	return DecodeProject(file)
}

func DecodeGlobal(reader io.Reader) (GlobalConfig, error) {
	var value GlobalConfig
	if err := decodeStrict(reader, &value); err != nil {
		return GlobalConfig{}, err
	}
	if value.Schema != CurrentSchema {
		return GlobalConfig{}, fmt.Errorf("global schema must be %d", CurrentSchema)
	}
	if err := validateGlobal(value); err != nil {
		return GlobalConfig{}, err
	}
	return value, nil
}

func DecodeProject(reader io.Reader) (ProjectConfig, error) {
	var value ProjectConfig
	if err := decodeStrict(reader, &value); err != nil {
		return ProjectConfig{}, err
	}
	if value.Schema != CurrentSchema {
		return ProjectConfig{}, fmt.Errorf("project schema must be %d", CurrentSchema)
	}
	if strings.TrimSpace(value.WikiID) == "" {
		return ProjectConfig{}, errors.New("wiki_id must not be empty")
	}
	if strings.ContainsAny(value.WikiID, "\r\n\x00") {
		return ProjectConfig{}, errors.New("wiki_id contains an invalid control character")
	}
	if err := validateOverrides(value.Bundle, value.QMD, value.Derived, value.Agent); err != nil {
		return ProjectConfig{}, err
	}
	return value, nil
}

func decodeStrict(reader io.Reader, target any) error {
	decoder := yaml.NewDecoder(reader)
	decoder.KnownFields(true)
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return errors.New("multiple YAML documents are not supported")
		}
		return fmt.Errorf("decode trailing YAML: %w", err)
	}
	return nil
}

func Merge(global GlobalConfig, project ProjectConfig) (Config, error) {
	if err := validateGlobal(global); err != nil {
		return Config{}, err
	}
	if err := project.Validate(); err != nil {
		return Config{}, err
	}
	base := Default(project.WikiID)
	if global.Bundle != nil {
		base.Bundle = *global.Bundle
	}
	if global.Derived != nil {
		base.Derived = *global.Derived
	}
	if global.QMD != nil {
		base.QMD = mergeQMD(base.QMD, *global.QMD)
	}
	if global.Agent != nil {
		base.Agent = mergeAgent(base.Agent, *global.Agent)
	}
	base.WikiID = project.WikiID
	if project.Bundle != nil {
		base.Bundle = *project.Bundle
	}
	if project.Derived != nil {
		base.Derived = *project.Derived
	}
	if project.QMD != nil {
		base.QMD = mergeQMD(base.QMD, *project.QMD)
	}
	if project.Agent != nil {
		base.Agent = mergeAgent(base.Agent, *project.Agent)
	}
	if err := base.Validate(); err != nil {
		return Config{}, err
	}
	return base, nil
}

func mergeQMD(base, override QMDConfig) QMDConfig {
	if strings.TrimSpace(override.Executable) != "" {
		base.Executable = override.Executable
	}
	if strings.TrimSpace(override.Collection) != "" {
		base.Collection = override.Collection
	}
	if strings.TrimSpace(override.ProjectDirectory) != "" {
		base.ProjectDirectory = override.ProjectDirectory
	}
	return base
}

func mergeAgent(base, override AgentConfig) AgentConfig {
	if validPolicy(override.Evolution.Upgrade) {
		base.Evolution.Upgrade = override.Evolution.Upgrade
	}
	if validPolicy(override.Evolution.Issues.Bugs) {
		base.Evolution.Issues.Bugs = override.Evolution.Issues.Bugs
	}
	if validPolicy(override.Evolution.Issues.Improvements) {
		base.Evolution.Issues.Improvements = override.Evolution.Issues.Improvements
	}
	if validPolicy(override.Evolution.Issues.Reviews) {
		base.Evolution.Issues.Reviews = override.Evolution.Issues.Reviews
	}
	return base
}

func validateGlobal(value GlobalConfig) error {
	if value.Bundle != nil {
		if err := validateRepositoryRelative("bundle", *value.Bundle); err != nil {
			return err
		}
	}
	if value.Derived != nil {
		if err := validateRepositoryRelative("derived", *value.Derived); err != nil {
			return err
		}
	}
	if value.QMD != nil {
		if err := validateQMD(*value.QMD); err != nil {
			return err
		}
	}
	if value.Agent != nil {
		return validateAgent(*value.Agent)
	}
	return nil
}

func validateOverrides(bundle *string, qmd *QMDConfig, derived *string, agent *AgentConfig) error {
	if bundle != nil {
		if err := validateRepositoryRelative("bundle", *bundle); err != nil {
			return err
		}
	}
	if derived != nil {
		if err := validateRepositoryRelative("derived", *derived); err != nil {
			return err
		}
	}
	if qmd != nil {
		if err := validateQMD(*qmd); err != nil {
			return err
		}
	}
	if agent != nil {
		return validateAgent(*agent)
	}
	return nil
}

func validateQMD(value QMDConfig) error {
	if strings.TrimSpace(value.Executable) != "" && strings.ContainsAny(value.Executable, "\r\n\x00") {
		return errors.New("qmd.executable contains an invalid control character")
	}
	if strings.TrimSpace(value.Collection) != "" && strings.ContainsAny(value.Collection, "\r\n\x00") {
		return errors.New("qmd.collection contains an invalid control character")
	}
	if strings.TrimSpace(value.ProjectDirectory) != "" {
		return validateRepositoryRelative("qmd.project_directory", value.ProjectDirectory)
	}
	return nil
}

func validateAgent(value AgentConfig) error {
	for name, policy := range map[string]Policy{
		"agent.evolution.upgrade":             value.Evolution.Upgrade,
		"agent.evolution.issues.bugs":         value.Evolution.Issues.Bugs,
		"agent.evolution.issues.improvements": value.Evolution.Issues.Improvements,
		"agent.evolution.issues.reviews":      value.Evolution.Issues.Reviews,
	} {
		if policy != "" && !validPolicy(policy) {
			return fmt.Errorf("%s must be disabled, always-ask, or always-proceed", name)
		}
	}
	return nil
}

func validPolicy(value Policy) bool {
	return value == PolicyDisabled || value == PolicyAlwaysAsk || value == PolicyAlwaysProceed
}

func (value ProjectConfig) Validate() error {
	if value.Schema != CurrentSchema {
		return fmt.Errorf("project schema must be %d", CurrentSchema)
	}
	if strings.TrimSpace(value.WikiID) == "" {
		return errors.New("wiki_id must not be empty")
	}
	if strings.ContainsAny(value.WikiID, "\r\n\x00") {
		return errors.New("wiki_id contains an invalid control character")
	}
	return validateOverrides(value.Bundle, value.QMD, value.Derived, value.Agent)
}

func (value GlobalConfig) Validate() error {
	if value.Schema != CurrentSchema {
		return fmt.Errorf("global schema must be %d", CurrentSchema)
	}
	return validateGlobal(value)
}

func (value Config) EffectiveAgent() AgentConfig {
	agent := value.Agent
	defaults := DefaultAgent()
	if !validPolicy(agent.Evolution.Upgrade) {
		agent.Evolution.Upgrade = defaults.Evolution.Upgrade
	}
	if !validPolicy(agent.Evolution.Issues.Bugs) {
		agent.Evolution.Issues.Bugs = defaults.Evolution.Issues.Bugs
	}
	if !validPolicy(agent.Evolution.Issues.Improvements) {
		agent.Evolution.Issues.Improvements = defaults.Evolution.Issues.Improvements
	}
	if !validPolicy(agent.Evolution.Issues.Reviews) {
		agent.Evolution.Issues.Reviews = defaults.Evolution.Issues.Reviews
	}
	return agent
}

func SchemaURL(version string, global bool) string {
	name := "buda.schema.json"
	if global {
		name = "buda.global.schema.json"
	}
	version = strings.TrimPrefix(strings.TrimSpace(version), "v")
	if version == "" {
		version = "0.2.0"
	}
	return schemaReleaseBase + version + "/" + name
}

func MarshalProject(value ProjectConfig, version string) ([]byte, error) {
	if err := value.Validate(); err != nil {
		return nil, err
	}
	data, err := yaml.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("encode project configuration: %w", err)
	}
	return append([]byte(fmt.Sprintf("# yaml-language-server: $schema=%s\n", SchemaURL(version, false))), data...), nil
}

func MarshalGlobal(value GlobalConfig, version string) ([]byte, error) {
	if err := value.Validate(); err != nil {
		return nil, err
	}
	data, err := yaml.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("encode global configuration: %w", err)
	}
	return append([]byte(fmt.Sprintf("# yaml-language-server: $schema=%s\n", SchemaURL(version, true))), data...), nil
}
