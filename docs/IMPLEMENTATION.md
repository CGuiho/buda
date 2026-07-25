---
name: buda-implementation-plan
purpose: Record the approved vertical MVP implementation and validation plan.
description: File/module ownership, phase dependencies, acceptance criteria, validation commands, and release boundaries for the first Buda implementation.
created: 2026-07-26
owner: buda-docs
flags: []
tags:
  - implementation
  - validation
keywords:
  - Buda MVP
  - vertical slice
  - acceptance criteria
---

# Buda vertical MVP implementation

This plan implements the approved GUIHO RFC 0002 revision without versioning,
tagging, releasing, publishing, or opening a pull request.

## Phases and ownership

1. Foundation: Go module, thin entrypoint, fresh Cobra tree, live help tree and
   Markdown generation, strict output/exit behavior, and embedded-resource
   plumbing.
2. Repository and OKF: explicit wiki resolution, contained strict configuration,
   forward-compatible OKF parsing, initialization, provenance, deterministic
   health checks, capture, and ingest staging.
3. qmd integration: project-local qmd initialization, version/capability checks,
   collection containment, index/embed delegation, lexical/semantic/hybrid
   query, get, status, and doctor normalization. No retrieval fallback exists.
4. Agent behavior: embedded `guiho-s-0002-buda`, prompt/instruction resources,
   transactional global/local skill targets, bounded repository instructions,
   and the failure-isolated first-success reconciler.
5. Portability: deterministic canonical-bundle pack plus the eight pure-Go
   executable and eleven-artifact build contract.
6. Documentation and validation: XDocs metadata for every module, public usage
   and governing references, format/tidy/test/vet, XDocs checks, native smoke,
   and foreign build-only verification.

Dependencies flow from foundation and repository resolution into domain and qmd
commands. Agent maintenance depends only on resolved repository paths and
embedded bytes; it must not call qmd or the network. Packaging depends only on
the canonical bundle.

## MVP acceptance

- `init`, `capture`, `ingest`, `lint`, `index`, `query`, `get`, `status`,
  `pack`, and `doctor` exist in one Cobra tree.
- Every repository command fails without an explicit `--wiki` and reports the
  resolved absolute wiki in normal command output.
- Strict config rejects unknown fields and escaping paths; OKF parsing preserves
  unknown extension keys.
- qmd is the only indexing/retrieval implementation and every applicable
  invocation is explicitly collection-scoped in a project-local qmd index.
- Text is deterministic and JSON mode emits one document without progress or
  ANSI output.
- Base OKF conformance and Buda health are separate results.
- Agent resources and instruction blocks are idempotent, atomic, bounded, and
  tested for both global/local tool destinations and AGENTS/CLAUDE precedence.
- Packs contain only canonical bundle material plus deterministic manifest and
  checksums; they do not claim sharing safety or publish anything.
- The release definition contains exactly the approved eleven artifact names.

## Validation commands

```powershell
gofmt -w (rg --files -g '*.go')
go mod tidy
go test ./...
go vet ./...
go run . --help-tree
go run . --help-docs
xdocs meta . --documents --strict
xdocs tree
xdocs doctor .
go run ./devops/build-binaries.go --version 0.0.0-dev `
  --commit local --build-date 2026-07-26T00:00:00Z
```

Foreign binaries are build-only on this Windows host. qmd integration is
runtime-validated only when a supported qmd executable is installed; injected
process tests and structured fixtures do not substitute for that smoke test.
