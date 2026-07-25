---
name: buda-repository-agent-instructions
purpose: Define engineering, documentation, validation, and release boundaries for Buda.
description: Repository-local rules for the Go/Cobra Buda CLI, OKF wiki operations, qmd integration, XDocs metadata, and safe Git delivery.
created: 2026-07-26
owner: buda-package
flags: []
tags:
  - agents
  - repository-instructions
  - go
  - cobra
keywords:
  - Buda
  - Open Knowledge Format
  - qmd
  - XDocs
---

# Buda repository instructions

## Product contract

Buda is B-U-D-A: a repository-agnostic Go/Cobra CLI and embedded agent-skill
collection for maintaining one explicitly selected AI-maintained wiki in an
OKF-compatible form. Every repository-facing command requires `--wiki <path>`.
Never add a default wiki, global wiki registry, corpus role, visibility policy,
cross-repository operation, federated search, publishing workflow, or implicit
repository selection.

Google's canonical Open Knowledge Format `SPEC.md` governs base portable-format
conformance. Karpathy's LLM Wiki supplies the operating pattern. Buda-specific
health requirements must remain distinguishable from base OKF conformance.

qmd is the sole indexing and retrieval engine. Buda may invoke qmd through the
thin process adapter, validate containment, and normalize evidence. It must not
implement search, embeddings, ranking, reranking, or a retrieval fallback.

## Engineering

- Follow `guiho-s-0035-cli-engineer-go` for every CLI change.
- Use one fresh Cobra tree; only `-h` and root `-v` are short aliases.
- Strictly decode `buda.yaml` with `go.yaml.in/yaml/v3` and
  `KnownFields(true)`, then perform semantic validation.
- Preserve unknown OKF concept metadata; config strictness does not apply to
  forward-compatible OKF frontmatter.
- Keep `main.go` thin and inject I/O, time, process, filesystem, and network
  boundaries where used.
- Canonical writes are validated, same-filesystem staged, and atomically
  replaced where supported.
- Successful repository commands may schedule only the documented hidden
  local agent-resource reconciler. It performs no qmd, network, or cross-wiki
  work and never changes foreground output or exit status.
- Release builds are pure Go with `CGO_ENABLED=0` and the standard eight-binary,
  eleven-artifact contract.

## Documentation

Follow `guiho-s-xdocs` for every changed module. `xdocs.yaml` uses `ai.mode:
auto`, so update the owning named `*.xdocs.md` descriptor and companion-document
metadata in the same work unit. Keep parent/children links synchronized.

The approved implementation authority is GUIHO RFC 0002 at GUIHO commit
`30698669a2e72f1ded575574b5b8ff7f0b9b5c6e`. Preserve the canonical OKF
`SPEC.md`, official Google Cloud OKF announcement, Karpathy LLM Wiki gist, and
the brief accurate José António Ernesto naming note in Buda documentation.

## Validation and delivery

Run `gofmt`, `go mod tidy`, `go test ./...`, `go vet ./...`, focused XDocs
validation, and proportionate cross-build/release-contract checks. A foreign
cross-build is build-only unless executed on its native platform.

Implementation commits and pushes to this repository are allowed when cohesive
steps are complete. Do not publish a release, create a tag, apply a Mirror
version bump, open a pull request, or claim unimplemented automation without
separate user authorization.

<!-- BEGIN MIRROR — DO NOT EDIT THIS SECTION -->
## Semantic Project Versioning -- GUIHO Mirror

Invoke the guiho-s-mirror agent skill every time the user wants to bump, tag, release, plan, initialize, configure, or troubleshoot semantic project versioning with GUIHO Mirror.

Before editing release docs or changelogs, inspect mirror.yaml. If agents.write_changelog is false, skip changelog edits. If it is missing or true, changelog edits are allowed when the project has a changelog.

Use [agents].changelog_path as the changelog file path. If it is missing, use CHANGELOG.md in the project root.
<!-- END MIRROR -->
