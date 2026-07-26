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
   and failure-isolated first-success reconciliation split between bare
   global-only bootstrap and explicit-wiki instruction bootstrap.
5. Portability: deterministic canonical-bundle pack plus the eight pure-Go
   executable and eleven-artifact build contract.
6. Documentation and validation: XDocs metadata for every module, public usage
   and governing references, format/tidy/test/vet, XDocs checks, native smoke,
   and foreign build-only verification.

Dependencies flow from foundation and repository resolution into domain and qmd
commands. Bare agent maintenance depends only on embedded bytes and global skill
destinations. Wiki instruction maintenance additionally depends on the absolute
repository path already resolved from explicit `--wiki`; it must not discover
repositories, cross repository boundaries, call qmd, or use the network.
Packaging depends only on the canonical bundle.

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
- Bare `buda` success may reconcile only the two global skill destinations;
  repository instructions are reconciled only when `buda --wiki <path>` or a
  repository command supplies and resolves an explicit wiki.
- The managed instructions identify the selected wiki and bundle, declare Buda
  the required maintenance and retrieval tool, require explicit `--wiki` on
  every repository operation, require evidence evaluation and concept-path plus
  source-ID or resource citations, prohibit bypassing Buda or invoking qmd
  directly, and state that instructions are behavior rather than access
  control.
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
