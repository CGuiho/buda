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
