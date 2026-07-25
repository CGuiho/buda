package health

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestScanSeparatesConformanceFromBudaHealth(t *testing.T) {
	bundle := t.TempDir()
	write(t, filepath.Join(bundle, "index.md"), "---\nokf_version: \"0.2\"\n---\n# Knowledge\n")
	write(t, filepath.Join(bundle, "log.md"), "# Wiki Update Log\n\n## 2026-07-26\n* **Initialization**: Created wiki.\n")
	write(t, filepath.Join(bundle, "minimal.md"), "---\ntype: Note\nproducer_extension: preserved\n---\nBody.\n")

	report, err := Scan(bundle, "wiki-1", time.Date(2026, 7, 26, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if !report.Conformant {
		t.Fatalf("minimal OKF concept must remain conformant: %+v", report.Findings)
	}
	if report.Healthy {
		t.Fatal("concept without Buda extension unexpectedly healthy")
	}
	if !hasCode(report, "missing_buda_extension") {
		t.Fatalf("missing Buda finding: %+v", report.Findings)
	}
}

func TestScanFindsDeterministicDuplicatesLinksCitationsAndLifecycle(t *testing.T) {
	bundle := t.TempDir()
	write(t, filepath.Join(bundle, "index.md"), "---\nokf_version: \"0.2\"\n---\n# Knowledge\n")
	write(t, filepath.Join(bundle, "log.md"), "# Log\n\n## 2026-07-26\n* initialized\n")
	concept := `---
type: Decision
title: Example
description: An example.
status: stable
stale_after: 2026-07-26
sources:
  - id: spec
    resource: https://example.test/spec
buda:
  schema_version: "1"
  uid: duplicate
  wiki_id: wiki-1
---
See [missing](missing.md). Claim.[^unknown]

[^unknown]: Unknown
`
	write(t, filepath.Join(bundle, "concepts", "one.md"), concept)
	write(t, filepath.Join(bundle, "concepts", "two.md"), concept)
	report, err := Scan(bundle, "wiki-1", time.Date(2026, 7, 26, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	for _, code := range []string{"duplicate_uid", "duplicate_content", "broken_link", "citation_without_source", "stale_concept"} {
		if !hasCode(report, code) {
			t.Errorf("missing %s: %+v", code, report.Findings)
		}
	}
}

func TestNonReservedMarkdownWithoutFrontmatterIsNonConformant(t *testing.T) {
	bundle := t.TempDir()
	write(t, filepath.Join(bundle, "index.md"), "# Knowledge\n")
	write(t, filepath.Join(bundle, "log.md"), "# Log\n")
	write(t, filepath.Join(bundle, "plain.md"), "plain markdown\n")
	report, err := Scan(bundle, "wiki", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if report.Conformant || !hasCode(report, "invalid_concept") {
		t.Fatalf("report = %+v", report)
	}
}

func TestRawMarkdownEvidenceIsNotParsedAsConcept(t *testing.T) {
	bundle := t.TempDir()
	write(t, filepath.Join(bundle, "index.md"), "---\nokf_version: \"0.2\"\n---\n# Knowledge\n")
	write(t, filepath.Join(bundle, "log.md"), "# Log\n\n## 2026-07-26\n* initialized\n")
	write(t, filepath.Join(bundle, "references", "raw", "source.md"), "raw Markdown evidence without frontmatter\n")
	report, err := Scan(bundle, "wiki", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if !report.Conformant || hasCode(report, "invalid_concept") {
		t.Fatalf("raw evidence was treated as a concept: %+v", report.Findings)
	}
}

func TestScanRejectsEscapedLinksAndTamperedRawEvidence(t *testing.T) {
	bundle := t.TempDir()
	write(t, filepath.Join(bundle, "index.md"), "---\nokf_version: \"0.2\"\n---\n# Knowledge\n\n[escape](../outside.md)\n")
	write(t, filepath.Join(bundle, "log.md"), "# Log\n\n## 2026-07-26\n* initialized\n")
	write(t, filepath.Join(bundle, "references", "raw", "evidence.source"), "tampered")
	write(t, filepath.Join(bundle, "sources", "evidence.md"), `---
type: Reference
title: Evidence
description: Evidence record.
resource: /references/raw/evidence.source
sources:
  - id: original
    resource: /references/raw/evidence.source
buda:
  schema_version: "1"
  uid: evidence
  wiki_id: wiki
  source_digests:
    original: sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
---
Claim.[^original]

[^original]: Evidence.
`)
	report, err := Scan(bundle, "wiki", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	for _, code := range []string{"escaped_link", "source_digest_mismatch"} {
		if !hasCode(report, code) {
			t.Errorf("missing %s: %+v", code, report.Findings)
		}
	}
}

func TestDuplicateContentIgnoresDistinctIdentityFrontmatter(t *testing.T) {
	bundle := t.TempDir()
	write(t, filepath.Join(bundle, "index.md"), "---\nokf_version: \"0.2\"\n---\n# Knowledge\n")
	write(t, filepath.Join(bundle, "log.md"), "# Log\n\n## 2026-07-26\n* initialized\n")
	for _, uid := range []string{"one", "two"} {
		write(t, filepath.Join(bundle, uid+".md"), "---\ntype: Note\ntitle: "+uid+"\ndescription: Example.\nbuda:\n  schema_version: \"1\"\n  uid: "+uid+"\n  wiki_id: wiki\n---\nSame substantive body.\n")
	}
	report, err := Scan(bundle, "wiki", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if !hasCode(report, "duplicate_content") {
		t.Fatalf("missing duplicate content finding: %+v", report.Findings)
	}
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func hasCode(report Report, code string) bool {
	for _, finding := range report.Findings {
		if finding.Code == code {
			return true
		}
	}
	return false
}
