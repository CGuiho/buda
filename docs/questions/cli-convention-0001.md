---
name: buda-cli-convention-0001-questions
purpose: Record resolved implementation decisions and review-gated validation boundaries for Convention 0001.
description: Resolved answers selected during unattended Convention 0001 implementation.
created: 2026-08-16
owner: buda-questions
flags: []
tags: [questions, cli, convention-0001]
keywords: [Buda, GUIHO CLI Convention 0001, stable launcher, release manifest]
status: resolved-review-ready
---

# Convention 0001 execution question ledger

The plan executor continued unattended using the approved audit, plan, and
acceptance matrix. Naming, artifact-model, lifecycle-policy, and validation
boundary decisions are resolved below. Independent review remains required
before merge; none authorizes a tag, release, or production action.

## Q01 — Complete release asset count versus historical exact-eleven guidance

- Question: Should the implementation preserve the older exact-11 authored
  asset model or publish the complete Convention 0001 manifest-derived set?
- Evidence: Convention 0001 requires every payload, launcher, agent artifact,
  instruction, schema, example, manifest, and checksum. The approved plan and
  matrix explicitly mark the old exact-eleven assertions as obsolete.
- Candidate answers: retain exact eleven; use the manifest-derived complete
  set.
- Chosen answer: use `artifacts.json` as authority and publish 25 files in the
  current build (24 manifest artifacts plus `checksums.txt`).
- Rationale: the newer controlling convention and approved plan supersede the
  older Go skill/AGENTS exact-eleven wording while preserving all eight CLI
  payload targets.
- Confidence: high. Reversibility: high before release publication.
- Action: updated release builder, workflows, tests, docs, and acceptance
  matrix to validate counts from the manifest rather than a hard-coded eleven.
- Decision status: resolved; independent implementation review remains an
  acceptance gate.

## Q02 — Stable launcher replacement while the invoking launcher is active

- Question: How should a Windows upgrade replace the canonical stable launcher
  when the launcher process is the parent waiting on the payload?
- Evidence: the transaction never overwrites the active payload, stages and
  checksums the selected launcher asset, and verifies the existing stable
  launcher after pointer activation. Windows executable replacement can be
  locked while the launcher parent remains active.
- Candidate answers: detach a helper; replace the running launcher in place;
  retain the generic verified launcher and stage the selected launcher inside
  the immutable release until a compatible synchronous replacement mechanism
  is approved.
- Chosen answer: fail closed on unsafe in-place replacement; if the canonical
  launcher is missing, install the verified selected launcher transactionally;
  if it already exists, verify the version-independent launcher and leave it
  untouched while the invoking Windows process may hold the file open. The
  selected launcher is always stored and checksummed in the immutable release.
- Rationale: detached helpers are forbidden as upgrade authority, and the
  existing launcher is version-independent and can dispatch every immutable
  payload. This preserves synchronous activation without claiming an unsafe
  in-place replacement succeeded.
- Confidence: high. Reversibility: high before merge.
- Action: implemented missing-launcher installation, complete rollback of a
  newly created launcher, and post-activation launcher verification. No helper
  or deferred `scheduled` replacement was introduced.
- Decision status: resolved; Windows locked-launcher mutation remains a native
  acceptance scenario for independent review.

## Q03 — Native installer and migration execution on this Windows host

- Question: Can the live host safely exercise remote/fake-release installer,
  legacy migration, and locked-file rollback scenarios end to end?
- Evidence: POSIX syntax, PowerShell parser, manifest/checksum checks, native
  payload smoke, and isolated stable-launcher smoke pass. Running installer
  mutations against a user home or simulating locked native images would alter
  user state and is outside the current safe validation fixture.
- Candidate answers: mutate the real user home; execute a fully isolated fake
  release fixture; record the native behavior as review-gated.
- Chosen answer: record native installer/migration acceptance as review-gated
  and leave production/user installation state untouched. The repository
  scripts remain syntax-validated and their fake-release path is represented
  by `BUDA_RELEASE_ASSET_DIR`; only an isolated fixture or native CI runner
  should exercise mutation and lock/recovery behavior.
- Rationale: the user authorized repository delivery, not destructive mutation
  of an existing personal installation; the scripts already have syntax and
  static contract coverage.
- Confidence: high. Reversibility: high.
- Action: added explicit remaining-gap evidence to `docs/todo/` and the PR
  handoff.
- Decision status: resolved validation boundary; native acceptance remains an
  explicit independent-review gate.
