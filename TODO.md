---
name: buda-todo
purpose: Track approval and execution of the Buda GUIHO CLI Convention 0001 implementation program.
description: Durable repository checklist for the breaking convention migration, implementation units, validation gates, aggregate review, and separately authorized minor release.
created: 2026-08-16
owner: buda-package
flags: []
tags:
  - todo
  - cli
  - convention
keywords:
  - Buda
  - GUIHO CLI Convention 0001
  - implementation units
  - acceptance matrix
  - minor release
---

# Buda TODO

## Sources of truth

- [Compliance audit](docs/GUIHO_CLI_CONVENTION_0001_COMPLIANCE_AUDIT.md)
- [Implementation plan](docs/GUIHO_CLI_CONVENTION_0001_IMPLEMENTATION_PLAN.md)
- [Acceptance matrix](docs/GUIHO_CLI_CONVENTION_0001_ACCEPTANCE_MATRIX.md)
- [Historical 0.0.2 implementation record](docs/IMPLEMENTATION.md)

The implementation plan owns scope and order. The acceptance matrix owns
closure evidence. This file records progress only; checking an item without its
required exact-head evidence does not complete the unit.

## Planning

- [x] Audit Buda against GUIHO CLI Convention 0001.
- [x] Write the dependency-ordered implementation plan.
- [x] Map every audit finding to acceptance evidence.
- [x] Obtain human approval of the implementation plan through the explicit implementation delegation.

## Required human confirmations

- [x] Confirm CLI home directory name `buda`.
- [x] Confirm retained main skill name `guiho-s-0002-buda`.
- [x] Confirm main setup prompt ID `guiho-p-buda`.
- [x] Confirm Buda ships no agent definitions.
- [x] Confirm canonical repository and issue URLs.
- [x] Confirm the next minor release target `0.2.0` from `0.1.1`.

## Implementation units

- [x] **U00 — Convention authority and naming:** update repository instructions, record all accepted names/URLs/release intent, mark the 0.0.2 plan historical, and record the external shared Go CLI skill amendment follow-up.
- [x] **U01 — Tooling baseline:** Go 1.26, complete RunX catalog, corrected Mirror mapping, one XDocs tree, and mandatory CI tooling gates.
- [x] **U02 — Public Cobra surface:** raw version, current-command help docs, `max | >1` tree depth, global-flag control, and agent `upgrade` commands.
- [x] **U03 — Configuration:** dual strict config, merge semantics, agent evolution, schemas, examples, pinned URLs, and legacy config migration.
- [x] **U04 — Agent artifacts:** evolution guidance, confirmed setup prompt, typed resource families, version parity, and optional confirmed definitions.
- [x] **U05 — Lifecycle primitives:** canonical paths, manifests, ownership, staging, locks, journals, snapshots, and legacy discovery.
- [x] **U06 — Launcher/payload:** stable launcher, strict pointers, fallback, hidden self-test, instance registry, and native Windows removal proof.
- [x] **U07 — Complete release unit:** payload/launcher matrix, resources, schemas/examples, artifacts manifest, checksums, and manifest-derived gates.
- [x] **U08 — Init reconciler:** explicit wiki, interactive policies, AGENTS, configs, canonical agent projections, qmd/domain setup, and final summary.
- [x] **U09 — Installers/migration:** full catalogs and channels, complete transactional install/repair/downgrade, PATH, legacy 0.1.1 migration, and post-install init.
- [x] **U10 — Synchronous upgrade:** recovery-first output, complete manifest transaction, process safety, rollback, verification, and post-upgrade init.
- [x] **U11 — Uninstall:** two scripts plus Cobra parity, full plan/confirmation/preservation, ownership safety, and native Windows completion.
- [x] **U12 — Documentation:** current README/architecture/XDocs/agent guidance, migration instructions, and final Uninstall section.
- [x] **U13 — Aggregate CI/release gates:** workflow/tooling implementation, manifest publication, and exact-head evidence preparation; native lifecycle and migration gates remain review-gated.
- [ ] **U14 — Aggregate review and release preparation:** accepted integration PR, TODO reconciliation, independent validation, and clean Mirror minor plan for 0.2.0.

## Aggregate acceptance gates

- [ ] `INV-01` through `INV-06` product invariants pass.
- [ ] `ACC-A1` through `ACC-I5` audit-finding acceptance rows pass.
- [ ] `GATE-01` format and `GATE-02` module checks pass.
- [ ] `GATE-03` full tests and `GATE-04` vet pass.
- [ ] `GATE-05` Mirror config, RunX, and strict XDocs pass.
- [ ] `GATE-06` schemas/examples and `GATE-07` complete release build pass.
- [ ] `GATE-08` native lifecycle and `GATE-09` legacy migration pass.
- [ ] `GATE-10` transaction recovery and `GATE-11` documentation pass.
- [ ] `GATE-12` independent exact-head review/validation is accepted.
- [ ] `GATE-13` Mirror minor plan for 0.2.0 is complete, clean, and reviewed.

