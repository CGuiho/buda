// Package capture records text the user explicitly asked to save. It does not
// call a model or infer a target concept; callers must provide the concept path
// and title.
package capture

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/CGuiho/buda/internal/okf"
	"github.com/CGuiho/buda/internal/repository"
	"github.com/CGuiho/buda/internal/source"
	"go.yaml.in/yaml/v3"
)

type Request struct {
	Target      string
	Title       string
	Description string
	Type        string
	Text        []byte
	Actor       string
	Now         time.Time
	Replace     bool
}

type Result struct {
	Path      string `json:"path"`
	UID       string `json:"uid"`
	SourceID  string `json:"source_id"`
	Digest    string `json:"digest"`
	Created   bool   `json:"created"`
	Updated   bool   `json:"updated"`
	Unchanged bool   `json:"unchanged"`
}

type frontmatter struct {
	Type        string           `yaml:"type"`
	Title       string           `yaml:"title"`
	Description string           `yaml:"description"`
	Status      string           `yaml:"status"`
	Generated   generated        `yaml:"generated"`
	Sources     []okf.Source     `yaml:"sources"`
	Buda        okf.BudaMetadata `yaml:"buda"`
}

type generated struct {
	By string `yaml:"by"`
	At string `yaml:"at"`
}

func Run(selected repository.Repository, request Request) (Result, error) {
	request.Target = filepath.Clean(request.Target)
	if strings.TrimSpace(request.Target) == "" || request.Target == "." || filepath.IsAbs(request.Target) {
		return Result{}, errors.New("capture target must be an explicit bundle-relative Markdown path")
	}
	if strings.ToLower(filepath.Ext(request.Target)) != ".md" {
		return Result{}, errors.New("capture target must end in .md")
	}
	base := strings.ToLower(filepath.Base(request.Target))
	if base == "index.md" || base == "log.md" || strings.HasPrefix(filepath.ToSlash(request.Target), "references/raw/") {
		return Result{}, errors.New("capture target must be a non-reserved concept path outside references/raw")
	}
	if strings.TrimSpace(request.Title) == "" {
		return Result{}, errors.New("capture title must not be empty")
	}
	if strings.TrimSpace(request.Type) == "" {
		request.Type = "Note"
	}
	if strings.TrimSpace(request.Actor) == "" {
		return Result{}, errors.New("capture actor must not be empty")
	}
	if len(bytes.TrimSpace(request.Text)) == 0 {
		return Result{}, errors.New("capture text must not be empty")
	}
	if request.Now.IsZero() {
		request.Now = time.Now().UTC()
	}
	if request.Description == "" {
		request.Description = "Captured knowledge: " + request.Title + "."
	}
	absolute, err := repository.ResolveContained(selected.Bundle, request.Target)
	if err != nil {
		return Result{}, err
	}
	relative := filepath.ToSlash(request.Target)
	sourceID := "capture-input"
	uid := source.StableUID("buda-concept:"+selected.Config.WikiID, relative)
	// Compute the digest over the trimmed body-before-marker so it always
	// matches the stored body: the health check re-hashes
	// bytes.TrimSpace(body-before-marker), and the body written below is
	// itself trimmed. Hashing the raw input instead would mismatch whenever
	// the input ends with trailing whitespace (e.g. piping
	// `cat file | buda capture`).
	body := bytes.TrimSpace(request.Text)
	marker := []byte("[^" + sourceID + "]")
	// Strip a trailing marker so the digest covers only the captured evidence
	// (the body-before-marker), matching what the health check re-hashes.
	digestBase := body
	if bytes.HasSuffix(body, marker) {
		digestBase = bytes.TrimSpace(bytes.TrimSuffix(body, marker))
	}
	digestBytes := sha256.Sum256(digestBase)
	digest := "sha256:" + hex.EncodeToString(digestBytes[:])
	// Append the footnote marker and definition. If the input already ends
	// with the marker, do not append it a second time: only the footnote
	// definition is added so the body is not mangled into a doubled marker
	// and the digest still matches the trimmed body-before-marker.
	definition := []byte("\n\n[^" + sourceID + "]: Explicit user-directed capture input.\n")
	if bytes.HasSuffix(body, marker) {
		body = append(append([]byte(nil), body...), definition...)
	} else {
		body = append(append([]byte(nil), body...), append(marker, '\n')...)
		body = append(body, definition...)
	}
	metadata := frontmatter{
		Type:        request.Type,
		Title:       request.Title,
		Description: request.Description,
		Status:      "draft",
		Generated:   generated{By: request.Actor, At: request.Now.UTC().Format(time.RFC3339)},
		Sources: []okf.Source{{
			ID: sourceID, Resource: "buda:capture", Title: "Explicit user-directed capture input", Author: request.Actor,
		}},
		Buda: okf.BudaMetadata{
			SchemaVersion: "1", UID: uid, WikiID: selected.Config.WikiID,
			SourceDigests: map[string]string{sourceID: digest},
		},
	}
	front, err := yaml.Marshal(metadata)
	if err != nil {
		return Result{}, fmt.Errorf("encode captured concept frontmatter: %w", err)
	}
	data := append([]byte("---\n"), front...)
	data = append(data, []byte("---\n\n")...)
	data = append(data, body...)

	created, updated, unchanged := false, false, false
	if existing, err := os.ReadFile(absolute); err == nil {
		document, parseErr := okf.ParseConcept(relative, existing)
		if parseErr != nil {
			return Result{}, fmt.Errorf("refusing to replace invalid existing concept %q: %w", relative, parseErr)
		}
		existingBody := bytes.TrimSpace(bytes.Split(document.Body, []byte("[^"+sourceID+"]"))[0])
		if document.String("type") == request.Type && document.String("title") == request.Title && document.String("description") == request.Description && bytes.Equal(existingBody, bytes.TrimSpace(request.Text)) {
			unchanged = true
		} else if !request.Replace {
			return Result{}, fmt.Errorf("concept %q already exists with different content; pass explicit replacement approval", relative)
		} else {
			updated = true
		}
	} else if errors.Is(err, os.ErrNotExist) {
		created = true
	} else {
		return Result{}, fmt.Errorf("inspect capture target %q: %w", relative, err)
	}
	if !unchanged {
		if err := repository.AtomicWriteFile(absolute, data, 0o644); err != nil {
			return Result{}, err
		}
	}
	if err := updateIndex(selected.Bundle, relative, request.Title, request.Description); err != nil {
		return Result{}, err
	}
	if err := updateLog(selected.Bundle, relative, request.Title, request.Now, updated); err != nil {
		return Result{}, err
	}
	return Result{Path: relative, UID: uid, SourceID: sourceID, Digest: digest, Created: created, Updated: updated, Unchanged: unchanged}, nil
}

