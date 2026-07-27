package qmd

import (
	"context"
	"time"
)

const (
	DefaultMinimumVersion = "2.5.0"
	DefaultMaximumVersion = "3.0.0"
	TestedPackageVersion  = "2.5.3"
	// TestedUpstreamRevision pins the official qmd source used to map CLI
	// arguments and JSON fixtures. qmd remains an external runtime dependency.
	TestedUpstreamRevision = "53232770867ccb16538c2c6034e7d891dffc9ce3"
)

// Config defines one project-local qmd integration. All paths must already be
// absolute and resolved by Buda's repository package.
type Config struct {
	Executable       string
	WikiRoot         string
	BundleRoot       string
	ProjectDirectory string
	Collection       string
	MinimumVersion   string
	MaximumVersion   string
	Timeout          time.Duration
}

type Request struct {
	Executable string
	Arguments  []string
	Directory  string
}

type ProcessResult struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode int
}

// Runner is the only process boundary used by Adapter.
type Runner interface {
	Run(context.Context, Request) (ProcessResult, error)
}

type Version struct {
	Raw   string `json:"raw"`
	Major int    `json:"major"`
	Minor int    `json:"minor"`
	Patch int    `json:"patch"`
}

type Compatibility struct {
	Version      Version  `json:"version"`
	Capabilities []string `json:"capabilities"`
}

type SearchMode string

const (
	ModeLexical  SearchMode = "lexical"
	ModeSemantic SearchMode = "semantic"
	ModeHybrid   SearchMode = "hybrid"
)

type SearchOptions struct {
	Mode    SearchMode
	Text    string
	Limit   int
	Explain bool
}

// Match preserves qmd's ordering and score. Buda may enrich these values with
// OKF data, but it must never rerank them.
type Match struct {
	Rank        int            `json:"rank"`
	DocumentID  string         `json:"document_id"`
	Score       float64        `json:"score"`
	Path        string         `json:"path"`
	Title       string         `json:"title,omitempty"`
	Snippet     string         `json:"snippet,omitempty"`
	Line        int            `json:"line,omitempty"`
	Context     string         `json:"context,omitempty"`
	Explanation any            `json:"explanation,omitempty"`
	Raw         map[string]any `json:"-"`
}

type Document struct {
	Path    string `json:"path"`
	Title   string `json:"title,omitempty"`
	Context string `json:"context,omitempty"`
	Body    string `json:"body"`
}

type Diagnostic struct {
	Capability string `json:"capability"`
	Version    string `json:"version,omitempty"`
	State      string `json:"state"`
	Checks     int    `json:"checks,omitempty"`
	Warnings   int    `json:"warnings,omitempty"`
	Failures   int    `json:"operational_failures,omitempty"`
	Documents  int    `json:"documents,omitempty"`
	Vectors    int    `json:"vectors,omitempty"`
	Pending    int    `json:"pending_embeddings,omitempty"`
	Output     string `json:"-"`
}

type IndexResult struct {
	Capability string `json:"capability"`
	State      string `json:"state"`
	Output     string `json:"-"`
}
