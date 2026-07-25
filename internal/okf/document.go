// Package okf parses Open Knowledge Format v0.2 documents while preserving
// producer-defined metadata. Unlike buda.yaml, concept frontmatter is
// intentionally forward-compatible and never uses KnownFields.
package okf

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"go.yaml.in/yaml/v3"
)

const Version = "0.2"

var logDateHeading = regexp.MustCompile(`(?m)^## (\d{4}-\d{2}-\d{2})[ \t]*\r?$`)

type Document struct {
	Path         string
	Frontmatter  *yaml.Node
	Body         []byte
	LineEnding   string
	original     []byte
	frontChanged bool
}

type Source struct {
	ID       string `yaml:"id" json:"id,omitempty"`
	Resource string `yaml:"resource" json:"resource"`
	Title    string `yaml:"title" json:"title,omitempty"`
	Author   string `yaml:"author" json:"author,omitempty"`
}

type BudaMetadata struct {
	SchemaVersion string            `yaml:"schema_version" json:"schema_version"`
	UID           string            `yaml:"uid" json:"uid"`
	WikiID        string            `yaml:"wiki_id" json:"wiki_id"`
	SourceDigests map[string]string `yaml:"source_digests" json:"source_digests,omitempty"`
}

func ParseConcept(path string, data []byte) (*Document, error) {
	frontmatter, body, lineEnding, err := splitFrontmatter(data)
	if err != nil {
		return nil, fmt.Errorf("parse OKF concept %q: %w", path, err)
	}
	var node yaml.Node
	if err := yaml.Unmarshal(frontmatter, &node); err != nil {
		return nil, fmt.Errorf("parse OKF frontmatter %q: %w", path, err)
	}
	if len(node.Content) != 1 || node.Content[0].Kind != yaml.MappingNode {
		return nil, fmt.Errorf("parse OKF frontmatter %q: YAML root must be a mapping", path)
	}
	document := &Document{
		Path:        path,
		Frontmatter: &node,
		Body:        append([]byte(nil), body...),
		LineEnding:  lineEnding,
		original:    append([]byte(nil), data...),
	}
	if strings.TrimSpace(document.String("type")) == "" {
		return document, errors.New("OKF concept frontmatter requires a non-empty type")
	}
	return document, nil
}

func (document *Document) String(key string) string {
	node := document.value(key)
	if node == nil || node.Kind != yaml.ScalarNode {
		return ""
	}
	return node.Value
}

func (document *Document) Strings(key string) []string {
	node := document.value(key)
	if node == nil || node.Kind != yaml.SequenceNode {
		return nil
	}
	values := make([]string, 0, len(node.Content))
	for _, child := range node.Content {
		if child.Kind == yaml.ScalarNode {
			values = append(values, child.Value)
		}
	}
	return values
}

func (document *Document) Sources() ([]Source, error) {
	node := document.value("sources")
	if node == nil {
		return nil, nil
	}
	var sources []Source
	if err := node.Decode(&sources); err != nil {
		return nil, fmt.Errorf("decode sources: %w", err)
	}
	return sources, nil
}

func (document *Document) Buda() (BudaMetadata, bool, error) {
	node := document.value("buda")
	if node == nil {
		return BudaMetadata{}, false, nil
	}
	var metadata BudaMetadata
	if err := node.Decode(&metadata); err != nil {
		return BudaMetadata{}, true, fmt.Errorf("decode buda extension: %w", err)
	}
	return metadata, true, nil
}

// MetadataCopy returns a generic copy useful to structured output. Unknown
// keys are retained because decoding begins from the full YAML node.
func (document *Document) MetadataCopy() (map[string]any, error) {
	var metadata map[string]any
	if err := document.Frontmatter.Content[0].Decode(&metadata); err != nil {
		return nil, err
	}
	return metadata, nil
}

func (document *Document) Set(key string, value any) error {
	encoded := &yaml.Node{}
	if err := encoded.Encode(value); err != nil {
		return fmt.Errorf("encode frontmatter %q: %w", key, err)
	}
	valueNode := encoded
	if encoded.Kind == yaml.DocumentNode && len(encoded.Content) == 1 {
		valueNode = encoded.Content[0]
	}
	mapping := document.Frontmatter.Content[0]
	for index := 0; index < len(mapping.Content); index += 2 {
		if mapping.Content[index].Value == key {
			mapping.Content[index+1] = valueNode
			document.frontChanged = true
			return nil
		}
	}
	mapping.Content = append(mapping.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key},
		valueNode,
	)
	document.frontChanged = true
	return nil
}

