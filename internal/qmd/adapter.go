package qmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"go.yaml.in/yaml/v3"
)

type Adapter struct {
	config  Config
	runner  Runner
	mu      sync.RWMutex
	version string
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
	result, err := adapter.runAt(ctx, adapter.compatibilityDirectory(), "version", "--version")
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
	adapter.setKnownVersion(fmt.Sprintf("%d.%d.%d", version.Major, version.Minor, version.Patch))
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
	result, err := adapter.runAt(ctx, adapter.compatibilityDirectory(), "capabilities", "--help")
	if err != nil {
		return Compatibility{}, err
	}
	help := string(result.Stdout)
	// qmd 2.5+ provides doctor, but its root help does not list that command.
	// Version gating establishes doctor availability; root-help proof covers the
	// commands and flags that the root help actually advertises.
	required := []string{"qmd search", "qmd vsearch", "qmd query", "qmd multi-get", "qmd status", "qmd update", "qmd embed", "--format", "--collection"}
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
	configured, err := adapter.localCollectionConfigured()
	if err != nil {
		return err
	}
	if !configured {
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
	configured, err := adapter.localCollectionConfigured()
	if err != nil {
		return err
	}
	if !configured {
		return fmt.Errorf("qmd project does not define configured collection %q", adapter.config.Collection)
	}
	return nil
}

func (adapter *Adapter) localCollectionConfigured() (bool, error) {
	configPath := filepath.Join(adapter.config.ProjectDirectory, "index.yml")
	if _, err := os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
		configPath = filepath.Join(adapter.config.ProjectDirectory, "index.yaml")
	}
	input, err := os.Open(configPath)
	if err != nil {
		return false, fmt.Errorf("open qmd project configuration %q: %w", configPath, err)
	}
	defer input.Close()
	var project struct {
		Collections map[string]struct {
			Path string `yaml:"path"`
		} `yaml:"collections"`
	}
	if err := yaml.NewDecoder(input).Decode(&project); err != nil {
		return false, fmt.Errorf("decode qmd project configuration %q: %w", configPath, err)
	}
	if len(project.Collections) == 0 {
		return false, nil
	}
	if len(project.Collections) != 1 {
		return false, fmt.Errorf("qmd project must contain exactly one collection, found %d", len(project.Collections))
	}
	collection, exists := project.Collections[adapter.config.Collection]
	if !exists {
		return false, fmt.Errorf("qmd project does not define configured collection %q", adapter.config.Collection)
	}
	mappedPath := collection.Path
	if !filepath.IsAbs(mappedPath) {
		mappedPath = filepath.Join(adapter.config.ProjectDirectory, mappedPath)
	}
	mappedPath, err = filepath.Abs(mappedPath)
	if err != nil {
		return false, err
	}
	expected, err := filepath.Abs(adapter.config.BundleRoot)
	if err != nil {
		return false, err
	}
	if !samePath(mappedPath, expected) {
		return false, fmt.Errorf("qmd collection %q maps to %q, expected canonical bundle %q", adapter.config.Collection, mappedPath, expected)
	}
	return true, nil
}

func (adapter *Adapter) Update(ctx context.Context) (IndexResult, error) {
	if err := adapter.ValidateCollection(ctx); err != nil {
		return IndexResult{}, err
	}
	result, err := adapter.run(ctx, "index", "update")
	if err != nil {
		return IndexResult{}, err
	}
	return IndexResult{Capability: "index", State: "refreshed", Output: strings.TrimSpace(string(result.Stdout))}, nil
}

func (adapter *Adapter) Embed(ctx context.Context) (IndexResult, error) {
	if err := adapter.ValidateCollection(ctx); err != nil {
		return IndexResult{}, err
	}
	result, err := adapter.run(ctx, "embed", "embed", "-c", adapter.config.Collection)
	if err != nil {
		return IndexResult{}, err
	}
	return IndexResult{Capability: "embed", State: "refreshed", Output: strings.TrimSpace(string(result.Stdout))}, nil
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
	output := strings.TrimSpace(string(result.Stdout))
	diagnostic := Diagnostic{Capability: capability, Version: fmt.Sprintf("%d.%d.%d", version.Major, version.Minor, version.Patch), State: "ready", Output: output}
	if capability == "doctor" {
		diagnostic.Checks, diagnostic.Warnings, diagnostic.Failures = parseDoctorChecks(output)
		if diagnostic.Checks == 0 {
			return Diagnostic{}, &CommandError{Capability: capability, Version: diagnostic.Version, Stderr: "qmd doctor output did not contain recognizable checks"}
		}
		if diagnostic.Failures > 0 {
			diagnostic.State = "degraded"
		}
	} else if capability == "status" {
		clean := ansiEscape.ReplaceAllString(output, "")
		if !statusTotal.MatchString(clean) || !statusVectors.MatchString(clean) {
			return Diagnostic{}, &CommandError{Capability: capability, Version: diagnostic.Version, Stderr: "qmd status output did not contain recognizable document and vector counts"}
		}
		diagnostic.Documents, diagnostic.Vectors, diagnostic.Pending = parseStatusCounts(output)
		if diagnostic.Pending > 0 {
			diagnostic.State = "degraded"
		}
	}
	return diagnostic, nil
}

