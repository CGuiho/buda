---
name: buda-guiho-cli-convention-0001-acceptance-matrix
purpose: Map every Convention 0001 audit finding to its implementation unit and required acceptance evidence.
description: Traceability matrix for the Buda CLI convention migration, covering all audit findings, product invariants, automated validation, native platform evidence, documentation, and release-readiness gates.
created: 2026-08-16
owner: buda-docs
flags: []
tags:
  - acceptance
  - validation
  - traceability
  - cli
keywords:
  - Buda
  - GUIHO CLI Convention 0001
  - audit findings
  - acceptance criteria
  - implementation units
  - release gates
---

# Buda Convention 0001 acceptance matrix

## How to use this matrix

This document is the closure contract between the
[`compliance audit`](GUIHO_CLI_CONVENTION_0001_COMPLIANCE_AUDIT.md), the
[`implementation plan`](GUIHO_CLI_CONVENTION_0001_IMPLEMENTATION_PLAN.md), and
[`TODO.md`](../TODO.md).

For each row:

- **Unit** names the implementation unit that owns closure.
- **Implementation evidence** is the state that must exist in source or
  documentation.
- **Acceptance evidence** is the minimum test, command, or native validation
  required on the exact unit or aggregate head.
- A row is not complete because code was written. It closes only when both
  evidence columns are satisfied and reviewed.

Test names below are stable acceptance IDs. Executors may split one ID across
Go, shell, PowerShell, and workflow tests, but the ID must remain visible in
test names, fixtures, RunX descriptions, or CI step names so the requirement is
traceable.

## Product-invariant gates

| Gate | Required invariant | Acceptance evidence |
|---|---|---|
| `INV-01` | Every repository-facing lifecycle/domain action operates on one explicit `--wiki`/script equivalent; no discovery or global wiki registry exists. | Command-tree and lifecycle tests reject missing wiki where project state is required; repository search and review find no ancestor search/registry. |
| `INV-02` | qmd is the sole index/retrieval engine. | Existing qmd adapter tests remain green; dependency review finds no search/embedding/ranking fallback. |
| `INV-03` | Canonical OKF bundle/raw evidence is user-owned and never an uninstall target. | Destructive-default uninstall tests snapshot and byte-compare the full OKF bundle before/after every interface. |
| `INV-04` | Base OKF conformance and Buda health remain distinct. | Existing lint/health fixtures and output assertions remain green. |
| `INV-05` | No lifecycle operation mutates another CLI, shared `.guiho` containers, or the shared PATH entry during uninstall. | Adversarial ownership fixtures contain neighboring CLIs and sentinel files; all remain unchanged. |
| `INV-06` | No production, secret, package-registry, or cross-repository publication behavior is introduced. | Workflow/static review and release dry-runs show GitHub Release source publication only; no secret-bearing files are inspected. |

## Audit finding traceability

### A. Technology and mandatory tooling

| Finding | Unit | Implementation evidence | Acceptance evidence |
|---|---|---|---|
| `A1` Obsolete creation-time Go version | U01 | `go.mod` and workflows use the approved Go 1.26 patch line. | `ACC-A1`: `go version`, `go env GOTOOLCHAIN`, tidy diff, tests, vet, and all target builds agree on the pinned line. |
| `A2` Missing RunX catalog | U01 | Root `runx.yaml` catalogs every supported repeatable operation with namespace, stable UIDs/IDs, exact commands, and descriptions. | `ACC-A2`: `runx check --format json`, `runx list --format json`, selector uniqueness test, and dry-run of every mutating/high-impact entry. |
| `A3` Mirror not sole version authority | U01, U13 | Dead JS paths are removed; actual version-bearing files are mapped or checked for parity; release gates derive versions from Mirror target metadata. | `ACC-A3`: `mirror config check`, clean `mirror version plan minor`, and a no-write parity test covering binary/resources/schemas/examples/docs/manifests. |
| `A4` Disconnected/stale XDocs ownership | U01, U12 | `.github` joins the single root tree; root files and lifecycle documentation are current. | `ACC-A4`: strict full-repo meta, tree containing `buda-github`, doctor with warnings as errors, and companion-document inventory parity. |

