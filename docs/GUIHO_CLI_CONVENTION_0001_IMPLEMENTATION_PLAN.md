---
name: buda-guiho-cli-convention-0001-implementation-plan
purpose: Define the complete dependency-ordered implementation program for bringing Buda into compliance with GUIHO CLI Convention 0001.
description: Proposed breaking-change plan covering governance, tooling, Cobra behavior, configuration, agent resources, installation layout, release manifests, launcher and payload architecture, init, upgrade, uninstall, migration, validation, and release preparation.
created: 2026-08-16
owner: buda-docs
flags: []
tags:
  - implementation
  - plan
  - cli
  - convention
keywords:
  - Buda
  - GUIHO CLI Convention 0001
  - breaking change
  - implementation units
  - stable launcher
  - artifacts.json
  - agent.evolution
  - minor release
---

# Buda GUIHO CLI Convention 0001 implementation plan

## Plan status

**Status: complete and released.**

This plan converts the findings in
[`GUIHO_CLI_CONVENTION_0001_COMPLIANCE_AUDIT.md`](GUIHO_CLI_CONVENTION_0001_COMPLIANCE_AUDIT.md)
into independently reviewable, dependency-ordered implementation units. The
trace from every audit finding to its implementation and verification evidence
is recorded in
[`GUIHO_CLI_CONVENTION_0001_ACCEPTANCE_MATRIX.md`](GUIHO_CLI_CONVENTION_0001_ACCEPTANCE_MATRIX.md).
Repository progress is tracked in [`../TODO.md`](../TODO.md).

The work is intentionally breaking in behavior and installation layout. The
user selected the next minor release target **Buda 0.2.0 from 0.1.1**; the
authorized Mirror minor plan applied `0.2.0`, the canonical tag `buda/v0.2.0`
was created and pushed after implementation, exact-head review, and all
validation gates passed, and the published GitHub Release carries the
manifest-derived asset set.

## Governing authority and conflict resolution

Apply authority in this order for this program:

1. the user's explicit request to adopt GUIHO CLI Convention 0001, including
   acceptance of breaking changes;
2. `C:\GUIHO\guiho\docs\conventions\guiho-convention-0001-cli.md` at
   convention commit `bd7dc4583fbc0722d3f05b78f7c70c7a970865e2`;
3. Buda's product contract in `AGENTS.md` and approved GUIHO RFC 0002 at commit
   `30698669a2e72f1ded575574b5b8ff7f0b9b5c6e`;
4. current repository architecture and tests where they do not conflict with
   the convention; and
5. the older `guiho-s-0035-cli-engineer-go` contract only where it remains
   compatible.

The older CLI skill and Buda instructions currently prescribe `update`, one
mutable executable, and exactly eleven release assets. Convention 0001 is
newer and explicitly replaces those rules with `upgrade`, a stable launcher,
immutable payloads, schemas, examples, manifests, and a complete release unit.
Implementation unit U00 updates repository-local authority and records the
required upstream GUIHO skill amendment. No executor may preserve the old
11-asset count merely to satisfy stale instructions or tests.

## Product invariants that must not change

- Buda remains B-U-D-A, repository agnostic and use-case agnostic.
- Every repository-facing action receives one explicit `--wiki <path>` or the
  equivalent full-name lifecycle-script parameter. There is no implicit
  current-directory selection, ancestor search, default wiki, or global wiki
  registry.
- Buda operates on one selected wiki per invocation and never federates,
  publishes, or copies knowledge across repositories.
- Google OKF remains the portable canonical format. Buda-specific health stays
  distinct from base OKF conformance.
- qmd remains the sole index and retrieval engine. No search, embedding,
  ranking, reranking, or fallback implementation is added.
- Canonical OKF content and raw evidence are user-owned repository data. They
  are never installation artifacts and are never deleted by Buda uninstall.
- `main.go` remains thin; Cobra remains the sole public command router; owned
  structured input remains strictly decoded and semantically validated.
- Release builds remain pure Go and preserve the eight payload targets,
  including Linux ARMv6, ARMv7, and ARM64. A cross-build is never called a
  runtime smoke.
- No implementation unit deploys or promotes production state, changes Cloud
  Run, DNS, databases, secrets, traffic, or containers.

## Accepted naming decisions

The convention-required naming and release decisions were explicitly accepted
by the implementation delegation and are recorded here as durable authority.
No naming decision remains open for implementation or review:

| Decision | Accepted value | Recorded in |
|---|---|---|
| CLI home directory | `buda` under `$HOME/.guiho/buda/` | U00/U05 and `TODO.md` |
| Main skill | retain `guiho-s-0002-buda` | U00/U04 and the bundled skill |
| Main setup prompt ID | `guiho-p-buda` | U00/U04 and the bundled prompt |
| Agent definitions | none | U00/U04/U07 and the release manifest |
| Canonical GitHub repository | `https://github.com/CGuiho/buda` | U00/U04 and the skill |
| Canonical issue creation URL | `https://github.com/CGuiho/buda/issues/new` | U00/U04 and the skill |
| Release line | next minor `0.2.0` from `0.1.1`, applied only by Mirror after review | U00/U14 and `CHANGELOG.md` |

The existing product and released skill provide strong evidence for `buda` and
`guiho-s-0002-buda`, but the convention calls for explicit confirmation. The
setup prompt ID has no prior confirmation and is therefore a hard U04 gate.

