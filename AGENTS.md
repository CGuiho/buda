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

#### &copy; 2026 [GUIHO](https://guiho.co) as represented by [Cristóvão GUIHO](https://guiho.co/cguiho) All Rights Reserved.

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
older exact-eleven-asset and `update` guidance. GUIHO Agent Artifacts
Convention 0002 governs every bundled artifact name and metadata contract.
Buda's confirmed resource names are CLI home `buda`, main skill
`guiho-s-buda`, setup prompt `guiho-p-buda`, lifecycle prompts
`guiho-p-buda-install` and `guiho-p-buda-uninstall`, and managed instruction
`guiho-i-buda`; no agent definitions are shipped. The canonical repository is
`https://github.com/CGuiho/buda` and issues are created at
`https://github.com/CGuiho/buda/issues/new`.

Installer scripts are global-only: they install and verify Buda and its global
skill without accepting a wiki, invoking qmd, or running `buda init`. Wiki
creation and reconciliation occur only through a later explicit
`buda init --wiki <path>` operation.

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
## GUIHO Mirror Instruction Block

Run plain `mirror` once in a repository to verify the global Mirror skill and
this bounded instruction block. Repeated runs are idempotent.

Use `mirror version plan <target>` and `mirror version apply <target>` for semantic versioning.
`mirror init` defaults to `v{version}` tags and enables release commits and
pushes; explicit interactive or flag selections remain authoritative.

When `mirror.yaml` defines hooks, follow AI instructions only at the
agent-controlled everything, plan, and apply boundaries. Treat command hooks as
repository code: pass `--run-hooks` or `--skip-hooks` only with explicit
authorization, independently of `--yes`.
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

<!-- BEGIN GUIHO MIRROR - DO NOT EDIT THIS SECTION -->
## Semantic Project Versioning -- GUIHO Mirror

Invoke the guiho-s-mirror agent skill every time the user wants to bump, tag, release, plan, initialize, configure, or troubleshoot semantic project versioning with GUIHO Mirror.

Before editing release docs or changelogs, inspect mirror.config.toml. If [agents].write_changelog is false, skip changelog edits. If it is missing or true, changelog edits are allowed when the project has a changelog.

Use [agents].changelog_path as the changelog file path. If it is missing, use CHANGELOG.md in the project root.
<!-- END GUIHO MIRROR -->
## Mandume

GUIHO Buda.

Managed by the GUIHO Mandume swarm ([CGuiho/mandume](https://github.com/CGuiho/mandume)); the full worker-registry example lives at `example/AGENTS.md` there.

### Mode

```yaml
execution: dnd  # dnd | interruptible — orchestrator NEVER stops during execution/review
notifications: off  # off | on
harness: opencode  # only harness, YOLO, full permission
tmux-session: buda  # orchestrator session on su-57; convention = this project's name
```

### Coordination

- GitHub repository: https://github.com/CGuiho/buda.git
- GitHub Project: pending — CG binds one per project; it is the source of truth for new items (`todo.md` mirrors executable state; every item carries its issue URL)
- To-do file: `todo.md` (repo root)
- Reserved port: pending — reserve in `apps.md` (`CGuiho/guiho`)

### Workers

| Worker       | Class      | Model (opencode ID, OpenCode Zen)                                                                                     | Thinking | Usage        |
| ------------ | ---------- | --------------------------------------------------------------------------------------------------------------------- | -------- | ------------ |
| `mastermind` | mastermind | Muse Spark 1.3 Contributor (`opencode/muse-spark-1.3-contributor-free`) — unavailable until fixed; fallback `opencode/glm-5.3-flash` | max      | api-always   |
| `engineer`   | workhorse  | DeepSeek V4.1 Flash (`opencode/deepseek-v4-flash`)                                                                    | max      | api-always   |
| `engineer`   | workhorse  | GLM 5.3 Flash (`opencode/glm-5.3-flash`)                                                                              | max      | api-always   |

### Contract

- The orchestrator is pure orchestration on `main`, always working, always ready to answer CG; subagents are the workers above, called with full permission via `guiho-s-0440-hand-off`.
- Never stop during execution/review: questions are answered with the safest reversible choice and ledgered under `docs/questions/`. Questions to CG only when CG is present and available, or during brainstorming.
- Use the Mandume skills (`guiho-s-mandume` + lifecycle skills) and the Essentials skills (`guiho-s-0001-guiho`, `guiho-s-0004-working-with-cg`, `guiho-s-0040-explorer`, `guiho-s-0032-git-commit`). Conventions: `conventions/` in `CGuiho/guiho` (`apps.md` for ports).