### B. Cobra version, commands, flags, and help

| Finding | Unit | Implementation evidence | Acceptance evidence |
|---|---|---|---|
| `B1` Prefixed version output | U02 | Root version template emits only raw SemVer and bypasses notices. | `ACC-B1`: exact byte tests for `-v`/`--version` on built payload and launcher; stdout is `<semver>\n`, stderr empty. |
| `B2` Recursive `--help-docs` | U02 | Renderer emits Markdown only for the invoked command. | `ACC-B2`: root/group/leaf snapshots contain one command page and no descendant page separators. |
| `B3` Wrong help-tree depth grammar | U02 | Depth parser accepts `max` or integer `>1`, default `max`, in space/equals forms. | `ACC-B3`: accepts `max`, `2`, large valid values; rejects empty, `0`, `1`, negative, noninteger, overflow. |
| `B4` Missing global-flag switch | U02 | Presence-only `--help-tree-global-flags` controls descendant repetition. | `ACC-B4`: root and nested snapshots show globals once by default and at every descendant only when present. |
| `B5` Prohibited agent `update` commands | U02 | Skill/instruction trees expose `upgrade`; code/docs/scripts contain no public `update` route. | `ACC-B5`: fresh-tree inventory equals required command set; repository search rejects obsolete invocations; unknown `update` exits with usage error. |

### C. Configuration and schemas

| Finding | Unit | Implementation evidence | Acceptance evidence |
|---|---|---|---|
| `C1` No distinct global/project config | U03 | Typed `buda.global.yaml` and project `buda.yaml` contracts and paths exist. | `ACC-C1`: synthetic-home tests resolve exact native paths and never substitute one filename/schema for the other. |
| `C2` No merge/inheritance | U03 | Global baseline merges with project per-field overrides; project-only fields remain project-only. | `ACC-C2`: table tests cover absent/partial/full overrides, immutable `wiki_id`, and deterministic effective config. |
| `C3` Missing `agent.evolution` | U03 | Exact enum policy exists for upgrade/bugs/improvements/reviews with `always-ask` defaults. | `ACC-C3`: all three values accepted, Boolean/unknown rejected, inheritance/defaults verified, policy action gates tested. |
| `C4` Missing schemas/examples/pinned comments | U03, U07 | Two strict schemas, two complete examples, embedded copies, and exact-release schema comments exist and ship. | `ACC-C4`: examples validate; Go/schema parity; URLs contain the encoded exact canonical tag and never `main`/`latest`; published-set fixture includes all four files. |
| `C5` Partial runtime validator only | U03 | Offline runtime validation covers both decoded layers and the merged effective result. | `ACC-C5`: network-disabled config suite covers unknown fields, multiple docs, paths, enums, schema parity, and semantic merge errors. |

### D. Mandatory init sequence

| Finding | Unit | Implementation evidence | Acceptance evidence |
|---|---|---|---|
| `D1` Missing common init sequence | U08 | Init executes all 16 convention steps around Buda/qmd initialization. | `ACC-D1`: ordered service-call trace and postcondition suite prove common steps precede domain setup and final revalidation follows it. |
| `D2` Wrong interactive setup behavior | U08 | Existing values load first; only missing answers prompt; no-TTY missing values fail before writes. | `ACC-D2`: PTY/no-PTY fixtures, scope labels, skipped choices defaulting to `always-ask`, invalid answer retries, and zero-write failure assertions. |
| `D3` AGENTS not always canonical | U08 | `AGENTS.md` is always ensured and owns the one Buda block; retired Buda CLAUDE projection is safely removed/migrated. | `ACC-D3`: AGENTS-only/CLAUDE-only/both/neither/malformed marker fixtures preserve all unmanaged bytes. |
| `D4` Incomplete final validation/summary | U08 | Aggregate validator and deterministic text/JSON action/path summary exist. | `ACC-D4`: success requires every postcondition; failure never prints success; all result paths are absolute and actions are created/upgraded/verified/unchanged. |