func updateIndex(bundle, path, title, description string) error {
	indexPath := filepath.Join(bundle, "index.md")
	data, err := os.ReadFile(indexPath)
	if err != nil {
		return fmt.Errorf("read root index: %w", err)
	}
	link := "* [" + title + "](" + filepath.ToSlash(path) + ") - " + description
	if bytes.Contains(data, []byte("]("+filepath.ToSlash(path)+")")) {
		return nil
	}
	if len(data) > 0 && data[len(data)-1] != '\n' {
		data = append(data, '\n')
	}
	if !bytes.Contains(data, []byte("# Captured Knowledge\n")) {
		data = append(data, []byte("\n# Captured Knowledge\n")...)
	}
	data = append(data, []byte("\n"+link+"\n")...)
	return repository.AtomicWriteFile(indexPath, data, 0o644)
}

func updateLog(bundle, path, title string, now time.Time, updated bool) error {
	logPath := filepath.Join(bundle, "log.md")
	data, err := os.ReadFile(logPath)
	if err != nil {
		return fmt.Errorf("read root log: %w", err)
	}
	action := "Creation"
	if updated {
		action = "Update"
	}
	entry := fmt.Sprintf("* **%s**: %s [%s](%s).", action, action+"d", title, filepath.ToSlash(path))
	if bytes.Contains(data, []byte(entry)) {
		return nil
	}
	heading := "## " + now.UTC().Format("2006-01-02")
	text := string(data)
	if index := strings.Index(text, heading); index >= 0 {
		insert := index + len(heading)
		text = text[:insert] + "\n" + entry + text[insert:]
	} else {
		firstHeading := strings.Index(text, "\n## ")
		block := "\n\n" + heading + "\n" + entry + "\n"
		if firstHeading >= 0 {
			text = text[:firstHeading] + block + text[firstHeading:]
		} else {
			text = strings.TrimRight(text, "\r\n") + block
		}
	}
	return repository.AtomicWriteFile(logPath, []byte(text), 0o644)
}
