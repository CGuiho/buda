---
name: buda-changelog
purpose: Record user-visible Buda release scope and compatibility boundaries.
description: Chronological release notes for the Buda CLI, agent resources, OKF and qmd contracts, and distribution artifacts.
created: 2026-07-26
owner: buda-package
flags: []
tags:
  - changelog
  - releases
keywords:
  - Buda 0.0.1
  - release notes
  - eleven artifacts
---

# Changelog

This document records the scope prepared for Buda releases. An entry does not
by itself assert that a Git tag, hosted release, or binary asset was published.

## [0.0.1] - 2026-07-26

Initial release scope for the public B-U-D-A Go/Cobra CLI and embedded agent
resources.

### Added

- Complete vertical MVP command set: `init`, `capture`, `ingest`, `lint`,
  `index`, `query`, `get`, `status`, `pack`, and `doctor`.
- One fresh Cobra command tree with deterministic human output, one-document
  JSON output, root version support, and help, help-tree, depth-limited help,
  and generated help-document routes throughout the public tree.
- Strict typed `buda.yaml` decoding and semantic validation, OKF-aware wiki
  initialization, forward-compatible concept metadata, provenance and citation
  health checks, bounded source ingest, deterministic capture, and reproducible
  canonical-bundle packaging.
- Embedded `guiho-s-0002-buda` skill and Buda instruction/prompt resources,
  transactional global and explicit-wiki local skill management, bounded
  `AGENTS.md`/`CLAUDE.md` instruction management, and failure-isolated
  first-success bootstrap. Bare `buda` reconciles only the two global skill
  destinations; instruction reconciliation requires an explicitly resolved
  `--wiki` and remains confined to that selected wiki.

### Contracts

- Every repository operation requires explicit `--wiki <path>` selection. Buda
  has no default wiki, global wiki registry, implicit repository selection,
  federated search, cross-repository copy, publishing command, or fixed corpus
  role or visibility policy.
- Canonical durable knowledge uses Google's Open Knowledge Format. Buda owns
  OKF-aware orchestration, evidence normalization, provenance, citations,
  health, agent resources, and stable output while preserving unknown OKF
  metadata for forward compatibility.
- qmd is the sole external project-local indexing and retrieval engine. Buda's
  thin process adapter delegates BM25, semantic and hybrid retrieval,
  embeddings, ranking, reranking, document lookup, and index diagnostics; Buda
  provides no search or retrieval fallback.
- Hidden bootstrap workers perform no qmd, network, repository-discovery, or
  cross-repository work and cannot change foreground output or exit status.

### Distribution contract

The reproducible pure-Go build contract uses `CGO_ENABLED=0` and produces
exactly eleven artifacts:

- `buda-linux-amd64`
- `buda-linux-arm64`
- `buda-linux-armv7`
- `buda-linux-armv6`
- `buda-darwin-amd64`
- `buda-darwin-arm64`
- `buda-windows-amd64.exe`
- `buda-windows-arm64.exe`
- `guiho-s-0002-buda.zip`
- `guiho-i-buda.md`
- `checksums.txt`

These names define the build layout only. This entry does not claim that the
artifacts were published.
