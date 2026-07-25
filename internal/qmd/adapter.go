package qmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.yaml.in/yaml/v3"
)

type Adapter struct {
	config Config
	runner Runner
}

func New(config Config, runner Runner) (*Adapter, error) {
	if config.Executable == "" {
		config.Executable = "qmd"
	}
	if config.Collection == "" {
		return nil, errors.New("qmd collection is required")
	}
	if config.WikiRoot == "" || !filepath.IsAbs(config.WikiRoot) {
		return nil, errors.New("absolute qmd wiki root is required")
	}
	if config.BundleRoot == "" || !filepath.IsAbs(config.BundleRoot) {
		return nil, errors.New("absolute qmd bundle root is required")
	}
	if _, err := containedPath(config.WikiRoot, config.BundleRoot); err != nil {
		return nil, fmt.Errorf("qmd bundle: %w", err)
	}
	if config.ProjectDirectory == "" || !filepath.IsAbs(config.ProjectDirectory) {
		return nil, errors.New("absolute qmd project directory is required")
	}
	if _, err := containedPath(config.WikiRoot, config.ProjectDirectory); err != nil {
		return nil, fmt.Errorf("qmd project directory: %w", err)
	}
	if filepath.Base(config.ProjectDirectory) != ".qmd" {
		return nil, fmt.Errorf("qmd 2.x project-local index directory must be named .qmd: %q", config.ProjectDirectory)
	}
	if config.MinimumVersion == "" {
		config.MinimumVersion = DefaultMinimumVersion
	}
	if config.MaximumVersion == "" {
		config.MaximumVersion = DefaultMaximumVersion
	}
	if config.Timeout <= 0 {
		config.Timeout = 10 * time.Minute
	}
	if runner == nil {
		runner = OSRunner{}
	}
	return &Adapter{config: config, runner: runner}, nil
}

func (adapter *Adapter) Version(ctx context.Context) (Version, error) {
	result, err := adapter.run(ctx, "version", "--version")
	if err != nil {
		return Version{}, err
	}
	version, err := ParseVersion(string(result.Stdout))
	if err != nil {
		return Version{}, &CommandError{Capability: "version", Stderr: err.Error(), Cause: err}
	}
	minimum, err := parseConstraint(adapter.config.MinimumVersion)
	if err != nil {
		return Version{}, fmt.Errorf("invalid minimum qmd version: %w", err)
	}
	maximum, err := parseConstraint(adapter.config.MaximumVersion)
	if err != nil {
		return Version{}, fmt.Errorf("invalid maximum qmd version: %w", err)
	}
	if compareVersion(version, minimum) < 0 || compareVersion(version, maximum) >= 0 {
		return Version{}, fmt.Errorf("unsupported qmd version %d.%d.%d; supported range is >=%s <%s", version.Major, version.Minor, version.Patch, adapter.config.MinimumVersion, adapter.config.MaximumVersion)
	}
	return version, nil
}

// CheckCompatibility verifies the tested qmd version range and the public CLI
// capabilities used by Buda. It intentionally treats qmd as a process, not a
// library or database.
func (adapter *Adapter) CheckCompatibility(ctx context.Context) (Compatibility, error) {
	version, err := adapter.Version(ctx)
	if err != nil {
		return Compatibility{}, err
	}
	result, err := adapter.run(ctx, "capabilities", "--help")
	if err != nil {
		return Compatibility{}, err
	}
	help := string(result.Stdout)
	required := []string{"qmd search", "qmd vsearch", "qmd query", "qmd multi-get", "qmd status", "qmd doctor", "qmd update", "qmd embed", "--format", "--collection"}
	for _, capability := range required {
		if !strings.Contains(help, capability) {
			return Compatibility{}, fmt.Errorf("qmd %d.%d.%d lacks required capability %q", version.Major, version.Minor, version.Patch, capability)
		}
	}
	return Compatibility{Version: version, Capabilities: required}, nil
}