The local implementation pass has already demonstrated GATE-01 through
GATE-07's repository checks, but the aggregate gates remain unchecked until
the exact PR head receives independent review. Native installer/migration and
crash-recovery execution remain review-gated because this Windows checkout was
not allowed to mutate an existing personal installation.

## Kimi review fix pass

A review fix pass on `codex/cli-convention-0001` closed the release blockers
that remained after the first review round. Every fix below is evidenced by
source, tests, workflows, or the disposable-home native Windows smoke.

| Blocker | Closure |
|---|---|
| POSIX chmod before candidate execution | `devops/install.sh` chmods staged payload/launcher before `--version`; `internal/upgrade/transaction.go` runs `ensureExecutable` before the staged candidate executes. |
| Capture replacement preserves unknown OKF metadata | `internal/capture/capture.go` edits known fields through `okf.Document` setters instead of rebuilding frontmatter; `TestCaptureReplacePreservesUnknownOKFMetadata` byte-proves custom keys, tags, and nested maps survive. |
| Manifest projection apply/snapshot/retire/rollback and crash recovery | `internal/upgrade/transaction.go` snapshots, applies, retires, and rolls back manifest projections; `recoverInterrupted` states are tested exhaustively in `TestRecoverInterruptedJournalStates`. |
| Symlink/reparse-safe quarantined transactional uninstall | `internal/lifecycle` detects Windows reparse points via `FILE_ATTRIBUTE_REPARSE_POINT` (`linkcheck_windows.go`); `internal/uninstall/plan.go` quarantines under `.guiho/.temp`, restores on any failure, retains the quarantine instead of losing data, and edits instruction blocks before the wiki config is quarantined. |
| Init discovers all unanswered policies before any write | `cmd/init.go` resolves/migrates the global config without writing, prompts every missing policy, and persists only afterwards; `TestInitFailsBeforeAnyWriteWhenPoliciesAreUnansweredWithoutTerminal` proves zero writes. |
| Full release matrix, duplicate rejection, typed launcher metadata, ARMv6/v7 | Go catalog, `install.sh` jq filter, and `install.ps1` all require the complete 24-binary matrix and reject duplicate asset names; `setTargetMetadata` types payload and launcher entries; ARMv6/ARMv7 appear in every selector and the release manifest. |
| Meaningful zero-mutation self-test and probe instance bypass | `__self-test` verifies the build version and reads every embedded skill/prompt/schema/example; only exact `--version`/`-v`/`__self-test` invocations bypass instance registration; `TestVersionAndSelfTestDoNotMutateFileSystem` proves zero mutation. |
| Historical install migration exclusively through the launcher transaction | Installers remove the verified legacy binary only after launcher activation, from the exact historical paths (`~/.local/bin/buda`, `%LOCALAPPDATA%\GUIHO\bin\buda.exe`); no in-place mutable route remains. |
| Explicit wiki and grouped plan before uninstall confirmation | All three uninstall interfaces require the explicit wiki and print grouped `REMOVE`/`PRESERVE` plans before any confirmation or mutation. |
| Injected disposable home and real-profile sentinels | Lifecycle tests and the native smoke inject `USERPROFILE`/`HOME` into disposable directories and assert sentinel survival; the real user profile is never touched. |
| Embedded strict schema/runtime/example parity | `schemas/embed.go` plus `TestEmbeddedSchemasPresentAndValid`, `TestEmbeddedExamplesValidateAgainstRuntimeConfig`, and `TestDecodingRejectsMissingSchema` pin offline parity. |
| CI/native lifecycle/migration/interruption, RunX cross targets, XDocs, README | `ci.yml` adds POSIX and Windows native lifecycle jobs (install, repair, migration, synchronous uninstall) with disposable homes and qmd stubs; cross targets use the target-aware release builder; XDocs descriptors and the README describe the implemented truth. |
| Pointer EOF, canonical payload names, strict digest rules | `ReadPointer` rejects trailing documents and noncanonical payload filenames; manifests reject all-zero digests except `artifacts.json`; checksum readers reject duplicates, extras, and malformed entries. |

Evidence recorded during the fix pass:

- disposable-home native Windows lifecycle smoke: install, same-version
  repair, legacy 0.1.1 direct-binary migration, and synchronous uninstall all
  passed with sentinel preservation (real profile never inspected);
- `TestRecoverInterruptedJournalStates`, `TestManifestDrivenProjectionsApplyAndRollback`,
  `TestApplyQuarantineRollbackOnFailure`, and the lifecycle acceptance test run
  are cataloged in `runx.yaml` and CI;
- `actionlint` passes for both workflows; POSIX and PowerShell installer
  syntax checks pass.

## Release boundary

- [ ] Apply the separately authorized Mirror minor plan,
  create/push the release commit and annotated canonical tag, and publish the
  GitHub Release.
- [ ] After authorization, verify remote ancestry, workflow success, release
  publication, complete manifest-derived assets, checksums, and downloaded
  native launcher/payload smoke.
- [ ] Confirm explicitly that no production deployment, traffic, DNS, database,
  secret, or live-service action occurred.

No unchecked release-boundary item is implied by implementation approval.