## Delivery topology

Implementation uses a protected integration line so `main` never contains a
half-migrated lifecycle:

1. After the planning documents are accepted and committed, create
   `codex/cli-convention-0001` from the accepted `main` commit.
2. Create one unit branch named
   `codex/buda-convention-0001-uNN-<slug>` from the current integration head.
3. Open each unit PR against the integration branch. Require exact-head review,
   unit validation, and green CI before merge.
4. Rebase the next unit on the new integration head; do not implement multiple
   units against stale architecture assumptions.
5. Keep the tag-triggered publication workflow inactive: no tag exists on the
   integration branch.
6. After U13 passes, open one aggregate PR from the integration branch to
   `main`. Validate the exact aggregate head again.
7. U14 prepares, but does not apply, the Mirror minor plan. Applying the bump,
   pushing the release commit/tag, and publishing the GitHub Release require
   separate explicit release authorization.

Every unit owns its source tests, RunX catalog changes, XDocs descriptor
changes, and affected durable documentation. A unit is not mergeable when its
behavior exists only in code or CI and is absent from RunX/XDocs.

## Target installation and ownership architecture

### Canonical filesystem layout

```text
$HOME/.guiho/
├── bin/
│   └── buda[.exe]                       stable launcher
├── .temp/
│   └── buda-<operation>-<unique-id>/    one validated operation workspace
└── buda/
    ├── buda.global.yaml                 persistent global configuration
    ├── current.json                     active and previous relative pointers
    ├── installed-artifacts.json         installed ownership manifest
    ├── cache.json                       disposable update cache
    ├── state/
    │   ├── upgrade.lock                 ownership-token lock
    │   ├── transaction.json             recoverable lifecycle journal
    │   └── instances/                   current-user payload instance records
    └── versions/
        └── <semver>/
            ├── buda[.exe]               immutable platform payload
            └── artifacts/
                ├── artifacts.json
                ├── skills/
                ├── prompts/
                ├── instructions/
                ├── agents/
                ├── schemas/
                ├── examples/
                └── metadata/
```

Only the platform launcher is canonical outside the CLI home. Every projection
to `.agents`, `.claude`, `AGENTS.md`, or another supported agent surface is a
manifest-owned copy whose canonical source remains inside the active version
directory.

### Ownership and preservation table

| Path or state | Owner/class | Reinstall/upgrade | Default uninstall | Preservation option |
|---|---|---|---|---|
| `$HOME/.guiho/` | shared GUIHO infrastructure | preserve | preserve | not removable by Buda |
| `$HOME/.guiho/bin/` and shared PATH entry | shared infrastructure | preserve container; repair Buda launcher only | preserve container and PATH; remove Buda launcher only | not removable by Buda |
| Buda stable launcher | Buda replaceable artifact | verify/repair transactionally | remove | none |
| `$HOME/.guiho/buda/versions/<version>/` | Buda immutable release artifact | add candidate; retain immediate rollback version; collect older inactive versions | remove | none |
| `current.json` and installed manifest | Buda replaceable state | replace atomically | remove | none |
| `buda.global.yaml` | Buda persistent config | preserve and migrate | remove | `--preserve-config` |
| selected wiki `buda.yaml` | Buda project config | preserve and migrate | remove only from explicit `--wiki` | `--preserve-config` |
| selected wiki canonical OKF bundle and raw evidence | user-owned repository data | never mutate as lifecycle work | **never remove** | always preserved; outside CLI ownership |
| selected wiki `.buda/` | Buda-derived disposable state | may recreate; never treat as canonical knowledge | remove only from explicit `--wiki` when ownership is proven | not preserved by `--preserve-data` because it is disposable |
| selected wiki `.qmd/` | qmd-owned derived state | leave to qmd/Buda domain initialization | preserve unless a separately validated Buda-owned child is recorded | outside Buda installation ownership |
| future Buda databases or user-created application data inside CLI home | Buda persistent data | preserve | remove | `--preserve-data` |
| cache, stale transaction metadata, old operation temp | Buda disposable state | replace/remove safely | remove | none |
| global skill/prompt/definition projections and bounded instruction | Buda manifest-owned projection | complete replacement; remove retired projections | remove | none |

This table is normative for implementation. No implementation may infer that
all content under a selected wiki is Buda-owned. In particular, uninstall must
not delete the OKF bundle merely because Buda maintains it.

### Explicit-wiki lifecycle rule

Convention 0001 refers to the “current project,” but Buda forbids implicit
repository selection. Therefore:

- `buda init`, agent instruction actions, agent-managed upgrade reconciliation,
  and `buda uninstall` project cleanup use explicit `--wiki <path>`;
- Bash lifecycle scripts use full-name `--wiki <path>`;
- PowerShell lifecycle scripts use `-Wiki <path>`;
- project projections/config are never discovered from the process working
  directory; and
- no global list of initialized wikis is created.

Both installers require the explicit wiki parameter and fail before any write
when it is absent. This is an intentional Buda-specific full-name parameter on
top of the convention selectors. It satisfies the convention's project
instruction/config reconciliation without violating Buda's prohibition on
implicit repository selection. The installer runs `buda init --wiki` after the
release is verified and active.

## Target release unit

The release builder stops using an exact asset count. It produces a manifest-
derived set containing, at minimum:

- eight immutable Buda payload binaries with the existing target names;
- eight platform-matched stable launchers, named
  `buda-launcher-<os>-<arch>[.exe]`;
- `guiho-s-0002-buda.zip`;
- the confirmed setup prompt artifact;
- `guiho-i-buda.md`;
- every confirmed agent definition, if any;
- `buda.schema.json` and `buda.global.schema.json`;
- complete `buda.example.yaml` and `buda.global.example.yaml`;
- `artifacts.json`; and
- `checksums.txt` covering every release artifact except itself.

`artifacts.json` declares release version, target compatibility, artifact ID,
relative release path, canonical installed path, archive members, SHA-256 for
every non-manifest artifact/member, projection destinations, ownership class,
and whether a path is replaceable, persistent, or disposable. The manifest's
own release-file digest is carried by `checksums.txt`, avoiding a self-hash
cycle. Tests derive the expected release set from this manifest.

## Cross-cutting implementation rules

- All path constructors live in one injected, platform-native layout package;
  tests use synthetic homes and must reject absolute/traversing manifest paths.
- Every lifecycle mutation uses a validated unique child of
  `$HOME/.guiho/.temp/`; recursive cleanup first proves strict containment.
- Candidate validation happens before activation and includes checksum,
  executable format, target, raw version, embedded-resource read, and hidden
  `__self-test`.
- `current.json`, installed manifests, config, journals, and bounded managed
  documents use same-filesystem staging, sync where meaningful, and atomic
  replacement.
- Upgrade/uninstall process inspection verifies current user and executable
  path; a filename alone is never authority to terminate a process.
- The launcher forwards original arguments/streams, waits for the payload, and
  returns its exact exit code. It performs no network work.
- The launcher reads only strictly decoded relative pointers under the versions
  root and falls back once to the previous verified payload when active launch
  fails.
- The CLI does not expose `install`. Agent skill/instruction trees use
  `upgrade`, with no hidden or deprecated `update` alias.
- Human output and JSON remain deterministic. Recovery data has text and
  structured representations, and raw `--version` bypasses notices.
- Every implementation unit updates its owning XDocs descriptors and RunX
  entries. No repeatable acceptance command is CI-only.
- Tests run sequentially on Windows when shared Go caches contend; isolated
  workspace caches are acceptable and must be ignored.

## Expected module and file ownership

The exact inventory may grow when tests expose a missing boundary, but these
owners are fixed so units do not duplicate lifecycle logic:

| Owner | Planned files/directories | Responsibility |
|---|---|---|
| U00 | `AGENTS.md`, planning/decision Markdown | Authority, confirmations, historical/current plan routing. |
| U01 | `go.mod`, `go.sum`, `runx.yaml`, `mirror.yaml`, `xdocs.yaml`, `XDOCS.md`, root and GitHub XDocs descriptors, CI tooling steps | Mandatory toolchain and continuous metadata enforcement. |
| U02 | `cmd/root.go`, `cmd/agent.go`, `cmd/doctor.go`, `internal/help/`, focused tests | Public version, flags, help, and command names. |
| U03 | `internal/config/`, `schemas/`, `examples/`, their embeds/descriptors/tests | Dual config, merge, policy, schemas, examples, schema URLs, migration. |
| U04 | `internal/agent/`, `skills/`, `prompts/`, optional `agents/`, resource tests/descriptors | Typed agent artifact family and managed projections. |
| U05 | new `internal/installlayout/`, `internal/artifact/`, and `internal/lifecycle/` packages | Canonical paths, pointers/manifests, ownership, staging, lock, journal, snapshots, rollback. |
| U06 | new `cmd/buda-launcher/`, `internal/launcher/`, and `internal/instance/` packages; hidden payload self-test | Stable dispatch, fallback, instance registration, verified process inspection. |
| U07 | `devops/build-binaries.go`, build tests/fixtures, release-resource assembly | Complete deterministic manifest-derived release unit. |
| U08 | `cmd/init.go`, new `internal/bootstrap/`, repository/agent/config integration tests | Ordered interactive common init plus Buda/qmd initialization. |
| U09 | `devops/install.sh`, `devops/install.ps1`, shared selector/legacy fixtures, installer tests | Fresh/reinstall/downgrade/channel/legacy installation and PATH. |
| U10 | `cmd/upgrade.go`, new `internal/releasecatalog/`, lifecycle/instance integration; retire superseded `internal/selfmanage` replacement code | Complete synchronous upgrade and recovery output. |
| U11 | `cmd/uninstall.go`, `devops/uninstall.sh`, `devops/uninstall.ps1`, new `internal/uninstall/`, parity fixtures | Shared uninstall plan, confirmation, preservation, deletion transaction. |
| U12 | `README.md`, `CHANGELOG.md`, `docs/`, skills/prompts, all affected XDocs descriptors | Public and durable implemented-truth documentation. |
| U13 | `.github/workflows/`, `devops/*_test.go`, cross-platform acceptance fixtures | Exact-head aggregate CI and publication gates. |
| U14 | `TODO.md`, changelog/version surfaces, Mirror plan evidence | Aggregate reconciliation and release preparation only. |

Each new project-owned directory receives exactly one descriptor in the unit
that creates it, is attached to the XDocs containment tree, and is added to the
release/RunX inventories when applicable.

## Implementation units

### U00 — Adopt the convention and seal names

