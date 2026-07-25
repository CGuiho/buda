// Package health performs deterministic canonical-bundle checks. It keeps the
// deliberately permissive OKF v0.2 conformance result separate from Buda's
// stricter maintenance-profile health result.
package health

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/CGuiho/buda/internal/okf"
)

type Category string

const (
	OKFConformance Category = "okf_conformance"
	BudaHealth     Category = "buda_health"
)

type Severity string

const (
	Error   Severity = "error"
	Warning Severity = "warning"
)

type Finding struct {
	Category Category `json:"category"`
	Severity Severity `json:"severity"`
	Code     string   `json:"code"`
	Path     string   `json:"path,omitempty"`
	Message  string   `json:"message"`
	Related  []string `json:"related,omitempty"`
}

type Counts struct {
	Concepts int `json:"concepts"`
	Sources  int `json:"sources"`
	Indexes  int `json:"indexes"`
	Logs     int `json:"logs"`
}

type Report struct {
	Bundle     string    `json:"bundle"`
	Conformant bool      `json:"conformant"`
	Healthy    bool      `json:"healthy"`
	Counts     Counts    `json:"counts"`
	Findings   []Finding `json:"findings"`
}

type concept struct {
	path     string
	absolute string
	document *okf.Document
	raw      []byte
}

var (
	markdownLink    = regexp.MustCompile(`\[[^\]]*\]\(([^)]+)\)`)
	footnote        = regexp.MustCompile(`\[\^([^\]]+)\]`)
	footnoteDef     = regexp.MustCompile(`(?m)^\[\^([^\]]+)\]:`)
	footnoteDefLine = regexp.MustCompile(`(?m)^\[\^([^\]]+)\]:.*$`)
)

func Scan(bundle, wikiID string, today time.Time) (Report, error) {
	absolute, err := filepath.Abs(bundle)
	if err != nil {
		return Report{}, fmt.Errorf("resolve bundle: %w", err)
	}
	report := Report{Bundle: absolute, Conformant: true, Healthy: true}
	concepts := make([]concept, 0)
	indexBodies := make(map[string][]byte)
	rootIndex := false
	rootLog := false
	err = filepath.WalkDir(absolute, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(path), ".md") {
			return nil
		}
		relative, err := filepath.Rel(absolute, path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		// Immutable evidence under references/raw is canonical source material,
		// not an OKF concept. Raw Markdown must not be parsed as frontmatter.
		if strings.HasPrefix(strings.ToLower(relative), "references/raw/") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read canonical file %q: %w", relative, err)
		}
		switch strings.ToLower(filepath.Base(path)) {
		case "index.md":
			report.Counts.Indexes++
			isRoot := relative == "index.md"
			if isRoot {
				rootIndex = true
			}
			index, err := okf.ParseIndex(data, isRoot)
			if err != nil {
				report.add(OKFConformance, Error, "invalid_reserved_index", relative, err.Error())
			} else {
				indexBodies[relative] = index.Body
				if isRoot && index.Version != okf.Version {
					report.add(BudaHealth, Error, "okf_version_mismatch", relative, fmt.Sprintf("Buda requires root okf_version %q", okf.Version))
				}
			}
		case "log.md":
			report.Counts.Logs++
			if relative == "log.md" {
				rootLog = true
			}
			if _, err := okf.ParseLog(data); err != nil {
				report.add(OKFConformance, Error, "invalid_reserved_log", relative, err.Error())
			}
		default:
			document, err := okf.ParseConcept(relative, data)
			if err != nil {
				report.add(OKFConformance, Error, "invalid_concept", relative, err.Error())
				return nil
			}
			report.Counts.Concepts++
			if strings.HasPrefix(relative, "sources/") {
				report.Counts.Sources++
			}
			concepts = append(concepts, concept{path: relative, absolute: path, document: document, raw: data})
		}
		return nil
	})
	if err != nil {
		return Report{}, fmt.Errorf("scan bundle %q: %w", absolute, err)
	}
	if !rootIndex {
		report.add(BudaHealth, Error, "missing_root_index", "index.md", "Buda profile requires knowledge/index.md")
	}
	if !rootLog {
		report.add(BudaHealth, Error, "missing_root_log", "log.md", "Buda profile requires knowledge/log.md")
	}
	checkConcepts(&report, concepts, indexBodies, wikiID, today)
	report.sort()
	return report, nil
}