func (adapter *Adapter) Ready(ctx context.Context) (Compatibility, error) {
	compatibility, err := adapter.CheckCompatibility(ctx)
	if err != nil {
		return Compatibility{}, err
	}
	if err := adapter.ValidateCollection(ctx); err != nil {
		return Compatibility{}, err
	}
	return compatibility, nil
}

// EnsureProject creates or validates a project-local qmd index and its one
// Buda collection. It never touches qmd's global index.
func (adapter *Adapter) EnsureProject(ctx context.Context) error {
	if _, err := adapter.CheckCompatibility(ctx); err != nil {
		return err
	}
	configPath := filepath.Join(adapter.config.ProjectDirectory, "index.yml")
	alternateConfigPath := filepath.Join(adapter.config.ProjectDirectory, "index.yaml")
	if _, err := os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
		if _, alternateErr := os.Stat(alternateConfigPath); errors.Is(alternateErr, os.ErrNotExist) {
			if _, runErr := adapter.run(ctx, "initialize project", "init"); runErr != nil {
				return runErr
			}
		}
	}
	if _, err := adapter.run(ctx, "collection mapping", "collection", "show", adapter.config.Collection); err != nil {
		relativeBundle, relErr := filepath.Rel(adapter.config.WikiRoot, adapter.config.BundleRoot)
		if relErr != nil {
			return relErr
		}
		if _, addErr := adapter.run(ctx, "collection add", "collection", "add", filepath.Clean(relativeBundle), "--name", adapter.config.Collection); addErr != nil {
			return addErr
		}
	}
	return adapter.ValidateCollection(ctx)
}

func (adapter *Adapter) ValidateCollection(ctx context.Context) error {
	if err := adapter.validateLocalConfiguration(); err != nil {
		return err
	}
	result, err := adapter.run(ctx, "collection mapping", "collection", "show", adapter.config.Collection)
	if err != nil {
		return err
	}
	var mappedPath string
	for _, line := range strings.Split(string(result.Stdout), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "Path:") {
			mappedPath = strings.TrimSpace(strings.TrimPrefix(trimmed, "Path:"))
			break
		}
	}
	if mappedPath == "" {
		return &CommandError{Capability: "collection mapping", Stderr: "qmd collection show omitted Path"}
	}
	if !filepath.IsAbs(mappedPath) {
		mappedPath = filepath.Join(adapter.config.WikiRoot, mappedPath)
	}
	mappedPath, err = filepath.Abs(mappedPath)
	if err != nil {
		return err
	}
	expected, err := filepath.Abs(adapter.config.BundleRoot)
	if err != nil {
		return err
	}
	if !samePath(mappedPath, expected) {
		return fmt.Errorf("qmd collection %q maps to %q, expected canonical bundle %q", adapter.config.Collection, mappedPath, expected)
	}
	return nil
}

func (adapter *Adapter) validateLocalConfiguration() error {
	configPath := filepath.Join(adapter.config.ProjectDirectory, "index.yml")
	if _, err := os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
		configPath = filepath.Join(adapter.config.ProjectDirectory, "index.yaml")
	}
	input, err := os.Open(configPath)
	if err != nil {
		return fmt.Errorf("open qmd project configuration %q: %w", configPath, err)
	}
	defer input.Close()
	var project struct {
		Collections map[string]struct {
			Path string `yaml:"path"`
		} `yaml:"collections"`
	}
	if err := yaml.NewDecoder(input).Decode(&project); err != nil {
		return fmt.Errorf("decode qmd project configuration %q: %w", configPath, err)
	}
	if len(project.Collections) != 1 {
		return fmt.Errorf("qmd project must contain exactly one collection, found %d", len(project.Collections))
	}
	collection, exists := project.Collections[adapter.config.Collection]
	if !exists {
		return fmt.Errorf("qmd project does not define configured collection %q", adapter.config.Collection)
	}
	mappedPath := collection.Path
	if !filepath.IsAbs(mappedPath) {
		mappedPath = filepath.Join(adapter.config.ProjectDirectory, mappedPath)
	}
	mappedPath, err = filepath.Abs(mappedPath)
	if err != nil {
		return err
	}
	expected, err := filepath.Abs(adapter.config.BundleRoot)
	if err != nil {
		return err
	}
	if !samePath(mappedPath, expected) {
		return fmt.Errorf("qmd collection %q maps to %q, expected canonical bundle %q", adapter.config.Collection, mappedPath, expected)
	}
	return nil
}