**Depends on:** accepted planning documents.

**Goal:** remove governance ambiguity before code changes.

**Changes:**

- Record explicit confirmation for CLI home, main skill, setup prompt, agent
  definitions, repository URL, issue URL, and breaking release intent.
- Update `AGENTS.md` to replace the exact-11, `update`, mutable-binary guidance
  with Convention 0001 authority while preserving Buda product boundaries.
- Mark `docs/IMPLEMENTATION.md` as the historical 0.0.2 delivery record and
  make this document the active implementation plan.
- Record the external follow-up required to amend
  `guiho-s-0035-cli-engineer-go`; this checkout cannot edit the installed
  shared skill package. Buda's `AGENTS.md`, plan, and release handoff make the
  newer Convention 0001 precedence explicit, so implementation and review can
  proceed without silently preserving the obsolete contract. The shared-skill
  amendment remains a separately tracked governance follow-up, not an
  unresolved Buda naming decision.

**Verification:** strict XDocs metadata/tree/doctor, no implementation claims,
and a reviewer confirms all naming/authority decisions are explicit.

**Exit:** no unresolved naming decision affects file paths, artifact IDs, or
public commands.

### U01 — Establish Go, Mirror, RunX, XDocs, and CI baseline

**Depends on:** U00.

**Goal:** make mandatory project tooling continuously enforceable before the
large implementation begins.

**Changes:**

- Move `go.mod` and CI to the approved current Go 1.26 patch line; update module
  metadata only through `go mod tidy`.
- Add root `runx.yaml` with a stable `buda` namespace and catalog entries for
  format check/fix, tidy check/fix, tests, vet, native build, eight cross-builds,
  schemas, XDocs, Mirror, installers, upgrade/uninstall acceptance, artifact
  verification, and release preparation. Add new entries in later units before
  their commands are considered supported.
- Remove dead `package.json`/`jsr.json` Mirror paths and make Mirror check every
  actual version-bearing Buda surface that its schema supports. Add a separate
  parity check for version-bearing resources that Mirror cannot rewrite.
- Attach `.github` and `.github/workflows` to the single XDocs root tree; add
  missing root tooling files to `buda.xdocs.md`; refresh `XDOCS.md`.
- Add non-mutating CI gates for `mirror config check`, `runx check/list`, strict
  XDocs meta/tree/doctor, gofmt, tidy diff, tests, and vet.

**Verification:** all catalog commands dry-run, mandatory tool checks pass from
the repository root, `xdocs tree` includes `.github`, and no source behavior
changes.

**Exit:** every current repeatable operation is discoverable and CI rejects
tooling/documentation drift.

### U02 — Correct the public Cobra surface

**Depends on:** U01.

**Goal:** implement the convention's breaking command, version, and developer-
help contract without lifecycle changes.

**Changes:**

- Emit raw SemVer only from root `-v`/`--version`; bypass update notices and all
  surrounding output.
- Change `--help-docs` to render only the invoked command.
- Implement `--help-tree-depth` as `max | integer > 1`, defaulting visibly to
  `max`.
- Add presence-only `--help-tree-global-flags`; show inherited globals once by
  default and repeat them only when requested.
- Replace `agent skill update` and `agent instruction update` with `upgrade`.
  Do not keep aliases because the convention explicitly prohibits `update`.
- Update doctor guidance, examples, JSON command labels, tests, current scripts,
  and workflow assertions in the same unit so no internal caller uses the old
  contract.

**Verification:** exhaustive fresh-tree tests, both space/equals flag forms,
raw version byte equality, tree snapshots at root and nested scopes, rejection
of depth `1`, no forbidden aliases, and stable Markdown output.

**Exit:** audit findings B1-B5 and their obsolete assertions are closed.

### U03 — Implement dual configuration, schemas, examples, and evolution policy

**Depends on:** U01 and confirmed CLI home.

**Goal:** establish a strict, offline, versioned configuration foundation.

**Changes:**

- Split `internal/config` into strict global, project, merge, schema-reference,
  and migration responsibilities while retaining `KnownFields(true)` and
  semantic path validation.
- Add `buda.global.yaml` under the CLI home and keep `buda.yaml` in the explicit
  wiki root.
- Define global baseline fields and project override fields. `wiki_id` remains
  project-only and immutable; applicable qmd/default and agent policy fields
  merge per field.
- Add typed `agent.evolution` enums and defaults for upgrade, bugs,
  improvements, and reviews. Reject unknown values and Booleans.
- Add embedded and released `buda.schema.json`,
  `buda.global.schema.json`, and complete examples.
- Generate version-pinned GitHub Release schema comments from linker/release
  metadata, never from `main` or `latest`.
- Define a one-time read-only migration from legacy
  `$HOME/.guiho/buda/buda.yaml`: validate before mapping, never treat its
  `wiki_id` as global, preserve the legacy file until the new config commit
  succeeds, and report every migrated field.

**Verification:** Go type/schema structural parity, strict unknown-field and
multi-document tests, enum/inheritance/default tests, exact schema URL tests,
valid examples, offline validation, migration fixtures, and no secret fields.

**Exit:** both configs can be resolved, merged, validated, written atomically,
and round-tripped without network access or value loss.

### U04 — Complete the agent artifact family

**Depends on:** U00 naming confirmation, U02 command names, and U03 policy
contract.