func checkConcepts(report *Report, concepts []concept, indexBodies map[string][]byte, wikiID string, today time.Time) {
	byUID := map[string][]string{}
	byPath := map[string][]string{}
	byResource := map[string][]string{}
	byDigest := map[string][]string{}
	byContent := map[string][]string{}
	inbound := map[string]int{}
	known := map[string]bool{}
	for _, item := range concepts {
		known[item.path] = true
	}
	for indexPath, body := range indexBodies {
		for _, target := range links(body) {
			resolved, internal := resolveLink(indexPath, target)
			if !internal {
				continue
			}
			if escapedBundle(resolved) {
				report.add(BudaHealth, Error, "escaped_link", indexPath, fmt.Sprintf("link target %q escapes the canonical bundle", target))
				continue
			}
			absolute := filepath.Join(report.Bundle, filepath.FromSlash(resolved))
			if _, err := os.Stat(absolute); err != nil {
				report.add(BudaHealth, Error, "broken_link", indexPath, fmt.Sprintf("link target %q does not exist", target))
				continue
			}
			if known[resolved] {
				inbound[resolved]++
			}
		}
	}
	for _, item := range concepts {
		document := item.document
		if strings.TrimSpace(document.String("title")) == "" {
			report.add(BudaHealth, Warning, "missing_recommended_title", item.path, "title is recommended by OKF and required for healthy Buda discovery")
		}
		if strings.TrimSpace(document.String("description")) == "" {
			report.add(BudaHealth, Warning, "missing_recommended_description", item.path, "description is recommended by OKF and required for healthy Buda discovery")
		}
		status := document.String("status")
		if status != "" && status != "draft" && status != "stable" && status != "deprecated" {
			report.add(BudaHealth, Error, "invalid_status", item.path, "status must be draft, stable, or deprecated")
		}
		if staleAfter := document.String("stale_after"); staleAfter != "" {
			date, err := time.Parse("2006-01-02", staleAfter)
			if err != nil {
				report.add(BudaHealth, Error, "invalid_stale_after", item.path, "stale_after must use YYYY-MM-DD")
			} else if !today.Before(date) {
				report.add(BudaHealth, Warning, "stale_concept", item.path, fmt.Sprintf("concept is stale on or after %s", staleAfter))
			}
		}
		metadata, present, err := document.Buda()
		if err != nil {
			report.add(BudaHealth, Error, "invalid_buda_extension", item.path, err.Error())
		} else if !present {
			report.add(BudaHealth, Error, "missing_buda_extension", item.path, "Buda profile requires a buda metadata mapping")
		} else {
			if metadata.SchemaVersion != "1" {
				report.add(BudaHealth, Error, "buda_schema_mismatch", item.path, "buda.schema_version must be \"1\"")
			}
			if strings.TrimSpace(metadata.UID) == "" {
				report.add(BudaHealth, Error, "missing_uid", item.path, "buda.uid must not be empty")
			} else {
				byUID[metadata.UID] = append(byUID[metadata.UID], item.path)
			}
			if metadata.WikiID != wikiID {
				report.add(BudaHealth, Error, "wiki_id_mismatch", item.path, fmt.Sprintf("buda.wiki_id %q does not match selected wiki %q", metadata.WikiID, wikiID))
			}
			for sourceID, digest := range metadata.SourceDigests {
				digestHex := strings.TrimPrefix(digest, "sha256:")
				_, decodeErr := hex.DecodeString(digestHex)
				if sourceID == "" || !strings.HasPrefix(digest, "sha256:") || len(digestHex) != sha256.Size*2 || decodeErr != nil {
					report.add(BudaHealth, Error, "invalid_source_digest", item.path, fmt.Sprintf("source digest %q must be sha256:<64 hex characters>", sourceID))
					continue
				}
				byDigest[digest] = append(byDigest[digest], item.path+"#"+sourceID)
			}
		}
		normalizedPath := strings.ToLower(filepath.ToSlash(filepath.Clean(item.path)))
		byPath[normalizedPath] = append(byPath[normalizedPath], item.path)
		if resource := strings.TrimSpace(document.String("resource")); resource != "" {
			byResource[resource] = append(byResource[resource], item.path)
		}
		// Frontmatter contains identity and provenance. Duplicate knowledge is a
		// duplicate normalized concept body, even when those fields differ.
		normalizedContent := strings.Join(strings.Fields(strings.ReplaceAll(string(document.Body), "\r\n", "\n")), " ")
		if normalizedContent != "" {
			digest := fmt.Sprintf("%x", sha256.Sum256([]byte(normalizedContent)))
			byContent[digest] = append(byContent[digest], item.path)
		}
		checkSourcesAndCitations(report, item)
		for _, target := range links(document.Body) {
			resolved, internal := resolveLink(item.path, target)
			if !internal {
				continue
			}
			if escapedBundle(resolved) {
				report.add(BudaHealth, Error, "escaped_link", item.path, fmt.Sprintf("link target %q escapes the canonical bundle", target))
				continue
			}
			absolute := filepath.Join(report.Bundle, filepath.FromSlash(resolved))
			if _, err := os.Stat(absolute); err != nil {
				report.add(BudaHealth, Error, "broken_link", item.path, fmt.Sprintf("link target %q does not exist", target))
				continue
			}
			if known[resolved] {
				inbound[resolved]++
			}
		}
	}
	addDuplicates(report, byUID, "duplicate_uid", "buda.uid")
	addDuplicates(report, byPath, "duplicate_normalized_path", "normalized concept path")
	addDuplicates(report, byResource, "duplicate_resource", "canonical resource")
	addDuplicates(report, byDigest, "duplicate_source_digest", "source digest")
	addDuplicates(report, byContent, "duplicate_content", "normalized concept content")
	for _, item := range concepts {
		if inbound[item.path] == 0 {
			report.add(BudaHealth, Warning, "orphan_concept", item.path, "concept has no inbound concept link")
		}
	}
}

