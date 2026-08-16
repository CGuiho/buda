---
name: guiho-s-0002-buda
purpose: Guide agents through explicit-wiki capture, ingest, cited retrieval, curation, and maintenance with Buda.
description: Maintain and retrieve one explicitly selected AI-maintained wiki through Buda, preserving OKF provenance and using Buda's qmd-backed commands for discovery.
created: 2026-07-26
version: "0.2.0"
owner: buda-skill-0002
flags: []
tags:
  - agent-skill
  - knowledge
keywords:
  - Buda
  - explicit wiki
  - OKF
  - qmd
metadata:
  version: "0.2.0"
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

## CLI Evolution and Feedback

The canonical Buda repository is <https://github.com/CGuiho/buda> and the
canonical issue-creation URL is <https://github.com/CGuiho/buda/issues/new>.
While using Buda, treat an observed incorrect result or failure as a bug, a
useful safety, clarity, reliability, or efficiency improvement as an
improvement, and an actionable review observation about the CLI or its
artifacts as a review finding. Read the effective `agent.evolution` policy
before upgrading Buda or creating an issue.

Each policy is independently one of `disabled`, `always-ask`, or
`always-proceed`. `disabled` prohibits the governed action and does not ask;
`always-ask` explains the proposed action and waits for permission;
`always-proceed` authorizes the action without another question and requires a
result report. When an upgrade is available, check it while using the CLI and
follow `agent.evolution.upgrade`: do not act when disabled, ask before acting
when always-ask, and upgrade when always-proceed. After every successful
upgrade, run `buda init --wiki <path>` and verify the raw `buda --version`
output.

For bugs, improvements, and reviews, apply the corresponding policy field
before posting. When issue creation is authorized, create the issue in the
canonical repository and provide the direct URL returned by GitHub. Never
claim that an issue was created without a successful response and URL.
