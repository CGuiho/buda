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
- Owned coherent main commits only; new independent review precedes delivery.
  The later reviewer has express parent plain-push authority after fresh full
  outgoing-ancestry review. Host installation has no Git commit claim.

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
- Both issues remain OPEN/Testing, with matching local mirrors. The original
  implementation handoff held commits for independent review and had no child
  push authority. The later independent acceptance and explicitly authorized
  delivery are recorded below; release, whole-family readiness and human
  acceptance remain separate.

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

## Independent Review — 2026-10-04

The bounded installation, policy and cited lexical capability are independently
accepted. Fresh verification confirms the exact launcher/payload/config/pointer
hashes above, all 26 checked-source manifest artifacts and canonical copies,
archive/source parity, **41/41 upstream qmd files**, and **16/16 projection
files** (eight in each supported destination). Runtime and embedded resources
remain identical to the checked build commit; review corrections touch policy
and task prose only. All 24 official `buda/v0.2.1` manifest assets were rechecked,
but that retired-skill release remains uninstalled. Installation-delta
recomputation confirms only expected additions and no monitored pre-existing
skill/bin file change or removal. No host installation was changed by review.

New independent fixture:
`/tmp/opencode/buda-review-2026-10-04/isolated-wiki/wiki`, ID
`cguiho-buda-independent-review-2026-10-04`. The original implementation fixture
is preserved. The reviewer pre-sealed the actual README source, its full hash
and 5,457 bytes, predicted raw/source/work-item paths, canonical write allowlist
and derived-state subtrees before any writes. Network namespaces, fixture-local
cache and `BUDA_DISABLE_MAINTENANCE=1` bounded every Buda smoke command.

Init, initial no-match typed-title ingest, cited capture, lint, index, lexical
query, get and idempotent reinit exit 0. Raw evidence is byte-identical to the
source and mode `0444`; capture `concepts/reviewer-context.md` cites the original
source record and sealed bytes while retaining separate `capture-input`
attribution. Lexical query `repository Go Cobra` returns that concept, document
`#f10a42`, score `0.55`; get returns exactly its canonical bytes. Lint and reinit
have zero wiki deltas. Index/query/get/doctor can change derived SQLite state;
the result is not an all-filesystem purity claim. Protected global policy,
projections, installed artifacts, legacy index and source bytes remain identical
through every operation. Cache additions are confined to the fixture, no GGUF
file appears, and no instance-registry delta is observed.

### Measured general and repeated ingestion gap

The ordinary human title `Buda repository overview` does not complete on initial
ingest or either of two repeated ingests in the model-free network-isolated
fixture. Each reaches the reviewer's **45-second deadline** and is killed as a
process group: recorded return `-9` (SIGKILL), not a natural Buda error exit.
Buda-wrapped traces show qmd `query` at `Expanding query...`; captured foreground
stdout/stderr is empty because child output is buffered until completion.

Repeating the initial `lex: <full-source-sha256>` title also reaches that deadline
at `Reranking 4 chunks...`. This proves the initial no-match success is not a
repeatable model-free ingest recipe. All four bounded failures preserve
canonical files and the registered raw/source/work item; only derived SQLite
bookkeeping changes. Public `ingest --mode lexical` is rejected as an unknown
flag with exit 2 and zero wiki delta.

Source/API gap: `cmd/ingest.go:36` hardcodes `qmd.ModeHybrid`, and line 42 also
normalizes candidates as hybrid. The adapter already maps `ModeLexical` to qmd
`search` (`internal/qmd/adapter.go:275-284`). A separately owned follow-up needs
an explicit supported public ingest retrieval-mode selection, consistent
selected-mode evidence normalization, and ordinary-title initial/repeated
model-free integration proof through Buda. Preserve sole-qmd retrieval, existing
source seals/idempotence and honest full semantic doctor failure. This review
does not implement that feature or authorize model downloads.

### Doctor and remaining gates

Independent full doctor exits 1: canonical healthy/conformant, configuration
valid, agent resources current, pack reproducible, repository resolved; qmd is
degraded with 12 checks, four warnings and three operational failures. The
Buda-wrapped trace confirms three missing default semantic models, four active
documents needing embeddings and no vector table. No semantic-doctor weakening,
fake embeddings or direct-qmd repair was performed.

The public embedded prompt metadata and trimmed bodies match source exactly;
the instruction matches source after real wiki/bundle substitution and Buda's
actual managed markers. Two verifier assertions initially assumed raw-file/body
equality and another CLI's marker spelling; both were corrected outside the
repository. Only the missing resource inspections were resumed, without
replaying the completed wiki smoke or discarding failed-ingest evidence.

Both issues stay OPEN/Testing; existing Project/Component/Status and item
identities are rechecked read-only. Four disabled external policies still need
later human ratification. XDocs/RunX CLI checks and missing task-helper descriptor
registration remain explicit separately owned gaps; no grants, legacy index,
descriptors or catalog were changed. Parent-authorized plain source delivery
requires fresh complete outgoing-ancestry review and live ref equality; final
proof is in `/tmp/opencode/2026-10-04-readiness-buda-review-result.md` and JSON.

Independent raw evidence is under `/tmp/opencode/buda-review-2026-10-04/`:
`host.json`, `smoke.json`, `smoke-commands.json`, `tracking-pre-doc.json`, source
checks, protected hashes and the preserved fixture's command output/traces.
The result certifies only the bounded prerequisite, not general ingestion,
semantic readiness, family readiness or human acceptance.