func (adapter *Adapter) Update(ctx context.Context) (IndexResult, error) {
	if err := adapter.ValidateCollection(ctx); err != nil {
		return IndexResult{}, err
	}
	result, err := adapter.run(ctx, "index", "update")
	if err != nil {
		return IndexResult{}, err
	}
	return IndexResult{Capability: "index", Output: strings.TrimSpace(string(result.Stdout))}, nil
}

func (adapter *Adapter) Embed(ctx context.Context) (IndexResult, error) {
	if err := adapter.ValidateCollection(ctx); err != nil {
		return IndexResult{}, err
	}
	result, err := adapter.run(ctx, "embed", "embed", "-c", adapter.config.Collection)
	if err != nil {
		return IndexResult{}, err
	}
	return IndexResult{Capability: "embed", Output: strings.TrimSpace(string(result.Stdout))}, nil
}

func (adapter *Adapter) Search(ctx context.Context, options SearchOptions) ([]Match, error) {
	if strings.TrimSpace(options.Text) == "" {
		return nil, errors.New("qmd query text is required")
	}
	command := "query"
	switch options.Mode {
	case "", ModeHybrid:
		options.Mode = ModeHybrid
	case ModeLexical:
		command = "search"
	case ModeSemantic:
		command = "vsearch"
	default:
		return nil, fmt.Errorf("unsupported qmd search mode %q", options.Mode)
	}
	if options.Limit <= 0 {
		options.Limit = 20
	}
	arguments := []string{command, "--format", "json", "-c", adapter.config.Collection, "-n", fmt.Sprint(options.Limit)}
	if options.Explain && options.Mode == ModeHybrid {
		arguments = append(arguments, "--explain")
	}
	arguments = append(arguments, options.Text)
	result, err := adapter.run(ctx, string(options.Mode), arguments...)
	if err != nil {
		return nil, err
	}
	var raw []map[string]any
	if err := json.Unmarshal(result.Stdout, &raw); err != nil {
		return nil, &CommandError{Capability: string(options.Mode), Stderr: "malformed qmd JSON output", Cause: err}
	}
	matches := make([]Match, 0, len(raw))
	for index, item := range raw {
		pathValue, _ := item["file"].(string)
		pathValue, err = ResolveResultPath(adapter.config.BundleRoot, adapter.config.Collection, pathValue)
		if err != nil {
			return nil, &CommandError{Capability: string(options.Mode), Stderr: err.Error(), Cause: err}
		}
		match := Match{
			Rank: index + 1, Path: pathValue, Raw: item,
			DocumentID: stringValue(item["docid"]), Title: stringValue(item["title"]),
			Snippet: stringValue(item["snippet"]), Context: stringValue(item["context"]),
			Score: numberValue(item["score"]), Line: int(numberValue(item["line"])),
		}
		if explanation, exists := item["explain"]; exists {
			match.Explanation = explanation
		}
		matches = append(matches, match)
	}
	return matches, nil
}