func (document *Document) SetBody(body []byte) {
	document.Body = append(document.Body[:0], body...)
	document.original = nil
}

func (document *Document) Marshal() ([]byte, error) {
	if !document.frontChanged && document.original != nil && bytes.Equal(document.Body, originalBody(document.original)) {
		return append([]byte(nil), document.original...), nil
	}
	frontmatter, err := yaml.Marshal(document.Frontmatter.Content[0])
	if err != nil {
		return nil, fmt.Errorf("marshal OKF frontmatter: %w", err)
	}
	lineEnding := document.LineEnding
	if lineEnding == "" {
		lineEnding = "\n"
	}
	frontmatter = bytes.ReplaceAll(frontmatter, []byte("\n"), []byte(lineEnding))
	result := make([]byte, 0, len(frontmatter)+len(document.Body)+16)
	result = append(result, []byte("---"+lineEnding)...)
	result = append(result, frontmatter...)
	result = append(result, []byte("---"+lineEnding)...)
	result = append(result, document.Body...)
	return result, nil
}

func (document *Document) value(key string) *yaml.Node {
	if document == nil || document.Frontmatter == nil || len(document.Frontmatter.Content) != 1 {
		return nil
	}
	mapping := document.Frontmatter.Content[0]
	for index := 0; index+1 < len(mapping.Content); index += 2 {
		if mapping.Content[index].Value == key {
			return mapping.Content[index+1]
		}
	}
	return nil
}

func splitFrontmatter(data []byte) ([]byte, []byte, string, error) {
	lineEnding := "\n"
	if bytes.Contains(data, []byte("\r\n")) {
		lineEnding = "\r\n"
	}
	opener := []byte("---" + lineEnding)
	if !bytes.HasPrefix(data, opener) {
		return nil, nil, lineEnding, errors.New("missing opening YAML frontmatter delimiter")
	}
	closing := []byte(lineEnding + "---" + lineEnding)
	index := bytes.Index(data[len(opener):], closing)
	if index < 0 {
		return nil, nil, lineEnding, errors.New("missing closing YAML frontmatter delimiter")
	}
	frontStart := len(opener)
	frontEnd := frontStart + index
	bodyStart := frontEnd + len(closing)
	return data[frontStart:frontEnd], data[bodyStart:], lineEnding, nil
}

func originalBody(data []byte) []byte {
	_, body, _, err := splitFrontmatter(data)
	if err != nil {
		return nil
	}
	return body
}

type Reserved struct {
	Kind       string
	Version    string
	Body       []byte
	LineEnding string
}

func ParseIndex(data []byte, bundleRoot bool) (Reserved, error) {
	reserved := Reserved{Kind: "index", Body: append([]byte(nil), data...), LineEnding: detectLineEnding(data)}
	if !bytes.HasPrefix(data, []byte("---\n")) && !bytes.HasPrefix(data, []byte("---\r\n")) {
		if bundleRoot {
			return reserved, nil
		}
		return reserved, nil
	}
	if !bundleRoot {
		return Reserved{}, errors.New("directory index.md must not contain frontmatter")
	}
	frontmatter, body, _, err := splitFrontmatter(data)
	if err != nil {
		return Reserved{}, err
	}
	var metadata map[string]any
	if err := yaml.Unmarshal(frontmatter, &metadata); err != nil {
		return Reserved{}, fmt.Errorf("parse root index frontmatter: %w", err)
	}
	if len(metadata) != 1 {
		return Reserved{}, errors.New("root index.md frontmatter may contain only okf_version")
	}
	version, ok := metadata["okf_version"].(string)
	if !ok || strings.TrimSpace(version) == "" {
		return Reserved{}, errors.New("root index.md okf_version must be a string")
	}
	reserved.Version = version
	reserved.Body = append([]byte(nil), body...)
	return reserved, nil
}

func ParseLog(data []byte) (Reserved, error) {
	if bytes.HasPrefix(data, []byte("---\n")) || bytes.HasPrefix(data, []byte("---\r\n")) {
		return Reserved{}, errors.New("log.md must not contain frontmatter")
	}
	for _, line := range strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n") {
		if strings.HasPrefix(line, "## ") && !logDateHeading.MatchString(line) {
			return Reserved{}, fmt.Errorf("log.md date heading %q must use YYYY-MM-DD", line)
		}
	}
	return Reserved{Kind: "log", Body: append([]byte(nil), data...), LineEnding: detectLineEnding(data)}, nil
}

func detectLineEnding(data []byte) string {
	if bytes.Contains(data, []byte("\r\n")) {
		return "\r\n"
	}
	return "\n"
}
