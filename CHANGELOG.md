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
  - Buda 0.2.0
  - Buda 0.1.1
  - release notes
  - manifest-derived release
  - GitHub Releases
---

# Changelog

This document records the scope prepared for Buda releases. An entry does not
by itself assert that a Git tag, hosted release, or binary asset was published.

## [0.2.0] - 2026-08-17

This release migrates Buda onto GUIHO CLI Convention 0001 while preserving
Buda's explicit-wiki and qmd-only product boundaries.

### Added

- Separate global and project configuration contracts with strict offline JSON
  Schemas, complete examples, version-pinned schema comments, and
  `agent.evolution` policy inheritance.
- The confirmed `guiho-p-buda` setup prompt, typed `guiho-i-buda` instruction,
  evolution guidance in `guiho-s-0002-buda`, stable launcher, immutable
  payload layout, token-owned lifecycle journals, and manifest-based ownership.
- `artifacts.json` and complete checksum-derived release packaging for all
  payloads, launchers, schemas, examples, and agent resources.
- Synchronous upgrade recovery output and ownership-safe uninstall options for
  config/data preservation.

### Changed

- Root `--version` now emits only raw SemVer. Help depth accepts `max` or an
  integer greater than one, and global flags repeat only with
  `--help-tree-global-flags`.
- Agent resource commands use `upgrade`; the prohibited `update` name is gone.
- Install and uninstall scripts use the shared `$HOME/.guiho` layout, explicit
  wiki parameters, immutable version directories, complete checksums, and
  manifest-derived state.

## [0.1.1] - 2026-08-07

This release makes GitHub Release publication resilient to delayed events for a
tag that has since been replaced with a newer target.

### Fixed

- The publish workflow now detects when its event commit does not match the
  tag's current peeled target and exits successfully without attempting a
  stale release build or reporting a false publication failure.

## [0.1.0] - 2026-08-06

This release improves capture idempotency and makes the installers more
reliable across shell and PowerShell environments.

### Fixed

- Capture normalizes the supplied text before calculating its digest and before
  appending its generated source footnote, so equivalent input is stored and
  compared consistently.
- `capture --replace` rewrites a target when its recorded input digest differs
  from the new input, including documents that have no prior Buda metadata.
- Shell and PowerShell installers verify that the installed Buda binary is
  reachable on `PATH`, probe supported qmd locations when qmd is not already on
  `PATH`, and use configurable agent-skill destination directories.
- Optional skill registration occurs only after a successful binary install and
  cannot turn that successful install into a failure.

## [0.0.2] - 2026-07-27

This release completes Buda's first installable GitHub Release workflow and
hardens the external qmd runtime boundary without introducing a package-manager
publication.

### Added

- Linux, macOS, and Windows installation instructions for latest and exact
  canonical `buda/v<version>` releases, plus an explicit 0.0.1-to-0.0.2 upgrade
  path.
- Local release-asset inputs for deterministic installer validation without
  weakening checksum verification or normal GitHub Release downloads.
- CI and tag-driven GitHub Release validation for the pure-Go CLI and exact
  eleven-artifact contract.
- Root `upgrade`, `upgrade check`, `upgrade list`, `upgrade rollback`, and
  `uninstall` routes with deterministic text and one-document JSON output.
- Checksum-verified self-upgrades with embedded-target asset selection,
  progress reporting, safe prior-backup rotation, Unix atomic replacement,
  detached Windows replacement/removal helpers, final executable verification,
  agent-resource reconciliation, and automatic rollback on failure.

### Changed

- Root JSON output now reports whether an explicit wiki was selected, and an
  argument-free invocation preserves the shared GUIHO greeting while never
  selecting a wiki implicitly.
- Shell and PowerShell installers now validate the canonical `buda/v<semver>`
  tag and compare an installed binary against the extracted SemVer. This fixes
  exact-version verification for tags such as `buda/v0.0.2`.
- The PowerShell installer prefers npm's `qmd.cmd` application launcher instead
  of the policy-sensitive `qmd.ps1` shim, and a restricted user-PATH update is
  now an actionable warning after a verified install rather than a rollback.
- Runtime guidance pins the tested external qmd 2.5.3 package while retaining
  Buda's supported `>=2.5.0,<3.0.0` compatibility range.
- Re-running the exact installer is the documented transactional upgrade from
  0.0.1 to 0.0.2; canonical wiki content remains untouched.
- An implicit upgrade to the already-installed latest version is a no-op, and
  upgrade, rollback, uninstall, and hidden helper routes are excluded from
  repository selection and background resource reconciliation.

### Distribution

- Buda release tags use `buda/v<version>` and publish only to GitHub Releases.
- Buda is not published to npm or another package registry. npm may be used by
  users to install the separate upstream `@tobilu/qmd` runtime dependency.
- The release payload remains exactly the eight binaries, embedded skill ZIP,
  instruction Markdown, and checksum manifest defined below.

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
