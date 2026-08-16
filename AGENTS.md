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
- Release builds are pure Go with `CGO_ENABLED=0` and the complete
  manifest-derived payload, launcher, resource, schema, example, and checksum
  contract defined by GUIHO CLI Convention 0001. ARMv6/ARMv7 cross-builds are
  build-only unless executed on native hosts.

## Convention 0001 authority

GUIHO CLI Convention 0001 is the current public CLI authority. It supersedes
older exact-eleven-asset and `update` guidance. Buda's confirmed resource
names are CLI home `buda`, main skill `guiho-s-0002-buda`, setup prompt
`guiho-p-buda`, and managed instruction `guiho-i-buda`; no agent definitions
are shipped. The canonical repository is
`https://github.com/CGuiho/buda` and issues are created at
`https://github.com/CGuiho/buda/issues/new`.

The active implementation plan is
`docs/GUIHO_CLI_CONVENTION_0001_IMPLEMENTATION_PLAN.md`; the audit and
acceptance matrix are its evidence sources. Lifecycle code must use the stable
launcher, immutable version directories, `current.json`, `artifacts.json`,
strict checksums, and manifest-driven ownership. Use `upgrade` in every agent
resource command tree; `update` is not a supported alias.

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
---
name: guiho-i-mirror
description: Mirror agent instruction block.
purpose: Provide the canonical managed project instruction for Mirror versioning.
created: 2026-07-18
owner: mirror-embed-prompts
flags: []
tags: [mirror, instruction, agents]
keywords: [version plan, version apply]
---

# GUIHO Mirror Instruction Block

Run plain `mirror` once in a repository to verify the global Mirror skill and
this bounded instruction block. Repeated runs are idempotent.

Use `mirror version plan <target>` and `mirror version apply <target>` for semantic versioning.
`mirror init` defaults to `v{version}` tags and enables release commits and
pushes; explicit interactive or flag selections remain authoritative.
<!-- END MIRROR -->

<!-- BEGIN RUNX — DO NOT EDIT THIS SECTION -->
## RunX Command Catalog

Load the `guiho-s-runx` skill whenever discovering commands, creating or
updating catalog entries, validating `runx.yaml`, inspecting command details,
or executing RunX commands.
Start with `runx check --format json` and `runx list --format json`, select
stable UIDs, use `runx describe <uid>`, and run
`runx run --dry-run <uid>` before unfamiliar or side-effecting work.
RunX options precede the selector; post-selector tokens belong to the child.
<!-- END RUNX -->