**Goal:** provide confirmed, separately typed, versioned agent resources with
managed lifecycle guidance.

**Changes:**

- Retain the confirmed main skill and add the exact
  `## CLI Evolution and Feedback` section, canonical URLs, Buda-specific bug /
  improvement / review guidance, all policy semantics, upgrade-check flow,
  post-upgrade init/version checks, and issue URL reporting.
- Add the confirmed setup prompt, explaining Buda, remote installation with an
  explicit wiki, raw version verification, init, and upgrade.
- Keep `guiho-i-buda.md` typed only as the managed instruction; stop exposing it
  as prompt ID `buda`.
- Add confirmed agent definitions if selected; otherwise explicitly declare an
  empty definitions set in the manifest contract.
- Refactor `internal/agent/resources.go` into typed skill, prompt, instruction,
  and definition catalogs. IDs and versions come from validated metadata.
- Make `list/show/install/uninstall/upgrade` operate on the complete typed set
  and canonical installed sources rather than guessed embedded paths.

**Verification:** embedding inventory, metadata/version parity, prompt raw
output, exact heading/URL/policy assertions, ID prefix rules, instruction marker
tests, atomic projection replacement, retired projection removal, and no
unmanaged content changes.

**Exit:** every selected agent artifact has one canonical ID, release source,
installed path, projection set, and independent CLI representation.

### U05 — Implement layout, ownership manifests, and transaction primitives

**Depends on:** U00 home confirmation, U03 config ownership, and U04 artifact
inventory.

**Goal:** create the common safe foundation used by installer, upgrade,
uninstall, and launcher work.

**Changes:**

- Add focused internal packages for native path layout, strict artifact
  manifests, installed manifests, ownership classification, staging,
  same-filesystem atomic files, locks, transaction journals, and rollback
  snapshots.
- Strictly validate manifest versions, IDs, unique paths, platform selectors,
  checksums, canonical paths, projection paths, ownership classes, and absence
  of absolute/traversing/shared-parent targets.
- Implement unique operation directories under `.guiho/.temp` with strict-child
  proof before cleanup.
- Implement token-owned locks and stale-owner recovery only after verified
  process death.
- Define journal phases: planned, staged, candidate-verified,
  projections-snapshotted, artifacts-replaced, activated, verified,
  rolling-back, rolled-back, complete.
- Add legacy synthetic-manifest discovery for the two old binary locations,
  old skill projections, and Buda-managed instruction markers. Discovery is
  allowlisted and read-only until a transaction commits.

**Verification:** adversarial paths, symlinks/reparse points, duplicate IDs and
checksums, interrupted phases, lock races, rollback snapshots, strict cleanup
containment, shared-directory protection, and legacy/non-Buda discrimination.

**Exit:** later units can mutate only through manifest-backed, rollback-capable
services; no lifecycle code directly deletes an unvalidated path.

### U06 — Build the stable launcher, payload self-test, and instance registry

**Depends on:** U02 and U05.

**Goal:** establish synchronous immutable activation and exact exit forwarding.

**Changes:**

- Add a small launcher entrypoint and internal launcher protocol. It resolves
  native home, strictly reads `current.json`, validates the active relative
  path, starts the payload with original args/streams/environment, waits, and
  returns the exact payload exit code.
- Implement one guarded fallback to the immediately previous verified payload
  when active payload resolution/start fails; never fall back after the payload
  starts and returns an application error.
- Add hidden payload `__self-test` covering version, build target, embedded
  manifest/resources, writable CLI-home diagnostics through an injected test
  root, and zero mutation of the real installation.
- Register each payload invocation under CLI-owned instance state with PID,
  ownership token, executable path, start identity/time, and cleanup on exit.
- Add platform adapters that verify process existence, current-user ownership,
  and executable path before any termination request. qmd and arbitrary child
  processes are never registered as Buda payloads.
- Prototype and prove the Windows locked-file behavior needed for canonical
  launcher removal/quarantine. U11 may not begin until the proof has a native
  Windows test and an agreed synchronous/unlocked cleanup boundary.

**Verification:** launcher/payload integration on native Windows and Unix,
argument/stdin/stdout/stderr fidelity, exit codes including 130, pointer
traversal rejection, active-missing fallback, no fallback on application error,
instance race/stale record tests, and no foreground network work.

**Exit:** the launcher protocol is backward-compatible, versioned, documented,
and suitable for installer repair and upgrade pointer activation.

### U07 — Replace release building with the complete manifest-derived unit

**Depends on:** U03-U06.

**Goal:** build a deterministic self-contained release that can drive every
lifecycle operation.

**Changes:**

- Refactor `devops/build-binaries.go` into data-driven payload and launcher
  matrices while preserving the eight CGO-disabled target tunings.
- Package every confirmed resource, both schemas/examples, instruction,
  manifest, metadata, launchers, and payloads from the same source commit and
  version.
- Generate `artifacts.json` deterministically, then generate sorted
  `checksums.txt` including the manifest and every published artifact except
  checksums itself.
- Validate archive members, normalized timestamps, file modes, exact metadata,
  and no undeclared files.
- Remove exact-11 assertions from Go tests, CI, release workflow, XDocs, README,
  and local instructions; tests compute completeness from the manifest.

**Verification:** reproducible second build comparison where supported,
manifest/schema validation, all checksums, archive-member checksums, eight
payloads and eight launchers, foreign build-only labels, and native payload /
launcher smoke on available runners.

