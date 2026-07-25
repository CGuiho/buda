---
name: guiho-s-0002-buda
description: Maintain and retrieve one explicitly selected AI-maintained wiki through Buda, preserving OKF provenance and using Buda's qmd-backed commands for discovery.
version: "0.1.0"
metadata:
  version: "0.1.0"
---

# Buda Wiki Agent

Use this skill whenever a user explicitly asks to save, ingest, find, cite,
curate, validate, or maintain knowledge in a Buda wiki.

## Governing references

- Karpathy's LLM Wiki pattern governs the persistent agent-maintained knowledge
  workflow: <https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f>
- Google's canonical OKF specification governs portable bundle conformance:
  <https://github.com/GoogleCloudPlatform/knowledge-catalog/blob/main/okf/SPEC.md>
- Google's official OKF announcement is informative context:
  <https://cloud.google.com/blog/products/data-analytics/how-the-open-knowledge-format-can-improve-data-sharing>

## Non-negotiable routing

Confirm one intended wiki from the active repository or session, then pass its
path explicitly through `--wiki` on every repository command. Never infer a
default wiki, search several wikis, copy knowledge between repositories, or
treat instruction files as access control.

Use Buda commands for all indexing and retrieval. Never invoke qmd directly and
never implement a retrieval fallback. Buda preserves repository selection,
OKF validation, evidence normalization, and qmd compatibility at that boundary.

## Workflows

Read the focused workflow before acting:

- [Capture](references/capture.md)
- [Ingest](references/ingest.md)
- [Cited retrieval](references/retrieval.md)
- [Curation](references/curation.md)
- [Maintenance](references/maintenance.md)

After a semantic write, run `buda lint --wiki <path>` and refresh discovery
with `buda index --wiki <path>`. Report changed concept paths and their source
evidence. Surface unverified, stale, deprecated, conflicting, or missing
evidence rather than smoothing it over.
