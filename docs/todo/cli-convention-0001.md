---
name: buda-cli-convention-0001-execution
purpose: Durable execution record for the approved GUIHO CLI Convention 0001 implementation.
description: Tracks the unattended implementation branch, plan units, provisional decisions, validation state, and release deferral for Buda Convention 0001.
created: 2026-08-16
updated: 2026-08-16T00:00:00Z
owner: buda-todo
status: review-ready
plan_unit: U00-U14 aggregate execution
skipped_phases:
  - brainstorm: approved convention audit and implementation plan already provide requirements.
  - requirements: acceptance matrix and explicit human confirmations are durable.
  - architecture: target lifecycle architecture is recorded in the implementation plan.
  - plan-review: the user explicitly delegated execution of the approved plan.
flags: []
tags:
  - todo
  - cli
  - convention
  - execution
keywords:
  - GUIHO CLI Convention 0001
  - Buda
  - 0.2.0
---

# GUIHO CLI Convention 0001 execution

## Scope

Execute the approved Buda Convention 0001 implementation plan on the dedicated
`codex/cli-convention-0001` branch. The accepted target is the next minor
release, `0.2.0` from `0.1.1`; Mirror apply, tagging, and GitHub Release
publication remain deferred to the integration owner after independent review.

## Ownership and handoff

- Branch: `codex/cli-convention-0001`
- Base: `09cba7a` / `origin/main` at delegation
- Worktree: `C:\GUIHO\buda`
- Executor: `guiho-a-0048-plan-executor` / `guiho-s-0023-plan-executor`
- Next handoff: `guiho-a-0049-implementation-reviewer`, then validation and PR integration.
- Question ledger: `docs/questions/cli-convention-0001.md` records the
  manifest-count, stable-launcher, and native-installer provisional answers.
- Production boundary: no deployment, traffic, DNS, database, secret, or live-service mutation.

## Provisional answers

| Question | Chosen answer | Evidence and rationale | Human review |
|---|---|---|---|
| Which release target follows the breaking convention implementation? | `0.2.0` minor from `0.1.1` | Explicit latest user direction overrides the plan's provisional major wording; Mirror remains authoritative for application. | pending |
| Should legacy direct self-replacement retain a successful deferred/scheduled state? | No; stable-launcher transactions are synchronous and legacy direct replacement must fail closed when it cannot replace in-process. | Convention requires a synchronous next invocation and rejects scheduled success. | pending |
| Which project instruction file is canonical? | Always reconcile `AGENTS.md`; reconcile existing `CLAUDE.md` as a projection. | Convention requires a managed AGENTS block and Buda preserves unrelated bytes. | pending |

## Milestones

- [x] U00-U04 authority, public surface, config, and agent resource implementation complete.
- [x] U05-U08 lifecycle primitives, launcher, release builder, and init integration complete.
- [x] U09-U11 installer, upgrade, and uninstall implementations complete; native mutation/recovery acceptance remains review-gated.
- [x] U12-U13 documentation, tooling, CI, and release-preparation implementation complete.
- [ ] U14 aggregate review and release preparation complete.
- [ ] Independent implementation review and validation accepted on exact head.

## Testing state

Current state is `review-ready`. The final local validation pass recorded:

- `gofmt` and `git diff --check`: clean.
- `go mod tidy -diff`, `go test -count=1 ./...`, and `go vet ./...`: passed.
- Pure-Go builds: Windows amd64/arm64 and Linux ARMv6/ARMv7 passed; foreign
  ARM targets are build-only on this Windows host.
- `runx check --format json`, `runx list --format json`, and dry-runs for all
  28 catalog entries: passed without executing catalog commands.
- `xdocs meta . --documents --strict`, `xdocs tree`, and
  `xdocs doctor . --warnings-as-errors`: 35 descriptors, zero errors/warnings.
- `mirror`, `mirror config check`, `mirror version current`, and
  `mirror version plan minor`: passed; current `0.1.1`, planned `0.2.0`, tag
  `buda/v0.2.0`. Mirror apply was not run.
- Local release build: 25 generated files and 24 manifest artifacts (eight
  payloads, eight launchers, seven managed resources, manifest, and checksums);
  SHA-256 verification, native Windows payload smoke, and stable-launcher smoke
  passed. Generated `dist`, `.gocache`, and `.gomodcache` were removed.
- POSIX `install.sh`/`uninstall.sh` syntax and PowerShell installer/uninstaller
  parsing: passed.

Remaining gates are the independent exact-head review, full native installer
mutation/reinstall and crash-recovery suites, legacy 0.1.1 migration evidence,
and the separately authorized Mirror apply/tag/release boundary. Do not mark
the record complete until the PR head is reviewed and pushed and the deferred
release boundary is explicitly recorded.