**Exit:** one selected release contains every file needed for fresh install,
repair, upgrade, rollback, init, and uninstall ownership.

### U08 — Rebuild `init` as the common interactive reconciler

**Depends on:** U03-U07.

**Goal:** implement the fixed convention sequence while preserving Buda's
explicit-wiki/qmd domain initialization.

**Changes:**

- Inject terminal detection, question/answer I/O, filesystem, and policy
  decision services; keep Cobra handlers thin.
- Require explicit `--wiki`, resolve/validate it, and never search ancestors.
- Reconcile canonical installed skill sources to all supported global
  projections, ensure `AGENTS.md`, and replace only the Buda managed block.
- Remove/migrate retired Buda blocks previously projected to `CLAUDE.md`
  according to the installed/legacy manifest while preserving all other bytes.
- Load/migrate/create global and project config; ask only unanswered values;
  explain three policy choices; recommend and offer `always-proceed`; keep
  skips at `always-ask`; distinguish global/project ownership in every prompt.
- Validate all answers before atomic writes; preserve existing valid values
  unless the user explicitly confirms replacement.
- Run existing Buda/qmd repository initialization only after common state is
  valid, then revalidate the whole state.
- Print deterministic text/JSON action lists and absolute paths. With no TTY
  and missing answers, fail before writes.

**Verification:** PTY and non-PTY tests, first-run/all-current/partial/migration
cases, each policy branch, invalid answers, write failure rollback, exact
outside-marker preservation, idempotent rerun, qmd failure, and final summary
path/action assertions.

**Exit:** `init` cannot report success unless every common and Buda-specific
postcondition passes.

### U09 — Rewrite Bash and PowerShell installers, including legacy migration

**Depends on:** U07 and U08.

**Goal:** implement equivalent fresh install, exact/channel install, repair,
downgrade, and legacy migration on Unix/macOS and Windows.

**Changes:**

- Implement full-name exact version/channel selectors, mutual exclusion,
  explicit wiki, complete GitHub pagination, exact first-prerelease-ID channel
  matching, target selection, and incomplete-release rejection.
- Download the complete selected manifest into a unique `.guiho/.temp`
  operation; reject missing/duplicate/malformed checksums; validate payload
  version and hidden self-test before installation.
- Install immutable candidate payload/resources, install/repair the compatible
  launcher, snapshot/replace all managed projections, atomically write pointer
  and installed manifest, verify through launcher, and roll back everything on
  failure.
- Add the shared bin directory to user PATH idempotently. PowerShell updates the
  Windows user environment. POSIX manages one documented bounded entry in the
  appropriate user startup file and prints the precise reload requirement.
- Re-running the same version repairs missing artifacts and removes retired
  manifest-owned projections instead of short-circuiting.
- Migrate a verified legacy 0.1.x direct-binary installation only after the new
  launcher passes. Remove an old binary only from the exact historical Buda
  path and only when its identity/version is verified. Preserve both configs,
  OKF data, and all non-Buda files.
- With explicit wiki, run `buda init --wiki` after installation; distinguish an
  init-only failure from a valid installed release and print recovery details.

**Verification:** shared selector fixtures in Go/Bash/PowerShell; local fake
release server; multi-page catalogs; stable/canary/alpha/beta/nightly/custom
channels; unsupported targets; checksum attacks; same-version repair;
upgrade/downgrade; rollback at every phase; PATH idempotency; legacy 0.1.1
migration; native Windows and Linux/macOS acceptance.

**Exit:** both installers produce byte-for-byte equivalent declared installed
state for the same release and platform class.

### U10 — Rebuild upgrade as a synchronous manifest transaction

**Depends on:** U06-U09.

**Goal:** make `buda upgrade` a complete, recovery-first, synchronous release
activation.

**Changes:**

- Add exact `--version` and `--channel` selectors with mutual exclusion and
  default latest stable; extend list/check output with channel, compatible
  artifact, current/latest markers, and complete pagination evidence.
- After local validation and before network work, print the platform installer
  recovery command preserving the requested selector. Route all terminal
  outcomes through a final block pinned to the resolved exact version.
- Acquire the token-owned lock, repair interrupted journals, download/verify the
  complete manifest, run candidate version/self-test, and install a new
  immutable version.
- Verify/terminate only other current-user Buda payload processes whose
  registered and OS-observed executable path equals the old payload. Never
  terminate the invoking payload or child qmd commands.
- Snapshot all replaceable artifacts/projections, remove old manifest-owned
  projections, install the complete new set, atomically activate
  `current.json`, verify through launcher and post-activation self-test, then
  commit installed manifest.
- On any failure restore pointer, artifacts, projections, and journaled state
  before returning. Retain only the immediate previous verified payload for
  fallback; defer only garbage collection of locked inactive files and never
  report `scheduled`.
- After success, run one explicit-wiki `init` reconciliation and raw version
  verification; report init failure separately without rolling back a verified
  installation.

**Verification:** first/final recovery blocks for success/up-to-date/dry-run /
failure; full catalogs; locks; crash injection at every journal phase; process
identity attacks; complete artifact removal/replacement; pointer rollback;
fallback; post-init failure; native Windows and Unix synchronous acceptance.

**Exit:** the original invocation observes and verifies the new active release;
no detached process is success authority.