func (adapter *Adapter) Get(ctx context.Context, conceptPathOrID string) (Document, error) {
	value := strings.TrimSpace(conceptPathOrID)
	if value == "" {
		return Document{}, errors.New("qmd document path or result id is required")
	}
	if !strings.HasPrefix(value, "#") {
		path, err := ResolveResultPath(adapter.config.BundleRoot, adapter.config.Collection, value)
		if err != nil {
			return Document{}, err
		}
		value = "qmd://" + adapter.config.Collection + "/" + path
	}
	result, err := adapter.run(ctx, "get", "multi-get", value, "--format", "json", "--no-line-numbers", "--max-bytes", "1073741824")
	if err != nil {
		return Document{}, err
	}
	var documents []struct {
		File    string `json:"file"`
		Title   string `json:"title"`
		Context string `json:"context"`
		Body    string `json:"body"`
		Skipped bool   `json:"skipped"`
		Reason  string `json:"reason"`
	}
	if err := json.Unmarshal(result.Stdout, &documents); err != nil {
		return Document{}, &CommandError{Capability: "get", Stderr: "malformed qmd JSON output", Cause: err}
	}
	if len(documents) != 1 {
		return Document{}, &CommandError{Capability: "get", Stderr: fmt.Sprintf("qmd returned %d documents, expected exactly one", len(documents))}
	}
	if documents[0].Skipped {
		return Document{}, &CommandError{Capability: "get", Stderr: "qmd skipped document: " + documents[0].Reason}
	}
	path, err := ResolveResultPath(adapter.config.BundleRoot, adapter.config.Collection, documents[0].File)
	if err != nil {
		return Document{}, &CommandError{Capability: "get", Stderr: err.Error(), Cause: err}
	}
	return Document{Path: path, Title: documents[0].Title, Context: documents[0].Context, Body: documents[0].Body}, nil
}

func (adapter *Adapter) Status(ctx context.Context) (Diagnostic, error) {
	return adapter.diagnostic(ctx, "status")
}

func (adapter *Adapter) Doctor(ctx context.Context) (Diagnostic, error) {
	return adapter.diagnostic(ctx, "doctor")
}

func (adapter *Adapter) diagnostic(ctx context.Context, capability string) (Diagnostic, error) {
	compatibility, err := adapter.Ready(ctx)
	if err != nil {
		return Diagnostic{}, err
	}
	result, err := adapter.run(ctx, capability, capability)
	if err != nil {
		return Diagnostic{}, err
	}
	version := compatibility.Version
	return Diagnostic{Capability: capability, Version: fmt.Sprintf("%d.%d.%d", version.Major, version.Minor, version.Patch), Output: strings.TrimSpace(string(result.Stdout))}, nil
}

func (adapter *Adapter) run(ctx context.Context, capability string, arguments ...string) (ProcessResult, error) {
	requestContext, cancel := context.WithTimeout(ctx, adapter.config.Timeout)
	defer cancel()
	result, err := adapter.runner.Run(requestContext, Request{Executable: adapter.config.Executable, Arguments: append([]string(nil), arguments...), Directory: adapter.config.WikiRoot})
	if requestContext.Err() != nil {
		return result, &CommandError{Capability: capability, ExitCode: result.ExitCode, Stderr: adapter.sanitizeDiagnostics(result.Stderr), Cause: requestContext.Err()}
	}
	if err != nil {
		return result, &CommandError{Capability: capability, ExitCode: result.ExitCode, Stderr: adapter.sanitizeDiagnostics(result.Stderr), Cause: err}
	}
	if result.ExitCode != 0 {
		return result, &CommandError{Capability: capability, ExitCode: result.ExitCode, Stderr: adapter.sanitizeDiagnostics(result.Stderr)}
	}
	return result, nil
}

func (adapter *Adapter) sanitizeDiagnostics(value []byte) string {
	diagnostic := sanitizeStderr(value)
	diagnostic = strings.ReplaceAll(diagnostic, adapter.config.WikiRoot, "<wiki>")
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		diagnostic = strings.ReplaceAll(diagnostic, home, "~")
	}
	return diagnostic
}

func samePath(left, right string) bool {
	if filepath.Separator == '\\' {
		return strings.EqualFold(filepath.Clean(left), filepath.Clean(right))
	}
	return filepath.Clean(left) == filepath.Clean(right)
}

func stringValue(value any) string {
	result, _ := value.(string)
	return result
}

func numberValue(value any) float64 {
	result, _ := value.(float64)
	return result
}
