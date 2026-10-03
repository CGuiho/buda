---
name: Buda Local Wiki Prerequisites
purpose: Define the bounded local install and lexical wiki verification for Buda issue 11
description: Authorized installation, strict nonsecret configuration, exact global projections and isolated provenance-preserving wiki smoke requirements.
created: 2026-10-03T21:52:37Z
flags: [active]
tags: [readiness, prerequisites, buda, qmd]
keywords: [Bun, lexical, source-seal, global-policy, issue-11]
---

#### &copy; 2026 [GUIHO](https://guiho.co) as represented by [Cristóvão GUIHO](https://guiho.co/cguiho) All Rights Reserved.

# Buda Local Wiki Prerequisites

## Todo Index

- Index: [TODO.md](../../TODO.md), current task 2.
- GitHub project item: [Buda issue #11](https://github.com/CGuiho/buda/issues/11).
- GitHub component: `buda`, [GUIHO Project #2](https://github.com/users/CGuiho/projects/2).
- Status: testing (issue OPEN; Project #2 / Component `buda` / Status Testing freshly read back).
- Policy companion: [native background policy](native-background-policy.md), issue #10.

## Outcome and Authorized Scope

The parent expressly authorizes one serialized host owner to install a verified
official Buda release or checked source build, plus supported pinned upstream
qmd `>=2.5.0 <3.0.0` through Bun. No npm/Node invocation is authorized. Source
changes to Buda behavior are not presumed. Actual runtime/package incompatibility
must be recorded precisely while independently valid policy work proceeds.

For a fresh global Buda configuration, all four external evolution actions
(`upgrade`, `issues.bugs`, `issues.improvements`, `issues.reviews`) may be set to
`disabled` under the actual strict schema. This is a task-scoped reversible
default pending later human ratification, not permanent CG policy. Existing
values must survive. Tracking issue #11 has separate explicit parent creation
authority; these disabled settings do not authorize additional feedback issues.

Only exact source-declared Buda global skill projections may be reconciled;
list/hash before and after, preserve extras and other resources. Future per-wiki
workers must find these projections byte-identical/current or return to this
host owner. No agent/model/provider/permission/auth configuration edits.

## Acceptance Signals

- Installed executable/package identity, full hashes and verified provenance.
- Exact global configuration path/values and required projections recorded.
- One explicitly selected isolated `/tmp/opencode` wiki initialized with a
  truthful new ID; one explicit nonsecret source full-SHA256-sealed before ingest.
- Cited capture, Buda lint/index/doctor, lexical Buda query and retrieval evidence.
  All qmd operations are invoked by Buda, never directly or through a fallback.
- Filesystem deltas, immutable evidence equality, canonical versus derived
  artifacts and deterministic pre-write sealing guidance documented.
- Both issues OPEN/Testing after review checks; local mirrors match readback.
- Owned coherent main commits only; parent performs new independent review and
  delivery. Host installation has no Git commit claim.

## Decisions and Boundaries

Mode: dnd. Parent waived the broad planning lifecycle only for this narrowly
assigned prerequisite. No recursive workers, CG questions, private source,
environment files, keys, authentication downloads, semantic-model downloads,
production, releases, Mirror bumps, new repositories/branches or history rewrite.
Semantic readiness is outside the lexical acceptance signal; do not require a
provider or model unless actual lexical operation proves it necessary.

The actual full GUIHO conventions 0011/0002/0007/CLI 0001 and owning source/schema
are mandatory inputs. Current unsafe XDocs/RunX automatic resource behavior is
separately owned; preserve the legacy index and report skipped validation.

## Installed Identity and Reversible Decisions

The checked source build is from
`8ed5b0ac9675ed10338f214df45f0d16de4dd8b0`, built with Go 1.27.1 and
`CGO_ENABLED=0`, installed through the repository's `devops/install.sh` after
full manifest/checksum and skill archive/source byte-parity verification. Its
raw version is `0.2.0` and hidden self-test returns `ok`. That identifier retains
the source's embedded resource version: it is **not an official release build**,
repository version bump, or publication. The verified official `buda/v0.2.1`
manifest still declares retired `guiho-s-0002-buda` resources, so it was not
installed over the current source-owned `guiho-s-buda` contract.

- Launcher: `/root/.guiho/bin/buda`, SHA-256
  `06d24ad2a94bbe714a80186533a4b0712e2d939f5b9964e8315865f4aa23ab95`.
- Payload: `/root/.guiho/buda/versions/0.2.0/buda`, SHA-256
  `3bfe87ea7fc4e3869525a2db3168cebf7c774eaeb3a86ea1e13e85c1ce78bc37`.
- Upstream `@tobilu/qmd@2.5.3`: installed with Bun 1.4.2 and lifecycle scripts
  disabled; registry tarball SHA-512 integrity verified and all 41 installed
  package files byte-identical. Actual package location is
  `/root/.bun/install/global/node_modules/@tobilu/qmd`.
- `/root/.local/bin/qmd` invokes that exact packaged `dist/cli/qmd.js` through
  Bun. The upstream `bin/qmd` entry routes through Node, so the task-owned
  launcher preserves the no-Node constraint without modifying package bytes.
- Fresh `/root/.guiho/buda/buda.global.yaml` uses that executable and all four
  external evolution policies `disabled`; later human ratification is pending.
  Existing values were preserved. No host agent/provider/model/permission setting
  was changed.
- `/root/.agents/skills/guiho-s-buda` and
  `/root/.claude/skills/guiho-s-buda` each contain exactly the eight
  source-declared files, with tree digest
  `sha256:49681908e373e4c2df0686b9385c4b18cb9ce58c9c70e4cc41ed90e44ee83ec8`.
  Pre-existing skill/bin files remained byte-identical; no extras were removed.

The installer adds the user-level PATH entry. A login shell resolves Buda;
existing harness shells should use the absolute launcher path or an explicit
PATH prefix. Package engine metadata declares Node >=22; upstream documents
Bun support, and this exact Bun/package combination passed the bounded smoke.
That evidence does not certify every upstream semantic/native capability.

## Isolated Smoke and Filesystem Evidence

Selected wiki: `/tmp/opencode/buda-readiness-smoke-2026-10-03/wiki`.
Wiki ID: `cguiho-buda-readiness-smoke-2026-10-03`. Every Buda command ran inside
`unshare --net` with fixture-local `XDG_CACHE_HOME` and
`BUDA_DISABLE_MAINTENANCE=1`. All qmd operations were invoked by Buda.

The original `/root/swe/buda/README.md` was sealed **before init/ingest** with
full SHA-256
`6015c122073583c49c57cb5ac71c01343671efea83afb116819b6a68d03d85ff`.
Immutable `knowledge/references/raw/<full-sha256>.source` matches its 5,457
bytes exactly; `knowledge/sources/source-6015c122073583c4.md` retains original
source provenance. Cited capture `knowledge/concepts/repository-context.md`
links both artifacts and states the checked README facts. Capture metadata
truthfully records separate `capture-input` provenance, not original-source
provenance impersonation.

Init, ingest, capture, lint, index, lexical query, get and idempotent reinit
each exited 0. Query `repository Go Cobra` with `--mode lexical` returned
`concepts/repository-context.md`, document `#a65654`, score `0.55`; get bytes
equal the canonical document exactly. Canonical writes matched the pre-sealed
allowlist. Derived qmd/ingest state stayed within `.qmd/` and `.buda/`; runtime
Bun/Mesa cache writes stayed in the fixture cache. Reinit had zero wiki deltas.
Global policy/projections and original source bytes remained unchanged through
the commands; no GGUF model file was created.

For each later authorized wiki, seal its real ID, source path/full hash/bytes,
predicted raw/source/ingest artifact paths, capture target, exact canonical
write allowlist and derived-state subtrees **before writes**. Recheck source
and raw-byte equality after ingest, preserve cited capture-input attribution,
then lint/index/query/get through Buda and compare filesystem deltas. Do not
copy this fixture's wiki identity or hash into a different repository.

## Remaining Capability Gaps and Review Gate

- `buda doctor` exited 1. Canonical conformance/health, strict configuration,
  agent resources and pack reproducibility passed; qmd semantic readiness is
  degraded. A process trace around **Buda**, not direct qmd invocation, records
  three missing models, four active documents needing embeddings and no vector
  table. Semantic-model downloads remain outside this task.
- Ingest currently hardcodes `qmd.ModeHybrid`. This initial no-match fixture
  passed a typed `lex: <full-source-sha256>` title through Buda; upstream qmd
  avoided model expansion/reranking in that bounded case. This does **not**
  verify model-free general or repeated ingestion; do not use it as blanket
  family-setup acceptance. No Buda runtime source was changed.
- XDocs/RunX CLI validation is skipped pending independent acceptance of their
  safe runtime boundaries. The legacy `XDOCS.md` remains intact. No descriptor
  grant or named-descriptor authorization exists for these new task helpers,
  so no descriptor or configuration change was inferred.
- Both issues remain OPEN/Testing, with matching local mirrors. Independent
  parent review and explicit child-push authority are still required; no push,
  release, whole-family readiness or human acceptance is claimed.

## Evidence and Handoff

- `/tmp/opencode/buda-installed-identity.json`: full installed identities,
  exact build command, package provenance and runtime versions.
- `/tmp/opencode/buda-global-deltas.json` and
  `/tmp/opencode/qmd-installed-parity.json`: scoped global additions and parity.
- `/tmp/opencode/buda-smoke-evidence.json` and the preserved fixture:
  exact commands/exits, pre-write seal, per-command filesystem snapshots,
  retrieval and provenance evidence; doctor trace is fixture-local.
- `/tmp/opencode/buda-source-checks.json` and `buda-go-tests.log`:
  successful Go tests, vet, tidy-diff, formatting and whitespace checks.
- `/tmp/opencode/buda-read-testing-task-readbacks.json`: fresh read-only remote
  OPEN/Testing/Project/Component verification for issues #10 and #11.
- Return packet: `/tmp/opencode/2026-10-03-readiness-buda-prerequisite-result.md`
  and matching JSON, for new independent parent review.