### U11 — Implement manifest-based uninstall across Cobra and both scripts

**Depends on:** U05, U06 Windows removal proof, U08, and U09.

**Goal:** deliver one ownership-safe destructive-default contract without
touching canonical wiki knowledge or shared GUIHO infrastructure.

**Changes:**

- Build a shared Go planning engine that consumes strict installed/release
  manifests plus the explicit wiki and emits ordered `REMOVE`/`PRESERVE`
  records with ownership proof.
- Implement `--preserve-config`, `--preserve-data`, `--dry-run`, and `--yes`;
  remove `--keep-agent-resources`.
- Require confirmation on an interactive terminal without `--yes`; fail before
  mutation on noninteractive input without `--yes`.
- Remove all versions, Buda launcher, manifests, state/cache, projections,
  bounded instruction, and explicit wiki project config by default. Preserve
  shared directories/PATH, other CLIs, all unmanaged `AGENTS.md` bytes, qmd-
  owned state, and canonical OKF/wiki data.
- Preserve global/project config only with `--preserve-config`; preserve future
  declared CLI-home databases/data only with `--preserve-data`.
- Implement equivalent Bash/PowerShell script flags and explicit wiki
  parameters. Test their plans against the Go engine's golden fixtures.
- On Windows, remove canonical paths synchronously using the U06-proven
  quarantine/cleanup boundary. Do not report success while the canonical Buda
  launcher or CLI home remains active. Any unavoidable locked inactive garbage
  is clearly classified as cleanup, never as pending uninstall authority.

**Verification:** exact plan snapshots; interactive/noninteractive behavior;
all preservation combinations; partial/corrupt manifests; symlink/reparse path
attacks; other-CLI/shared-path protection; OKF data invariance; rollback/failure
injection; native scripts and Cobra parity; Windows locked-image acceptance.

**Exit:** all three interfaces implement the same verified plan and leave no
Buda-owned active installation after success.

### U12 — Rewrite public and durable documentation for the new lifecycle

**Depends on:** U02-U11 behavior frozen.

**Goal:** make public instructions exact, current, and safe.

**Changes:**

- Rewrite README installation for PowerShell and Linux/macOS with exact remote
  commands, explicit wiki, raw version verification, selectors, PATH reload,
  legacy 0.1.x migration, upgrade, recovery, and layout.
- Make `## Uninstall` the final operational section. Include remote scripts,
  destructive default, dry run, confirmation, and combined config/data
  preservation examples; explicitly state that canonical OKF content is never
  an uninstall target.
- Update `ARCHITECTURE.md` to implemented launcher/payload/config/manifest /
  transaction boundaries and preserve the qmd/OKF product rules.
- Keep `IMPLEMENTATION.md` historical and link forward to this plan and the
  acceptance matrix.
- Refresh every affected XDocs descriptor, root index, changelog unreleased
  section, agent guidance, and workflow descriptions. Remove current-facing
  0.0.2/11-asset claims while keeping clearly historical evidence.

**Verification:** every README command exists in RunX or is a documented remote
entry point; URL/version checks; strict XDocs; link checks; help/docs parity;
fresh-user and migration walkthrough reviews.

**Exit:** a new user and a 0.1.1 user can follow the docs without unstated
commands or unsafe assumptions.

### U13 — Install the complete CI and release acceptance gate

**Depends on:** U01-U12.

**Goal:** prevent regression to the old contract and prove the complete unit on
the exact integration head.

**Changes:**

- Make CI invoke RunX-cataloged format, tidy, test, vet, schema, manifest,
  XDocs, Mirror, installer, launcher, upgrade, uninstall, and migration checks.
- Build all payload and launcher targets; verify `artifacts.json`, archives,
  checksums, raw versions, embedded targets, and undeclared/missing assets.
- Run native launcher/payload/install/upgrade/uninstall smokes on Linux AMD64 /
  ARM64, macOS AMD64/ARM64, and Windows AMD64/ARM64 where runners exist.
  Clearly label ARMv6/ARMv7 foreign builds as build-only.
- Add end-to-end tests from a synthetic 0.1.1 legacy installation, fresh
  installation, same-version repair, upgrade, downgrade, interrupted upgrade,
  launcher fallback, all uninstall preservation modes, and reinstall recovery.
- Change publication reconciliation to the manifest-derived release set. Never
  delete a declared artifact because of a hard-coded count.
- Add a gate that the released main skill/instruction/prompt/schema/example /
  manifest versions agree with the Mirror target version.

**Verification:** CI green on the exact integration head; artifacts downloaded
from the workflow and independently rechecked; no package/source publication,
tag, or production action.

**Exit:** every acceptance-matrix row has automated or explicitly native manual
evidence tied to one commit.

### U14 — Aggregate review and authorized minor-release preparation

**Depends on:** U13 and upstream shared CLI skill alignment.

**Goal:** prepare a release-ready aggregate without crossing the release
authorization boundary.

**Changes:**

- Rebase/merge the integration line, then perform independent implementation
  review and validation on the exact aggregate head.
- Reconcile `TODO.md`, audit finding statuses, plan unit statuses, changelog,
  README, XDocs, Mirror config, RunX, schemas, manifests, and workflows.
- Fetch/prune and prove branch ancestry and a clean worktree without discarding
  user work.