var (
	ansiEscape    = regexp.MustCompile(`\x1b\[[0-9;]*m`)
	statusTotal   = regexp.MustCompile(`(?m)^\s*Total:\s*([0-9,]+)\s+files indexed`)
	statusVectors = regexp.MustCompile(`(?m)^\s*Vectors:\s*([0-9,]+)\s+embedded`)
	statusPending = regexp.MustCompile(`(?m)^\s*Pending:\s*([0-9,]+)\s+need embedding`)
)

func parseDoctorChecks(output string) (checks, warnings, failures int) {
	clean := ansiEscape.ReplaceAllString(output, "")
	for _, line := range strings.Split(clean, "\n") {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "✓ "):
			checks++
		case strings.HasPrefix(line, "⚠ "):
			checks++
			warnings++
			label := strings.TrimSpace(strings.SplitN(strings.TrimPrefix(line, "⚠ "), ":", 2)[0])
			if operationalDoctorLabel(label) {
				failures++
			}
		}
	}
	return checks, warnings, failures
}

func operationalDoctorLabel(label string) bool {
	switch label {
	case "SQLite runtime", "sqlite-vec", "index config", "model cache", "legacy fingerprint adoption",
		"embedding freshness", "mixed named embedding fingerprints", "embedding fingerprints", "embedding vector sample":
		return true
	default:
		return false
	}
}

func parseStatusCounts(output string) (documents, vectors, pending int) {
	clean := ansiEscape.ReplaceAllString(output, "")
	documents = matchCount(statusTotal, clean)
	vectors = matchCount(statusVectors, clean)
	pending = matchCount(statusPending, clean)
	return documents, vectors, pending
}

func matchCount(pattern *regexp.Regexp, value string) int {
	match := pattern.FindStringSubmatch(value)
	if len(match) != 2 {
		return 0
	}
	normalized := strings.ReplaceAll(match[1], ",", "")
	var result int
	_, _ = fmt.Sscan(normalized, &result)
	return result
}

func (adapter *Adapter) run(ctx context.Context, capability string, arguments ...string) (ProcessResult, error) {
	return adapter.runAt(ctx, adapter.config.WikiRoot, capability, arguments...)
}

func (adapter *Adapter) runAt(ctx context.Context, directory, capability string, arguments ...string) (ProcessResult, error) {
	requestContext, cancel := context.WithTimeout(ctx, adapter.config.Timeout)
	defer cancel()
	result, err := adapter.runner.Run(requestContext, Request{Executable: adapter.config.Executable, Arguments: append([]string(nil), arguments...), Directory: directory})
	if requestContext.Err() != nil {
		return result, &CommandError{Capability: capability, Version: adapter.knownVersion(), ExitCode: result.ExitCode, Stderr: adapter.sanitizeDiagnostics(result.Stderr), Cause: requestContext.Err()}
	}
	if err != nil {
		return result, &CommandError{Capability: capability, Version: adapter.knownVersion(), ExitCode: result.ExitCode, Stderr: adapter.sanitizeDiagnostics(result.Stderr), Cause: err}
	}
	if result.ExitCode != 0 {
		return result, &CommandError{Capability: capability, Version: adapter.knownVersion(), ExitCode: result.ExitCode, Stderr: adapter.sanitizeDiagnostics(result.Stderr)}
	}
	return result, nil
}

func (adapter *Adapter) setKnownVersion(version string) {
	adapter.mu.Lock()
	defer adapter.mu.Unlock()
	adapter.version = version
}

func (adapter *Adapter) knownVersion() string {
	adapter.mu.RLock()
	defer adapter.mu.RUnlock()
	return adapter.version
}

func (adapter *Adapter) compatibilityDirectory() string {
	directory := adapter.config.WikiRoot
	for {
		if info, err := os.Stat(directory); err == nil && info.IsDir() {
			return directory
		}
		parent := filepath.Dir(directory)
		if parent == directory {
			return directory
		}
		directory = parent
	}
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