func checkSourcesAndCitations(report *Report, item concept) {
	sources, err := item.document.Sources()
	if err != nil {
		report.add(BudaHealth, Error, "invalid_sources", item.path, err.Error())
		return
	}
	ids := map[string]int{}
	resources := map[string]string{}
	for _, source := range sources {
		if strings.TrimSpace(source.Resource) == "" {
			report.add(BudaHealth, Error, "missing_source_resource", item.path, "each sources entry must contain resource")
		}
		if strings.TrimSpace(source.ID) == "" {
			report.add(BudaHealth, Error, "missing_source_id", item.path, "each Buda source entry must contain an id for claim citation joins")
			continue
		}
		ids[source.ID]++
		resources[source.ID] = strings.TrimSpace(source.Resource)
		if resource := strings.TrimSpace(source.Resource); strings.HasPrefix(resource, "/references/") {
			local := strings.TrimPrefix(resource, "/")
			if escapedBundle(local) {
				report.add(BudaHealth, Error, "escaped_source_resource", item.path, fmt.Sprintf("source %q resource escapes the canonical bundle", source.ID))
			} else if _, err := os.Stat(filepath.Join(report.Bundle, filepath.FromSlash(local))); err != nil {
				report.add(BudaHealth, Error, "missing_source_artifact", item.path, fmt.Sprintf("source %q local resource %q does not exist", source.ID, resource))
			}
		}
	}
	if metadata, present, err := item.document.Buda(); err == nil && present {
		for sourceID := range resources {
			if _, exists := metadata.SourceDigests[sourceID]; !exists {
				report.add(BudaHealth, Error, "source_without_digest", item.path, fmt.Sprintf("source %q has no matching buda.source_digests entry", sourceID))
			}
		}
		for sourceID, expected := range metadata.SourceDigests {
			resource, exists := resources[sourceID]
			if !exists {
				report.add(BudaHealth, Error, "digest_without_source", item.path, fmt.Sprintf("source digest %q has no matching sources[].id", sourceID))
				continue
			}
			if resource == "buda:capture" {
				captured := bytes.TrimSpace(bytes.Split(item.document.Body, []byte("[^"+sourceID+"]"))[0])
				actual := fmt.Sprintf("sha256:%x", sha256.Sum256(captured))
				if !strings.EqualFold(actual, expected) {
					report.add(BudaHealth, Error, "source_digest_mismatch", item.path, fmt.Sprintf("source %q digest does not match captured evidence", sourceID))
				}
				continue
			}
			if !strings.HasPrefix(resource, "/references/raw/") {
				continue
			}
			local := strings.TrimPrefix(resource, "/")
			if escapedBundle(local) {
				continue
			}
			data, readErr := os.ReadFile(filepath.Join(report.Bundle, filepath.FromSlash(local)))
			if readErr != nil {
				continue
			}
			actual := fmt.Sprintf("sha256:%x", sha256.Sum256(data))
			if !strings.EqualFold(actual, expected) {
				report.add(BudaHealth, Error, "source_digest_mismatch", item.path, fmt.Sprintf("source %q digest does not match local raw evidence", sourceID))
			}
		}
	}
	for id, count := range ids {
		if count > 1 {
			report.add(BudaHealth, Error, "duplicate_source_id", item.path, fmt.Sprintf("source id %q occurs %d times", id, count))
		}
	}
	definitions := map[string]bool{}
	for _, match := range footnoteDef.FindAllSubmatch(item.document.Body, -1) {
		definitions[string(match[1])] = true
	}
	claims := map[string]bool{}
	claimBody := footnoteDefLine.ReplaceAll(item.document.Body, nil)
	for _, match := range footnote.FindAllSubmatch(claimBody, -1) {
		id := string(match[1])
		claims[id] = true
		if ids[id] == 0 {
			report.add(BudaHealth, Error, "citation_without_source", item.path, fmt.Sprintf("footnote %q has no matching sources[].id", id))
		}
		if !definitions[id] {
			report.add(BudaHealth, Error, "citation_without_definition", item.path, fmt.Sprintf("footnote %q has no definition", id))
		}
	}
	for id := range definitions {
		if ids[id] == 0 {
			report.add(BudaHealth, Error, "citation_without_source", item.path, fmt.Sprintf("footnote definition %q has no matching sources[].id", id))
		}
	}
	for id := range ids {
		if !claims[id] {
			report.add(BudaHealth, Error, "source_without_claim_citation", item.path, fmt.Sprintf("source %q is not cited by a claim footnote", id))
		}
		if !definitions[id] {
			report.add(BudaHealth, Error, "source_without_footnote_definition", item.path, fmt.Sprintf("source %q has no footnote definition", id))
		}
	}
}

