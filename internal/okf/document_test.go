package okf

import (
	"bytes"
	"strings"
	"testing"
)

func TestParseConceptPreservesUnknownMetadataAndBytes(t *testing.T) {
	input := []byte("---\r\ntype: Decision\r\ntitle: Example\r\nproducer_extension:\r\n  nested: true\r\nbuda:\r\n  schema_version: \"1\"\r\n  uid: abc\r\n  wiki_id: wiki\r\n---\r\n\r\nBody.\r\n")
	document, err := ParseConcept("concepts/example.md", input)
	if err != nil {
		t.Fatal(err)
	}
	metadata, err := document.MetadataCopy()
	if err != nil {
		t.Fatal(err)
	}
	if metadata["producer_extension"] == nil {
		t.Fatal("unknown metadata was lost")
	}
	output, err := document.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(output, input) {
		t.Fatalf("unchanged round trip changed bytes:\n%s", output)
	}
	if err := document.Set("title", "Updated"); err != nil {
		t.Fatal(err)
	}
	output, err = document.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(output, []byte("producer_extension:")) || !bytes.Contains(output, []byte("title: Updated")) {
		t.Fatalf("changed round trip lost metadata:\n%s", output)
	}
}

func TestParseConceptRequiresOnlyTypeForBaseConformance(t *testing.T) {
	if _, err := ParseConcept("minimal.md", []byte("---\ntype: Unknown Producer Type\nnew_key: yes\n---\n")); err != nil {
		t.Fatalf("minimal forward-compatible concept rejected: %v", err)
	}
	if _, err := ParseConcept("invalid.md", []byte("---\ntitle: Missing Type\n---\n")); err == nil {
		t.Fatal("concept without type accepted")
	}
}

func TestVerifiedMappingAndSourcesDecode(t *testing.T) {
	input := []byte(`---
type: Reference
verified: { by: human:owner, at: 2026-01-01T00:00:00Z }
sources:
  - id: spec
    resource: https://example.test/spec
    extra_signal: keep-me
---
Claim.[^spec]

[^spec]: Specification
`)
	document, err := ParseConcept("reference.md", input)
	if err != nil {
		t.Fatal(err)
	}
	sources, err := document.Sources()
	if err != nil || len(sources) != 1 || sources[0].Resource == "" {
		t.Fatalf("Sources() = %#v, %v", sources, err)
	}
	output, err := document.Marshal()
	if err != nil || !strings.Contains(string(output), "extra_signal: keep-me") {
		t.Fatalf("Marshal() lost source extension: %v\n%s", err, output)
	}
}

func TestReservedFiles(t *testing.T) {
	index, err := ParseIndex([]byte("---\nokf_version: \"0.2\"\n---\n# Knowledge\n"), true)
	if err != nil || index.Version != "0.2" {
		t.Fatalf("ParseIndex() = %#v, %v", index, err)
	}
	if _, err := ParseIndex([]byte("---\nokf_version: \"0.2\"\n---\n"), false); err == nil {
		t.Fatal("nested index frontmatter accepted")
	}
	if _, err := ParseLog([]byte("# Log\n\n## July 1\n")); err == nil {
		t.Fatal("invalid log date accepted")
	}
}