### E. Agent resources

| Finding | Unit | Implementation evidence | Acceptance evidence |
|---|---|---|---|
| `E1` Missing CLI Evolution and Feedback section | U04 | Main skill has exact heading, URLs, categories, policies, upgrade/init/version flow, and issue URL rule. | `ACC-E1`: metadata/content contract test asserts exact heading and all mandatory semantic clauses. |
| `E2` No setup prompt | U04 | Confirmed prompt is distinct from instruction and explains install, verify, init, and upgrade. | `ACC-E2`: prompt catalog/list/show tests, ID prefix validation, and required-content assertions. |
| `E3` Missing definitions/confirmation evidence | U00, U04 | Naming decision record exists; every selected definition is typed and catalogued, or confirmed set is explicitly empty. | `ACC-E3`: release inventory equals confirmed decision record; no unconfirmed/inferred ID appears. |
| `E4` Resource version drift | U04, U07, U13 | Resource metadata and manifest version derive from the one release target. | `ACC-E4`: build rejects any skill/prompt/instruction/definition/schema/example version mismatch. |

### F. Installation and release unit

| Finding | Unit | Implementation evidence | Acceptance evidence |
|---|---|---|---|
| `F1` Missing uninstall scripts | U11 | `devops/uninstall.sh` and `.ps1` exist and share the Go uninstall plan contract. | `ACC-F1`: native script syntax plus behavior/golden-plan parity on Windows and Unix. |
| `F2` Wrong installer selector interface | U09 | Bash `--version/--channel/--wiki`; PowerShell `-Version/-Channel/-Wiki`; wiki is required and selectors are mutually exclusive. | `ACC-F2`: shared selector vectors test exact stable/prerelease/channel/default, missing wiki, and every invalid combination before writes. |
| `F3` Incomplete catalog selection | U09, U10 | Install/upgrade exhaust pagination and reject drafts/malformed/incomplete/incompatible releases. | `ACC-F3`: multi-page fake GitHub catalog proves highest exact channel selection and zero mutation when no compatible release exists. |
| `F4` Obsolete exact-11 release | U07 | Release set is derived from complete manifest and includes payloads, launchers, resources, schemas/examples, manifest, checksums. | `ACC-F4`: build inventory equals manifest-derived expected set; hard-coded count search is empty outside historical docs. |
| `F5` Missing ownership manifests | U05, U07 | Strict release and installed manifests record IDs, versions, paths, checksums, projections, and ownership. | `ACC-F5`: round-trip/schema/adversarial path tests and install/upgrade/uninstall all consume manifests rather than prefixes. |
| `F6` Partial checksum verification | U07, U09 | Every downloaded release artifact is uniquely checksummed; archive members are manifest-verified. | `ACC-F6`: missing, duplicate, malformed, mismatched, extra, and truncated checksum fixtures all fail before mutation. |
| `F7` Wrong staging/bin paths | U05, U09 | Operations stage under validated Buda children of `.guiho/.temp`; stable launcher lives in `.guiho/bin`. | `ACC-F7`: native path tests, strict-descendant cleanup attacks, user-level PATH idempotency, and other-CLI sentinels. |
| `F8` No stable launcher/immutable payload | U06 | Launcher, strict `current.json`, active/previous pointers, immutable versions, and guarded fallback exist. | `ACC-F8`: native forwarding/exit/fallback/pointer security suite and post-upgrade immediate next-invocation version. |
| `F9` Candidate validated after activation/no self-test | U06, U09 | Raw version, format/target, resources, and hidden self-test run in staging before pointer activation. | `ACC-F9`: corrupt/wrong-target/wrong-version/missing-resource candidates never alter installed state. |
| `F10` Incomplete install transaction/repair | U05, U09 | Complete pointer/artifact/projection rollback and same-version repair exist; persistent state is preserved. | `ACC-F10`: phase-by-phase failure injection, repair missing/retired artifacts, config/OKF byte invariance, and old installation usability. |