func links(body []byte) []string {
	matches := markdownLink.FindAllSubmatch(body, -1)
	result := make([]string, 0, len(matches))
	for _, match := range matches {
		result = append(result, strings.TrimSpace(strings.Trim(string(match[1]), "<>")))
	}
	return result
}

func resolveLink(from, target string) (string, bool) {
	if target == "" || strings.HasPrefix(target, "#") {
		return "", false
	}
	parsed, err := url.Parse(target)
	if err != nil || parsed.IsAbs() || parsed.Host != "" || strings.HasPrefix(target, "//") {
		return "", false
	}
	path := parsed.Path
	if path == "" {
		return "", false
	}
	if strings.HasPrefix(path, "/") {
		path = strings.TrimPrefix(path, "/")
	} else {
		path = filepath.ToSlash(filepath.Join(filepath.Dir(filepath.FromSlash(from)), filepath.FromSlash(path)))
	}
	clean := filepath.ToSlash(filepath.Clean(filepath.FromSlash(path)))
	if clean == ".." || strings.HasPrefix(clean, "../") {
		return clean, true
	}
	return clean, true
}

func escapedBundle(path string) bool {
	clean := filepath.ToSlash(filepath.Clean(filepath.FromSlash(path)))
	return clean == ".." || strings.HasPrefix(clean, "../") || filepath.IsAbs(filepath.FromSlash(path))
}

func addDuplicates(report *Report, values map[string][]string, code, label string) {
	keys := make([]string, 0, len(values))
	for value := range values {
		keys = append(keys, value)
	}
	sort.Strings(keys)
	for _, value := range keys {
		paths := values[value]
		if len(paths) < 2 {
			continue
		}
		sort.Strings(paths)
		report.Findings = append(report.Findings, Finding{
			Category: BudaHealth,
			Severity: Error,
			Code:     code,
			Path:     paths[0],
			Message:  fmt.Sprintf("duplicate %s %q", label, value),
			Related:  append([]string(nil), paths...),
		})
		report.Healthy = false
	}
}

func (report *Report) add(category Category, severity Severity, code, path, message string) {
	report.Findings = append(report.Findings, Finding{Category: category, Severity: severity, Code: code, Path: path, Message: message})
	if severity == Error {
		if category == OKFConformance {
			report.Conformant = false
		}
		report.Healthy = false
	}
}

func (report *Report) sort() {
	sort.SliceStable(report.Findings, func(left, right int) bool {
		a, b := report.Findings[left], report.Findings[right]
		return strings.Join([]string{string(a.Category), a.Path, a.Code, a.Message}, "\x00") < strings.Join([]string{string(b.Category), b.Path, b.Code, b.Message}, "\x00")
	})
}