- Run plain Mirror once if required by the managed instruction, then
  `mirror config check`, `mirror version current`, and
  `mirror version plan minor`. Inspect every file mutation, release commit,
  canonical tag, push, and workflow effect.
- Stop before `mirror version apply`. Report the proposed version (expected
  `0.2.0` from `0.1.1`), exact commit, complete validation, native versus
  build-only targets, artifact inventory, and remaining authorization.

**Verification:** final aggregate PR accepted; exact-head CI and independent
validation green; Mirror plan clean and reviewable; no tag/release exists.

**Exit:** one separately authorized action can apply the Mirror plan and publish
the complete release without source changes.

## Per-unit review and validation template

Every implementation unit PR must contain this evidence:

1. **Scope:** unit ID, owned files/modules, and explicit non-goals.
2. **Authority:** convention sections and audit finding IDs closed.
3. **Compatibility:** intentional public breaks, migration behavior, and data
   preservation statement.
4. **Tests:** focused unit/integration tests plus full `go test ./...` and
   `go vet ./...` unless a documented external runner is required.
5. **Formatting/modules:** clean gofmt and `go mod tidy -diff`.
6. **Tooling:** `runx check/list`, relevant dry-runs, Mirror config check,
   strict XDocs meta/tree/doctor.
7. **Platform evidence:** native execution named separately from foreign
   build-only results.
8. **Filesystem safety:** exact paths, containment proof, ownership source,
   rollback result, and preserved user data.
9. **Review:** independent review tied to the exact PR head and resolution of
   all actionable findings.
10. **Delivery boundary:** no version apply, tag, release, package publication,
    or production mutation unless separately authorized.

## Failure and rollback policy during implementation

- A unit with a failed source or acceptance test does not merge.
- Never repair a lifecycle test by weakening ownership, checksum, raw-version,
  confirmation, or rollback assertions.
- Preserve dirty user work. If in-scope work overlaps unknown changes, stop and
  identify the exact files; never reset, force-push, or erase them.
- A failed lifecycle transaction in tests must leave either the old verified
  installation active or no new installation; partial mixed-version state is a
  test failure.
- A failed init after a verified install/upgrade leaves the verified release
  installed and reports reconciliation failure separately.
- A failed uninstall before commit restores the old active installation. Once
  canonical removal commits, locked inactive garbage may be cleaned only under
  its validated Buda operation directory and must never restore an active
  installation unexpectedly.
- No test may read `.env`, `encrypted.env`, keys, or secret directories. Release
  tests use public GitHub data or an isolated local fake release server.

## Risk register and mandatory mitigations

| Risk | Consequence | Mandatory mitigation/gate |
|---|---|---|
| Windows running launcher/payload cannot be deleted synchronously | Cobra uninstall could falsely claim completion or leave an active command. | U06 native proof before U11; canonical paths must be removed/quarantined during the original operation, and remaining locked inactive garbage cannot be success authority. |
| Shell implementations drift from Go semantics | Version/channel or uninstall behavior differs by platform. | One shared machine-readable fixture corpus; Bash, PowerShell, and Go acceptance consume identical selector, manifest, plan, and preservation vectors. |
| Canonical tag contains `/` | Generated schema/release URLs may be invalid or mutable. | Central URL builder, exact encoded-tag fixtures, and public release URL smoke before U14. |
| Legacy 0.1.1 updater understands only one payload | Direct `buda upgrade` cannot bootstrap the complete new release safely. | Supported migration path is the new exact/channel installer; README/skill recovery guidance directs 0.1.x users to it, and U09 starts from an official-layout legacy fixture. |
| Destructive uninstall confuses maintained wiki with CLI-owned data | Irrecoverable canonical knowledge loss. | Normative ownership table, explicit wiki, manifest/config-derived targets, OKF byte-invariance tests across all uninstall interfaces, and independent destructive-path review. |
| Current project requirement encourages implicit discovery | Violates core Buda contract and may mutate the wrong repository. | Lifecycle scripts require full-name wiki parameter; CLI requires `--wiki`; no cwd/ancestor fallback and no wiki registry. |
| Old global Go CLI skill reintroduces 11 assets or `update` | Future agents regress the repository after migration. | U00 upstream amendment recorded; U14 blocks release until mandatory shared skill and local instructions align. |
| Partial main branch is accidentally released | Users receive an internally inconsistent lifecycle. | Integration branch strategy, no tags, manifest-based publication gate, exact-head aggregate PR, and separate Mirror apply authorization. |

## Definition of complete

Buda complies only when all of the following are true on the same exact commit:

- every row in the acceptance matrix is green with evidence;
- all U00-U13 TODO items are complete and U14 release preparation is ready;
- the effective product still requires one explicit wiki and delegates all
  discovery to qmd;
- raw version, help flags, required commands, dual config, evolution policy,
  schemas/examples, agent resources, init, launcher, manifest, installers,
  synchronous upgrade, and uninstall all pass native acceptance;
- complete release assets are derived from `artifacts.json`, checksummed, and
  built for all supported targets;
- RunX, Mirror config, XDocs, gofmt, tidy, tests, vet, cross-builds, and native
  smokes pass;
- documentation describes implemented truth and ends with the required
  uninstall section;
- an independent reviewer and validator accept the exact aggregate head; and
- no release claim is made until a separately authorized Mirror minor apply,
  remote tag, workflow, GitHub Release, assets, checksums, and native downloaded
  smoke are all verified.