### G. Upgrade lifecycle

| Finding | Unit | Implementation evidence | Acceptance evidence |
|---|---|---|---|
| `G1` Missing first/final recovery block | U10 | Executable installer command prints after local validation and again for every outcome, exact-pinned after resolution. | `ACC-G1`: text and JSON tests for success/up-to-date/dry-run/rollback/failure/default/version/channel, and no network call before first block. |
| `G2` Binary-only upgrade | U10 | Upgrade selects/applies the complete release manifest and removes retired projections. | `ACC-G2`: old/new manifests with added/changed/retired artifacts result in one consistent installed version. |
| `G3` Scheduled Windows outcome | U06, U10 | Pointer activation is synchronous; `Scheduled` model/helper authority is removed. | `ACC-G3`: native Windows upgrade returns only after launcher reports target; repository/API search contains no successful `scheduled` state. |
| `G4` Unix executing-path replacement | U06, U10 | Running payload remains immutable; only new version dir and pointer change. | `ACC-G4`: native Unix inode/path test shows executing payload untouched and new invocation selects candidate. |
| `G5` Missing lock/journal/process safety | U05, U06, U10 | Token lock, phase journal, interrupted recovery, registry, user/path-verified process handling exist. | `ACC-G5`: concurrency, stale lock, PID reuse, unrelated same-name process, qmd child, crash-at-phase, and unterminable-old-instance tests. |
| `G6` Incomplete verification/rollback | U10 | Launcher version/self-test, complete artifact/projection verification, pointer and resource rollback exist. | `ACC-G6`: injected failures after every mutation restore byte-equivalent old active state and leave journal recoverable/complete. |
| `G7` No post-upgrade init | U08, U10 | Successful agent-managed upgrade runs one explicit-wiki init and raw version verification; init-only failure is distinct. | `ACC-G7`: call-count/order tests, init failure leaves new verified release active, output/JSON reports both statuses. |
| `G8` No upgrade channel | U10 | Upgrade exposes channel selection using the same exact selector semantics as installer. | `ACC-G8`: channel/default/version/mutual-exclusion/full-pagination tests and recovery command preservation. |

### H. Uninstall lifecycle

| Finding | Unit | Implementation evidence | Acceptance evidence |
|---|---|---|---|
| `H1` Wrong uninstall flags | U11 | All interfaces expose preserve-config, preserve-data, dry-run, yes; keep-agent-resources removed. | `ACC-H1`: Cobra/Bash/PowerShell flag inventory and invalid/combined option tests. |
| `H2` No grouped plan/confirmation | U11 | Exact `REMOVE`/`PRESERVE` plan prints before mutation; TTY confirmation or explicit yes required. | `ACC-H2`: dry-run zero-write, interactive accept/decline, noninteractive refusal, and plan parity snapshots. |
| `H3` Incomplete default scope | U11 | All Buda-owned active artifacts/config/data/state/projections are planned; canonical OKF is explicitly excluded. | `ACC-H3`: synthetic full install removed; preserve modes retain exact classes; OKF and shared/qmd-owned sentinels unchanged. |
| `H4` Unsafe/nontransactional ownership | U05, U11 | Manifest proof precedes deletion; rollback/quarantine covers all mutable surfaces; shared parents remain. | `ACC-H4`: corrupt manifest, symlink/reparse, mid-removal failure, neighboring CLI, unmanaged AGENTS content, and restored old install tests. |
| `H5` Asynchronous Windows removal | U06, U11 | Native Windows canonical paths are removed/quarantined within the original operation; cleanup is not success authority. | `ACC-H5`: Windows locked launcher/payload acceptance verifies no active canonical Buda install when command/script returns success. |

### I. Documentation, instructions, workflows, and tests

| Finding | Unit | Implementation evidence | Acceptance evidence |
|---|---|---|---|
| `I1` Missing final README uninstall section | U12 | README ends with exact remote uninstall commands, destructive scope, dry-run, and combined preservation example. | `ACC-I1`: heading/order/content test and native command walkthrough. |
| `I2` Stale release documentation | U00, U12 | Historical 0.0.2 plan is labeled historical; current docs are release-neutral or target-current. | `ACC-I2`: link/version scan distinguishes historical references from current instructions; XDocs descriptions match. |
| `I3` CI/publication enforce old behavior | U07, U13 | CI/release gates use manifest-derived assets, raw versions, launcher lifecycle, and complete tooling. | `ACC-I3`: exact-head CI plus downloaded workflow-artifact revalidation; no hard-coded 11/prefixed-version current gate. |
| `I4` Tests assert forbidden behavior | U02-U13 | Obsolete assertions are replaced by Convention 0001 acceptance IDs. | `ACC-I4`: repository search confirms no current test expects depth 1, prefixed version, scheduled success, direct mutable install, or exact 11. |
| `I5` Local/shared instructions conflict | U00, U14 | Buda AGENTS and mandatory shared Go CLI skill align with the convention before release. | `ACC-I5`: instruction review/search and final exact-head agent-context audit produce no contradictory lifecycle command or artifact rule. |

## Aggregate validation gates

| Gate | Required command/evidence | Pass condition |
|---|---|---|
| `GATE-01` Formatting | RunX-cataloged gofmt check | No tracked Go file reported. |
| `GATE-02` Modules | RunX-cataloged `go mod tidy -diff` | No `go.mod`/`go.sum` drift. |
| `GATE-03` Go tests | `go test -count=1 ./...` with isolated cache if required | All packages pass. |
| `GATE-04` Vet | `go vet ./...` | All packages pass. |
| `GATE-05` Tooling | Mirror config, RunX check/list, strict XDocs meta/tree/doctor | Every command passes; XDocs has one complete root. |
| `GATE-06` Schemas | Schema generation/parity/examples | Generated files clean; both examples validate offline. |
| `GATE-07` Release build | Manifest-derived payload/launcher/resource build | Complete declared set, no undeclared/missing artifact, all checksums pass. |
| `GATE-08` Native lifecycle | Installer/launcher/upgrade/uninstall suites on supported native runners | Every available native platform passes; foreign targets explicitly build-only. |
| `GATE-09` Legacy migration | Start from official 0.1.1 layout and run new installer | New launcher active; legacy owned files removed; config/OKF preserved. |
| `GATE-10` Transaction recovery | Crash/failure injection at every journal phase | Old or new release is wholly active; no mixed state; next invocation repairs journal. |
| `GATE-11` Documentation | README commands, links, XDocs, help-doc snapshots | Docs match executable behavior and README's final operational heading is Uninstall. |
| `GATE-12` Review | Independent implementation review and validation on exact head | No unresolved actionable finding; evidence commit matches PR head. |
| `GATE-13` Mirror plan | `mirror version current` and clean `mirror version plan minor` | The approved 0.2.0 target and all mutations are complete and expected; apply not run. |
| `GATE-14` Release authorization | Separate explicit user authorization | Only after authorization may Mirror apply/tag/push and GitHub Release verification run. |

## Evidence retention

For each unit, retain in the PR or validation report:

- exact commit SHA;
- RunX selectors and rendered commands;
- test/vet/tooling outcomes;
- native runner/architecture and build-only labels;
- artifact/manifest/checksum inventories;
- filesystem before/after and preservation evidence for lifecycle work;
- review findings and resolutions; and
- explicit confirmation that no production action or unauthorized release
  action occurred.

This matrix is complete only when all audit rows, product invariants, and
aggregate gates are satisfied on the same aggregate commit.
