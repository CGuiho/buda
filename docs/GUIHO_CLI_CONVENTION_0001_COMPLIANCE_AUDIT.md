---
name: buda-guiho-cli-convention-0001-compliance-audit
purpose: Evaluate Buda against GUIHO CLI Convention 0001 and record every evidenced compliance gap.
description: Aspect-by-aspect audit of the Buda Go/Cobra CLI, tooling, configuration, agent resources, installation, upgrade, uninstall, documentation, and validation contracts against GUIHO CLI Convention 0001.
created: 2026-08-16
owner: buda-docs
flags: []
tags:
  - audit
  - cli
  - convention
  - compliance
keywords:
  - Buda
  - GUIHO CLI Convention 0001
  - Go
  - Cobra
  - Mirror
  - RunX
  - XDocs
  - noncompliance
---

# Buda compliance audit against GUIHO CLI Convention 0001

## Verdict

**Buda does not obey GUIHO CLI Convention 0001.**

Buda has a strong pre-convention Go/Cobra foundation: it constructs a fresh,
testable command tree; strictly decodes its existing project YAML; centralizes
errors and exit codes; builds the eight expected pure-Go target binaries; uses
deterministic checksums; paginates the GitHub release catalog in the CLI; and
has substantial unit and integration coverage. Those qualities are worth
preserving.

They are not sufficient for this convention. The repository fails mandatory
contracts in every lifecycle layer added by the convention: RunX, raw version
output, developer-help semantics, dual configuration and schemas,
`agent.evolution`, the required agent command names, the setup prompt, the
common `init` sequence, complete release manifests, canonical installation
paths, the stable launcher and immutable payload layout, synchronous upgrades,
safe uninstallers, and the required uninstall documentation. Several current
tests and workflows actively enforce the superseded behavior.

This is an architecture-level noncompliance, not a small documentation drift.
A conforming release should not be published until the installation, upgrade,
and uninstall architecture is replaced and convention-level acceptance tests
are green.

## Audit baseline and method

| Item | Audited state |
|---|---|
| Convention | `C:\GUIHO\guiho\docs\conventions\guiho-convention-0001-cli.md` at commit `bd7dc4583fbc0722d3f05b78f7c70c7a970865e2` |
| Buda | `C:\GUIHO\buda` at commit and tag `09cba7aa43d09e8dbbcd863d71e38ebcbc835712` / `buda/v0.1.1` |
| Repository state before the report | `main...origin/main`, clean |
| Toolchain observed | `go1.26.5 windows/amd64`; Cobra module `v1.10.2` |
| Evidence | Full convention review; repository instructions; source, tests, workflows, tooling manifests, release tooling, agent resources, README, and XDocs inspection; command-tree and configuration tracing; independent focused reviews |

The audit treats every **must**, **required**, and **mandatory** statement in the
convention as normative. A feature is not credited merely because a similarly
named feature exists: its behavior, filesystem layout, output, transaction
boundary, and failure semantics must match. Absence searches covered all
tracked source files while respecting repository ignore rules.

The Go version conclusion is time-based. Buda's foundation commit is dated
2026-07-26, while Go 1.26.5 was released on 2026-07-07. The module nevertheless
declares Go 1.23.0. See the official [Go release history](https://go.dev/doc/devel/release).
Cobra 1.10.2 is the latest listed official release and is correctly selected;
see [Cobra releases](https://github.com/spf13/cobra/releases).

## Status summary

| Area | Status | Summary |
|---|---|---|
| Technology stack | **Partial** | Cobra and native Go tooling comply; the creation-time Go version does not. |
| Mandatory project tooling | **Noncompliant** | Mirror config loads, but RunX is absent and XDocs/version ownership is incomplete. |
| Cobra tree and general flags | **Partial** | Fresh tree, long flags, aliases, streams, and exit mapping are sound; required agent command names are wrong. |
| Version and developer help | **Noncompliant** | Version output, help-doc scope, depth grammar, and global-flag rendering all violate the convention. |
| Configuration and schemas | **Noncompliant** | There is one unmerged config, no global config, no evolution policy, and no schemas/examples. |
| `init` | **Noncompliant** | Buda-specific initialization exists, but the mandatory common reconciliation sequence does not. |
| Agent resources | **Noncompliant** | The wiki skill/instruction exist, but evolution guidance, setup prompt, definitions, confirmation evidence, and version parity are absent. |
| Installation and release artifacts | **Noncompliant** | The repository implements the older direct-binary 11-asset model, not the convention's managed installation. |
| Upgrade | **Noncompliant** | Upgrade is partial, in-place, and asynchronous on Windows; required recovery, locking, journaling, activation, and rollback contracts are absent. |
| Uninstall | **Noncompliant** | Both scripts are missing; the CLI removal command lacks scope, preservation, planning, confirmation, and transaction safety. |
| Documentation and tests | **Noncompliant** | README lacks the mandatory final uninstall section; documentation is version-stale; tests/workflows assert obsolete behavior. |

## Detailed noncompliance findings

### A. Technology and mandatory project tooling

#### A1. The project was created against an obsolete Go language version

- **Convention:** use the latest Go version available when the project is
  created.
- **Evidence:** `go.mod:3` declares `go 1.23.0`. The foundation commit
  `d810c8aafd2f57ac109a30b07ad3fa17ef9a0765` is dated 2026-07-26. Go 1.26.5
  was available from 2026-07-07, and the audit host uses Go 1.26.5.
- **Gap:** the creation-time language baseline was already three Go feature
  releases behind.
- **Required correction:** move the module and CI baseline to the approved
  current Go release, then rerun formatting, tidy, tests, vet, native builds,
  and all cross-build checks.

#### A2. The mandatory RunX catalog is absent

- **Convention:** every repository must have a root `runx.yaml` cataloging
  every repeatable command with a namespace, stable `uid`, unique `id`, exact
  command, and short description.
- **Evidence:** `runx.yaml` does not exist. A `runx check --format json` probe
  cannot validate a catalog because no catalog is present.
- **Gap:** repeatable Go, XDocs, Mirror, installer, build, smoke, and release
  operations have no canonical command inventory or stable selectors.
- **Required correction:** add and validate a complete root catalog, including
  at least format, tidy, test, vet, builds, XDocs checks, Mirror checks/plans,
  installer tests, lifecycle acceptance, and release-contract verification.

#### A3. Mirror is loadable but is not the sole authority for all versioned surfaces

- **Convention:** `mirror.yaml` must declare Mirror as the sole project-version
  authority, cover every version-bearing file, create release commits/tags,
  and generate the changelog.
- **Evidence:** `mirror.yaml:1-22` passes `mirror config check`, uses Git SemVer,
  enables commits/pushes, and writes `CHANGELOG.md`. However its only version
  output is Git (`mirror.yaml:5-9`), while `mirror.yaml:10-14` retains
  `package.json` and `jsr.json` paths that do not exist in this Go repository.
  Version-bearing artifacts remain outside
  that mapping: `skills/guiho-s-0002-buda/SKILL.md:6,18` and
  `prompts/guiho-i-buda.md:6` are still `0.1.0` while the current tag is
  `buda/v0.1.1`; README and implementation documentation still prescribe
  `0.0.2` (`README.md:67-76,97-111`; `docs/IMPLEMENTATION.md:4-22,47,84-106`).
- **Gap:** Git can advance without synchronizing bundled resources and durable
  docs. The current repository proves that this has already happened.
- **Required correction:** make Mirror update or validate every project-version
  surface and reject release plans when embedded resource metadata, schema
  URLs, examples, README commands, descriptors, or workflow fixtures drift.

#### A4. The XDocs topology and metadata do not fully describe current ownership

- **Convention:** `xdocs.yaml` must cover every project-owned directory and
  maintain a complete descriptor tree plus generated root index.
- **Evidence:** the root descriptor's children omit `.github`
  (`buda.xdocs.md:5-11`), while `.github/github.xdocs.md:4-7` declares
  `parent: null`. `XDOCS.md:8-19` manually lists `.github`, producing two
  descriptor roots rather than one ownership tree. The root file catalog at
  `buda.xdocs.md:12-21` also omits tracked `mirror.yaml`. The docs descriptor still
  describes the `0.0.2` delivery (`docs/docs.xdocs.md:3,9,19`), and workflow
  metadata still promises exactly eleven assets
  (`.github/workflows/workflows.xdocs.md:3-8,24-35`).
- **Gap:** XDocs can parse the files, but the durable topology and descriptions
  are incomplete and stale relative to the current repository and convention.
- **Required correction:** attach `.github` to `buda-package`, refresh all
  stale version/lifecycle descriptions, include this audit in `buda-docs`, and
  require strict meta/tree/doctor validation in the command catalog and CI.

### B. Cobra command tree, version, flags, and developer help

#### B1. `--version` and `-v` do not print raw SemVer

- **Convention:** the top-level version flags must print only the raw SemVer,
  such as `1.2.3`.
- **Evidence:** `cmd/root.go:251` sets
  `{{.Name}} v{{.Version}}\n`, producing `buda v1.2.3`. That form is asserted
  by `cmd/root_test.go:56-58`, both installers
  (`devops/install.sh:208-211`; `devops/install.ps1:177-179`), CI
  (`.github/workflows/ci.yml:49-53,108-123`), and publication
  (`.github/workflows/publish.yml:125-133`).
- **Gap:** the public contract and all lifecycle verification use a format the
  convention explicitly forbids.
- **Required correction:** emit raw SemVer and update every parser, assertion,
  installer, smoke test, and workflow in the same change.

#### B2. `--help-docs` emits a descendant documentation set, not only the current command

- **Convention:** every command must expose `--help-docs` that prints Markdown
  documentation for the current command.
- **Evidence:** the flag is correctly persistent at `cmd/root.go:258-263`, but
  `internal/help/help.go:93-123` recursively visits all public descendants and
  joins separate generated pages.
- **Gap:** `buda --help-docs` and group-command invocations produce multiple
  commands' pages rather than the requested command's document.
- **Required correction:** render the invoked command only, keeping output
  deterministic and side-effect free.

#### B3. `--help-tree-depth` has the wrong type, default, and validation

- **Convention:** the value must be `max` or an integer greater than `1`; the
  visible default is `max`.
- **Evidence:** `cmd/root.go:176,262` stores an integer with default `0`, cannot
  parse `max`, and `cmd/root.go:186-189` accepts `1`. The test at
  `cmd/root_test.go:77-90` treats depth `1` as valid.
- **Gap:** neither the user-facing grammar nor its validation matches the
  convention, even though internal depth `0` happens to mean unlimited.
- **Required correction:** introduce a validated `max | integer > 1` value,
  expose `max` as the default, and test all accepted and rejected forms.

#### B4. `--help-tree-global-flags` is missing and inherited flags repeat by default

- **Convention:** every command must have the presence Boolean
  `--help-tree-global-flags`. With it absent, global flags appear once at the
  initial command; with it present, they repeat under descendants.
- **Evidence:** the flag list at `cmd/root.go:258-263` has no such option.
  `internal/help/help.go:48-81` unconditionally adds both local and inherited
  flags for every rendered command.
- **Gap:** the default is the opposite of the required behavior, and there is
  no user control.
- **Required correction:** carry initial-command/global-display state through
  the renderer and add presence-only tests at root and nested scopes.

#### B5. The required agent resource commands use the prohibited name `update`

- **Convention:** commands must be
  `agent skill install|uninstall|upgrade|list|show` and
  `agent instruction apply|remove|upgrade|show`; `update` must not be used in
  either tree.
- **Evidence:** `cmd/agent.go:40-63` creates skill `update`, and
  `cmd/agent.go:117-140` creates instruction `update`; neither tree has
  `upgrade`. The obsolete form is also prescribed by `cmd/doctor.go:52-65`,
  `cmd/upgrade.go:271-275`, and both installers.
- **Gap:** two required commands are absent and two explicitly forbidden
  commands are public.
- **Required correction:** replace the command names atomically across the
  CLI, docs, agents, installer, upgrade reconciler, tests, and workflows.

### C. Configuration, schemas, and agent-evolution policy

#### C1. Buda has no separate global and project configuration contracts

- **Convention:** use project `buda.yaml` and global
  `$HOME/.guiho/buda/buda.global.yaml`, with separate schemas.
- **Evidence:** `internal/config/config.go:15-18` defines only
  `FileName = "buda.yaml"`. Its resolver checks project `buda.yaml` and then
  `$HOME/.guiho/buda/buda.yaml` (`internal/config/config.go:50-83`).
- **Gap:** the global file has the wrong name and is treated as a fallback copy
  of the project shape, not a distinct baseline.
- **Required correction:** define typed global and project models, paths,
  loaders, validation, examples, and ownership boundaries.

#### C2. Configuration is explicitly not merged

- **Convention:** global configuration supplies the baseline and project
  configuration overrides it per field.
- **Evidence:** `internal/config/config.go:20-21` states that configuration is
  never merged; `Resolve` returns the first existing file. Repository opening
  loads only the selected project config (`internal/repository/repository.go:69-72`).
- **Gap:** inheritance and effective-value calculation do not exist.
- **Required correction:** implement deterministic per-field merge semantics,
  retaining strict decoding independently at both layers and validating the
  effective result.

#### C3. The mandatory `agent.evolution` contract is absent

- **Convention:** four policy values—`self_upgrade`, `automatic_bug_report`,
  `automatic_improvement_suggestion`, and `automatic_review`—must accept only
  `disabled`, `always-ask`, or `always-proceed`, default globally to
  `always-ask`, and support project overrides.
- **Evidence:** `internal/config/config.go:22-34` has only schema, wiki, bundle,
  qmd, and derived fields. Repository-wide searches found no evolution policy,
  enum values, inheritance, prompting, upgrade-policy execution, issue-policy
  execution, or effective-policy reporting.
- **Gap:** all mandatory AI-managed upgrade, bug, improvement, and review
  behavior is undefined.
- **Required correction:** add typed enums, strict semantic validation,
  defaults, merging, init questions, runtime policy gates, tests, schemas, and
  agent guidance together.

#### C4. Required JSON Schemas, examples, and version-pinned schema comments are absent

- **Convention:** ship separate `buda.schema.json` and
  `buda.global.schema.json`, complete valid examples, public release assets,
  embedded offline validation, and version-pinned HTTPS schema URLs at the top
  of generated YAML.
- **Evidence:** neither schema nor a global/project example pair exists.
  `internal/config/config.go:173-181` uses ordinary YAML marshaling, and
  `internal/repository/initialize.go:79-86` writes those bytes without a schema
  comment. The release builder has no schema/example assets.
- **Gap:** editor discovery, release-pinned validation, global configuration
  validation, and artifact completeness are missing.
- **Required correction:** author strict non-secret schemas and examples,
  embed the exact release copies, emit pinned schema comments, and make all
  config-creating commands validate against the same offline contract.

#### C5. Existing strict YAML validation is only a partial substitute

- **Convention:** runtime must validate locally without network access.
- **Evidence:** the existing project decoder correctly uses
  `KnownFields(true)`, rejects multiple documents, and performs semantic path
  validation (`internal/config/config.go:99-170`).
- **Gap:** this good typed validator covers only the old project shape. It does
  not cover the missing global shape, merge, evolution policy, examples, or
  release schema parity.
- **Required correction:** retain the current strictness while expanding it to
  the complete dual-config contract and add structural parity tests between Go
  types, embedded schemas, and published schemas.

### D. Mandatory `init` reconciliation

#### D1. `init` does not execute the mandatory common sequence

- **Convention:** `init` must reconcile project root, all bundled skills,
  `AGENTS.md`, the bounded instruction, global config, project config,
  evolution defaults, setup questions, CLI-specific state, final validation,
  and absolute-path summary in a fixed sequence.
- **Evidence:** `cmd/init.go:19-87` validates `--wiki-id`, initializes the Buda
  wiki/qmd state, installs global skills, and applies instructions. It never
  creates or validates `buda.global.yaml`, schemas, schema comments, merged
  effective config, evolution policy, canonical installed artifacts, or the
  full common final check.
- **Gap:** a Buda-specific initializer is presented as successful without
  satisfying most convention-wide prerequisites.
- **Required correction:** make common reconciliation the outer transaction,
  with Buda's current repository/qmd initialization as the CLI-specific step.

#### D2. Existing values and interactive/noninteractive setup semantics are wrong

- **Convention:** read existing values first, ask only unanswered questions in
  an interactive terminal, explain scope and policy choices, recommend
  `always-proceed`, and fail clearly if required input is unavailable
  noninteractively.
- **Evidence:** `cmd/init.go:20-27` rejects a missing `--wiki-id` immediately and
  constructs defaults before reading existing state. There is no terminal
  detection, question flow, policy explanation, global/project choice, or
  preservation of answered common setup values. Even an initialized wiki
  requires the immutable ID to be supplied again.
- **Gap:** interactive users are not asked; noninteractive failure covers only
  one Buda field; global and evolution questions do not exist.
- **Required correction:** load both configs first, derive unanswered items,
  separate global/project ownership, validate answers before writes, and keep
  unattended execution fail-closed.

#### D3. `init` does not always create the mandatory `AGENTS.md`

- **Convention:** ensure a project-level `AGENTS.md`, then reconcile the one
  bounded instruction block while preserving other content.
- **Evidence:** instruction target resolution creates `AGENTS.md` only when
  neither `AGENTS.md` nor `CLAUDE.md` exists
  (`internal/agent/instruction.go:174-190`). If only `CLAUDE.md` exists, Buda
  updates it and leaves `AGENTS.md` absent.
- **Gap:** a valid existing alternate instruction file suppresses the required
  canonical file.
- **Required correction:** always reconcile `AGENTS.md`; treat any additional
  projections as explicit, manifest-owned secondary copies.

#### D4. Final validation and the human summary are incomplete

- **Convention:** report actions plus absolute paths and never report success
  unless every common and CLI-specific check passes.
- **Evidence:** `cmd/init.go:77-86` prints wiki, bundle, qmd directory,
  collection, ID, counts, and generic "agent resources: reconciled". It omits
  the project/global config paths, schemas, `AGENTS.md`, skill destinations,
  effective policies, canonical artifact location, and individual actions.
- **Gap:** output cannot prove the required state, and `init` reports success
  despite not checking it.
- **Required correction:** run one aggregate final verification and report all
  absolute owned/resulting paths and action states only after it passes.

### E. Agent skills, instructions, prompts, and definitions

#### E1. The main skill lacks the exact evolution and feedback section

- **Convention:** the main skill must contain the exact heading
  `## CLI Evolution and Feedback` and explain repository/issue URLs, bug versus
  improvement versus review classification, all policy values, upgrade checks,
  upgrade and rerun-init flow, version verification, issue creation, and final
  issue URL reporting.
- **Evidence:** `skills/guiho-s-0002-buda/SKILL.md:21-59` contains governing
  references, explicit-wiki routing, and wiki workflows only. It contains none
  of the mandatory evolution text or policy values.
- **Gap:** agents cannot follow the convention's managed feedback or upgrade
  lifecycle from the shipped skill.
- **Required correction:** add the exact section with Buda's canonical URLs and
  commands after the evolution policies exist.

#### E2. There is no main install/setup prompt

- **Convention:** ship at least one user-confirmed main setup prompt explaining
  what the CLI is, how to install it, verify it, and upgrade it; additional
  prompts use the same confirmed family prefix.
- **Evidence:** the only prompt resource is `prompts/guiho-i-buda.md`, whose
  purpose is a bounded repository instruction (`prompts/guiho-i-buda.md:2-18`).
  Runtime exposes it as prompt ID `buda`
  (`internal/agent/resources.go:16-20,120-140`). It contains no installation,
  version verification, or upgrade workflow.
- **Gap:** an instruction template is mislabeled as the prompt catalog's only
  prompt, and no setup prompt exists.
- **Required correction:** obtain durable confirmation of the main prompt name,
  add the setup prompt (normally `guiho-p-buda`), and keep the instruction as a
  separately typed resource.

#### E3. Agent definitions and confirmation evidence are absent

- **Convention:** the complete selected release includes all confirmed skills,
  prompts, agent definitions, and the one managed instruction; main names are
  confirmed with the user rather than inferred.
- **Evidence:** `internal/agent/resources.go:47-58` embeds only a skill and the
  instruction/prompt filesystem. No agent-definition family or durable naming
  confirmation record exists.
- **Gap:** the artifact inventory cannot demonstrate a complete confirmed agent
  family.
- **Required correction:** record confirmations in durable requirements or
  decisions, add any selected definitions, and enumerate every resource in the
  release and installed manifests.

#### E4. Bundled agent resource versions lag the current release

- **Convention:** the installer must source and validate the selected release's
  exact versioned resources, and Mirror must govern version-bearing files.
- **Evidence:** current tag `buda/v0.1.1` contains skill and prompt metadata at
  `0.1.0` (`skills/guiho-s-0002-buda/SKILL.md:6,18`;
  `prompts/guiho-i-buda.md:6`).
- **Gap:** the embedded/released resource version is not equal to the project
  release and is not automatically synchronized.
- **Required correction:** make one release version authoritative and verify
  binary, manifest, schemas, skills, prompts, definitions, and instruction
  metadata against it before publication.

### F. Installation selection, artifacts, and filesystem architecture

#### F1. Two of the four mandatory lifecycle scripts are missing

- **Convention:** root CLI lifecycle requires `devops/install.sh`,
  `devops/install.ps1`, `devops/uninstall.sh`, and `devops/uninstall.ps1`.
- **Evidence:** the installers exist; both uninstall scripts are absent. The
  devops descriptor inventory confirms only installer/build tooling
  (`devops/devops.xdocs.md:6-12`).
- **Gap:** script-based uninstall is impossible and cannot share semantics with
  the Cobra command.
- **Required correction:** add both uninstallers around the same manifest,
  planning, preservation, confirmation, and rollback contract as the CLI.

#### F2. Installer selector interfaces do not match the convention

- **Convention:** Bash accepts `--version <semver>` and
  `--channel <name>`; PowerShell accepts `-Version <semver>` and
  `-Channel <name>`; the selectors are mutually exclusive; omission selects
  the latest stable release.
- **Evidence:** Bash reads one positional token (`devops/install.sh:24`) and
  expects a full `buda/v...` tag (`devops/install.sh:50-68`). PowerShell exposes
  `-Version` but no `-Channel` (`devops/install.ps1:1-4`) and likewise expects
  the canonical tag (`devops/install.ps1:34-49`).
- **Gap:** raw SemVer selection, channels, mutual exclusion, and equivalent
  cross-shell behavior do not exist.
- **Required correction:** implement one selection algorithm with shell-native
  flag syntax and identical outcomes.

#### F3. Installers do not inspect the complete paginated release catalog

- **Convention:** paginate the full public catalog; reject drafts, malformed
  tags, missing/malformed manifests and checksums, missing required artifacts,
  and ambiguous channel selection; select the highest matching SemVer.
- **Evidence:** both scripts use GitHub's single `releases/latest` endpoint for
  default selection (`devops/install.sh:50-59`;
  `devops/install.ps1:34-40`). Neither paginates or filters channels. The CLI's
  separate catalog does paginate and filter tags
  (`internal/selfmanage/catalog.go:129-190`), but `resolveUpgrade` only requires
  the current binary and `checksums.txt` (`cmd/upgrade.go:224-257`).
- **Gap:** no install path validates release completeness under the new
  convention.
- **Required correction:** share one strict catalog/selection implementation
  or equivalent test vectors across Bash, PowerShell, and Cobra.

#### F4. The release set is the obsolete 11-artifact contract

- **Convention:** the selected release must include all supported binaries,
  launcher artifacts, all skills/prompts/definitions/instruction, schemas,
  examples, required operational resources, `artifacts.json`, and checksums.
- **Evidence:** `devops/build-binaries.go:20-24,34-43,84-101` publishes eight
  payload binaries, one skill ZIP, `guiho-i-buda.md`, and `checksums.txt`, then
  fails unless the total is exactly eleven. There are no launchers, setup
  prompt, agent definitions, schemas, examples, or artifact manifest.
- **Gap:** the release cannot install or prove the complete convention-managed
  CLI.
- **Required correction:** define the complete inventory first, generate a
  deterministic manifest, checksum every declared release file, and derive
  workflow expectations from the manifest rather than a fixed count.

#### F5. `artifacts.json` and the installed manifest are absent

- **Convention:** `artifacts.json` must declare paths, IDs, versions, checksums,
  installed locations, projections, and ownership. The installed manifest must
  be stored under `$HOME/.guiho/buda/` and drive reinstall, upgrade, and
  uninstall.
- **Evidence:** neither `artifacts.json` nor `installed-artifacts.json` exists.
  Agent ownership is inferred from hard-coded IDs and paths
  (`internal/agent/skill.go:37-44,91-101`).
- **Gap:** there is no authoritative inventory, retired-artifact detection,
  path-ownership proof, or safe uninstall source.
- **Required correction:** introduce versioned release and installed manifests,
  strict schema validation, safe relative paths, and projection records.

#### F6. Checksums validate only an incomplete subset

- **Convention:** every selected-release artifact must be declared and verified
  before activation; malformed, missing, or ambiguous checksum data fails.
- **Evidence:** installers fetch and verify only the platform binary and skill
  ZIP (`devops/install.sh:113-137`; `devops/install.ps1:87-103`). They do not
  fetch or validate the released instruction, let alone the missing manifest,
  launchers, prompts, definitions, schemas, and examples. PowerShell and the Go
  checksum reader accept the first matching entry rather than rejecting a
  duplicate (`devops/install.ps1:94-101`;
  `internal/selfmanage/upgrade.go:99-111`).
- **Gap:** checksum success does not prove release completeness or unambiguous
  ownership.
- **Required correction:** validate one strict checksum entry for every
  manifest file and reject all undeclared, duplicate, missing, malformed, or
  mismatched entries.

#### F7. Installation uses the wrong staging and shared binary locations

- **Convention:** stage under
  `$HOME/.guiho/.temp/buda-install-<unique>/` and install the stable launcher in
  `$HOME/.guiho/bin/`.
- **Evidence:** Bash uses system `mktemp -d` and defaults to `~/.local/bin`
  (`devops/install.sh:25,73`). PowerShell uses `Path.GetTempPath()` and defaults
  to `%LOCALAPPDATA%\GUIHO\bin`
  (`devops/install.ps1:1-4,53-58`). Both permit arbitrary install directories.
- **Gap:** shared GUIHO containment, strict-child validation, and cross-platform
  layout equivalence are absent.
- **Required correction:** canonicalize and validate GUIHO home paths before any
  write or cleanup; make `$HOME/.guiho/bin` the one PATH target.

#### F8. There is no stable launcher, immutable payload, or activation pointer

- **Convention:** PATH resolves a stable launcher; versioned payloads live under
  `$HOME/.guiho/buda/versions/<version>/`; atomic `current.json` records active
  and previous relative payloads; the launcher has guarded fallback.
- **Evidence:** installers move the downloaded payload directly to the PATH
  command (`devops/install.sh:194-203`; `devops/install.ps1:164-172`).
  `main.go:17-29` is the payload entrypoint. There is no launcher source,
  `versions/`, `current.json`, pointer protocol, or fallback implementation.
- **Gap:** the core filesystem and process model required for safe upgrades does
  not exist.
- **Required correction:** build small platform launchers, define and validate
  the pointer protocol, store immutable version directories, and retain the
  previous verified payload until a later successful upgrade.

#### F9. Installation does not validate a complete candidate before activation

- **Convention:** validate all candidate files, binary format, target, exact raw
  version, executable permission, and a hidden installation self-test before
  activation; then run the launcher version check, self-test, and `init`.
- **Evidence:** installers replace the executable first and only then run
  `--version` (`devops/install.sh:194-211`;
  `devops/install.ps1:164-179`). No hidden installation self-test exists.
  Neither installer invokes `buda init` as its final lifecycle step.
- **Gap:** an invalid payload can become active before validation, and the
  required post-install common reconciliation never occurs.
- **Required correction:** validate inside the staged version directory,
  activate atomically only after all checks pass, verify through the launcher,
  then run `init` once without rolling back a valid installation if init alone
  fails.

#### F10. Installation is not a complete transaction and cannot safely reinstall

- **Convention:** installation/reinstallation must replace all versioned
  artifacts transactionally, remove retired manifest-owned projections, roll
  back every mutated surface, preserve configuration/data/database state, and
  support same-version repair.
- **Evidence:** the scripts back up only the binary
  (`devops/install.sh:73-89,194-211`;
  `devops/install.ps1:54-58,164-179,234-242`). Skill mutation occurs after
  binary activation; later failure restores only the executable. Optional
  projections are explicitly nonfatal. No manifest classifies persistent or
  disposable state.
- **Gap:** resource versions can diverge from the binary, failed installs can
  leave partial projections, retired files cannot be identified, and
  preservation cannot be proven.
- **Required correction:** snapshot all managed projections and pointer state,
  apply the complete selected manifest, validate, and restore every affected
  surface on failure while leaving config/data untouched.

### G. Upgrade lifecycle

#### G1. The mandatory recovery command is not the first and final visible output

- **Convention:** after local state validation, every upgrade path must print a
  complete reinstall command and preservation guarantee before any risky or
  network action, and print it again in the final success/failure block.
- **Evidence:** `cmd/upgrade.go:23-31` resolves a release over the network before
  visible output. Already-current, dry-run, success, and failure branches do
  not emit the required blocks (`cmd/upgrade.go:32-55,84-111`). The `Recovery`
  field is a file-copy description such as restore backup to executable
  (`internal/selfmanage/upgrade.go:137-140`), not a runnable remote installer
  command.
- **Gap:** users do not receive the required recovery path or preservation
  statement when upgrade fails.
- **Required correction:** generate the exact-version installer command from
  validated local/release state and route all outcomes through one first/final
  recovery renderer, including JSON-equivalent evidence if JSON is supported.

#### G2. Upgrade installs only one binary, not the selected release

- **Convention:** upgrade must install the complete release—launcher/payload,
  agent resources, schemas, examples, manifests, projections, and metadata.
- **Evidence:** `internal/selfmanage/upgrade.go:184-262` downloads one binary,
  rotates an adjacent `.old` backup, and replaces the current executable.
  `cmd/upgrade.go:271-284` separately updates one skill and optionally one wiki
  instruction.
- **Gap:** upgrade cannot guarantee that every installed resource is from the
  same release, cannot remove retired artifacts, and cannot repair missing
  resources.
- **Required correction:** make the release manifest—not the executable—the
  unit of selection, validation, installation, and rollback.

#### G3. Windows upgrade returns the explicitly prohibited `scheduled` outcome

- **Convention:** upgrade must complete synchronously; a detached helper is not
  authoritative for success; `scheduled` is not a successful state.
- **Evidence:** Windows copies itself into a helper, starts it detached, and
  returns before activation (`internal/selfmanage/replace_windows.go:20-47`).
  `cmd/upgrade.go:106-108` prints that the upgrade was scheduled, and
  `internal/selfmanage/upgrade.go:49,174` models `Scheduled` as a result.
- **Gap:** the invoking command can finish without knowing whether activation,
  validation, reconciliation, or cleanup succeeded.
- **Required correction:** use the stable launcher/current-pointer model so the
  command can activate a new immutable payload synchronously without replacing
  its running file.

#### G4. Unix upgrade replaces the executing payload path

- **Convention:** never replace the currently executing payload in place;
  activate an immutable candidate through `current.json`.
- **Evidence:** `internal/selfmanage/replace_unix.go:10-32` renames the current
  executable and installs the candidate at that same path.
- **Gap:** Unix follows the old mutable-binary design and has no verified
  fallback pointer.
- **Required correction:** install the candidate in a new version directory and
  atomically switch only the activation pointer.

#### G5. Locking, journaling, interrupted recovery, and peer-process handling are absent

- **Convention:** acquire an exclusive upgrade lock with ownership token,
  recover stale/interrupted transactions, journal phases, maintain an instance
  registry, verify process ownership/path, and terminate only verified old CLI
  instances before activation.
- **Evidence:** there is no lifecycle lock, token, journal, recovery-phase state,
  current/previous pointer, or process registry. Windows waits only for the
  invoking parent PID (`internal/selfmanage/replace_windows.go:43-47`); Unix has
  no peer-instance handling.
- **Gap:** concurrent/interrupted upgrades and other running instances have no
  safe, recoverable protocol.
- **Required correction:** implement these as one tested activation state
  machine with stale-state and adversarial path/process tests.

#### G6. Post-activation verification and rollback are incomplete

- **Convention:** verify launcher `--version`, hidden self-test, CLI home,
  complete resources/projections, and current pointer; on failure restore the
  old pointer and all projections/config mutations.
- **Evidence:** current verification only checks executable output, and it uses
  the wrong prefixed form (`internal/selfmanage/upgrade.go:315-325`). If agent
  reconciliation fails, `cmd/upgrade.go:87-96` rolls back only the executable.
  Skill or instruction mutations can remain paired with the old binary. There
  is no launcher, self-test, pointer, or complete projection validation.
- **Gap:** rollback is not atomic across the actual installed state.
- **Required correction:** snapshot pointer and all manifest-owned projections,
  validate the complete new state, and restore every surface on failure.

#### G7. Upgrade does not rerun the common `init` reconciliation

- **Convention:** after successful activation and resource replacement, rerun
  `init` once to reconcile global/project state; valid installation remains if
  init alone fails.
- **Evidence:** `cmd/upgrade.go:271-284` calls agent resource update operations,
  not `init`. There is no dual-config/evolution reconciliation to rerun.
- **Gap:** successful upgrades cannot repair or migrate the required common
  managed state.
- **Required correction:** implement the complete `init` contract, invoke it
  once after activation, and report init failure separately from installation
  success.

#### G8. Upgrade has no channel selector

- **Convention:** release selection is consistent across managed lifecycle
  entry points and supports an exact prerelease channel resolved from the full
  paginated catalog.
- **Evidence:** root upgrade exposes only `--version` and `--dry-run`
  (`cmd/upgrade.go:114-115`). There is no `--channel`, exact first-prerelease-ID
  matching, selector mutual exclusion, or channel-specific result reporting.
- **Gap:** even though the Go catalog can enumerate prereleases, users and
  agents cannot request the convention's channel contract during upgrade.
- **Required correction:** share the strict installer selection model with the
  upgrade command and test stable, exact-version, channel, malformed,
  incomplete, and ambiguous catalogs.

### H. Uninstall lifecycle

#### H1. Cobra uninstall exposes the wrong flags

- **Convention:** CLI and scripts share `--preserve-config`,
  `--preserve-data`, `--dry-run`, and `--yes` (with PowerShell-equivalent
  names). Replaceable resources are not preservation targets.
- **Evidence:** `cmd/uninstall.go:71-73` exposes only `--dry-run` and
  `--keep-agent-resources`.
- **Gap:** all required safety/preservation controls except dry-run are absent,
  while the one extra flag preserves content the convention says is
  replaceable.
- **Required correction:** adopt the shared flags and remove the incompatible
  preservation model.

#### H2. Uninstall has no complete grouped plan or confirmation gate

- **Convention:** print exact grouped `REMOVE` and `PRESERVE` targets, request
  confirmation unless `--yes`, and fail noninteractively without `--yes`.
- **Evidence:** dry-run output lists only executable, two skill projections,
  and optional wiki instructions (`cmd/uninstall.go:24-35`). Normal execution
  immediately removes skills (`cmd/uninstall.go:37-52`) and then the executable.
  There is no confirmation or terminal check.
- **Gap:** destructive work begins without informed approval and without a
  manifest-derived view of its scope.
- **Required correction:** generate the plan from validated installed and
  current release manifests, classify every path, and gate all mutations on
  confirmation or explicit `--yes`.

#### H3. Default uninstall scope is incomplete

- **Convention:** default uninstall removes all CLI-owned files, including
  launcher, CLI home, payloads, config, data/databases, caches, metadata,
  global skills, instruction blocks, prompts, definitions, schemas, examples,
  and project config. Preservation flags retain only documented config/data.
- **Evidence:** `cmd/uninstall.go:19-68` removes only the current executable,
  selected global skills, and optional explicit-wiki instructions. It does not
  inventory or remove CLI home, configuration, persistent data, databases,
  caches, metadata, prompts, definitions, schemas, examples, or all known
  project configs.
- **Gap:** the command called full uninstall does not implement the convention's
  destructive default or preservation exceptions.
- **Required correction:** derive complete ownership from manifests plus
  version-appropriate supplemental discovery, with an explicit persistent-state
  inventory and tests.

#### H4. Uninstall is neither transactional nor ownership-safe

- **Convention:** remove only proven Buda-owned files, preserve shared
  directories/PATH/other CLIs, and use a safe transaction or restore path.
- **Evidence:** skill projections are deleted before executable removal
  (`cmd/uninstall.go:37-55`); a later failure does not restore them. Removal is
  based on hard-coded locations or `os.Executable()` rather than a validated
  installed manifest. Unix deletion removes the supplied executable path
  (`internal/selfmanage/replace_unix.go:39-43`).
- **Gap:** partial failure can leave a damaged installation, and canonical
  ownership is not proven before deletion.
- **Required correction:** validate every target under an approved ownership
  root, stage reversible mutations where possible, preserve shared containers,
  and restore prior state if the uninstall transaction fails.

#### H5. Windows uninstall is asynchronous

- **Convention:** lifecycle operations must not claim success based on a
  detached helper.
- **Evidence:** Windows uses a detached helper for removal
  (`internal/selfmanage/replace_windows.go:90-116`) and reports scheduled
  removal (`cmd/uninstall.go:63-67`).
- **Gap:** the command cannot confirm that removal completed.
- **Required correction:** coordinate removal through the stable launcher and
  a synchronous, verifiable uninstall protocol shared with the PowerShell
  script.

### I. README, workflows, tests, and release gates

#### I1. README does not end with the mandatory `## Uninstall` section

- **Convention:** the last README section must be `## Uninstall`, show remote
  Bash and PowerShell uninstall commands, state that default uninstall removes
  binaries/config/data/databases/agent resources/instructions/project config,
  and show dry-run and preservation examples.
- **Evidence:** there is no `## Uninstall`; the final heading is
  `## Development` (`README.md:179`). Lines `103-126` show only the Cobra
  command and `--keep-agent-resources`, with no remote scripts, destructive
  scope, confirmation, or preservation examples.
- **Gap:** the most destructive operation is neither correctly implemented nor
  prominently documented.
- **Required correction:** add this only after the shared uninstall contract is
  implemented and tested, and keep it as the final section.

#### I2. README and durable implementation documentation are release-stale

- **Convention:** version-bearing docs and schemas must be governed by the
  project version/release contract.
- **Evidence:** the current tag is `buda/v0.1.1`, while README exact-install and
  upgrade examples use `buda/v0.0.2` (`README.md:67-76,97-111`) and
  `docs/IMPLEMENTATION.md` still describes the 0.0.2 delivery
  (`docs/IMPLEMENTATION.md:4-22,47,84-106`). The docs descriptor repeats 0.0.2.
- **Gap:** public guidance does not describe the current release, and Mirror/XDocs
  did not prevent the drift.
- **Required correction:** decide which docs are historical versus current,
  label historical records explicitly, make current commands version-derived
  or release-neutral, and enforce parity in validation.

#### I3. CI and publication enforce superseded behavior

- **Convention:** validation must prove the new launcher, manifest, complete
  artifacts, raw version, selector, installation, upgrade, and uninstall
  contracts.
- **Evidence:** CI explicitly builds exactly eleven assets and asserts
  `buda v0.0.2-ci` (`.github/workflows/ci.yml:39-54`). Publication likewise
  requires exactly eleven files and prefixed version output
  (`.github/workflows/publish.yml:125-133,181-221`). Adding mandatory launchers,
  manifests, schemas, examples, or agent resources would make these gates fail
  or cause publication reconciliation to delete them. CI also does not run
  Mirror configuration/current checks, RunX checks/listing, strict XDocs
  metadata/tree/doctor checks, schema/example parity, launcher tests, installed
  manifest tests, or complete uninstall acceptance.
- **Gap:** automation protects the old release model and blocks the new one.
- **Required correction:** derive expected assets from `artifacts.json` and add
  native launcher/payload, installer, upgrade, uninstall, checksum, projection,
  rollback, and preservation acceptance on supported platforms.

#### I4. Tests assert behavior the convention forbids

- **Convention:** the required contract must be covered by automated tests.
- **Evidence:** `cmd/root_test.go:56-58` expects prefixed version output;
  `cmd/root_test.go:77-90` allows help depth `1`;
  `devops/build-binaries_test.go:32-39` expects exactly eleven artifacts;
  `cmd/selfmanage_test.go:122-140` treats same-version upgrade as a no-op; and
  `internal/selfmanage/catalog_test.go:134-167` retains scheduled replacement
  behavior. There are no tests for channels, manifests, launcher/pointers,
  `.guiho/.temp`, hidden self-test, locks, journals, peer-process validation,
  complete resource rollback, dual configs, evolution inheritance,
  confirmation, preservation, or uninstall scripts.
- **Gap:** a fully green test suite currently proves conformance to the obsolete
  design, not to Convention 0001.
- **Required correction:** replace incompatible assertions and build a
  traceable requirement-to-test matrix before changing release gates.

#### I5. Repository-local instructions still mandate the superseded release contract

- **Convention:** repository instructions, tooling, implementation, and release
  gates must agree on the convention-managed complete release unit.
- **Evidence:** `AGENTS.md:54-55` still mandates the standard exact
  eleven-artifact release. The convention requires additional launcher,
  manifest, agent-resource, schema, example, and operational artifacts, so the
  exact old count cannot remain authoritative. The repository instructions also
  contain Mirror guidance but no RunX catalog requirement.
- **Gap:** an implementation agent following current repository instructions
  faithfully would preserve a release model that Convention 0001 rejects.
- **Required correction:** after the new architecture and artifact inventory
  are approved, update the repository instructions and Go CLI skill contract in
  one governed decision so local instructions, convention, tooling, and CI are
  not contradictory.

## Compliant foundations to preserve

The following were reviewed and do obey the relevant convention requirement or
provide a sound base for remediation:

- Cobra `v1.10.2` is current, and Buda uses one fresh, testable Cobra tree
  (`cmd/root.go:172-271`; `cmd/application.go:5-20`).
- Dependencies, streams, time, process, filesystem, and network boundaries are
  broadly injected and testable; `main.go` remains thin.
- Public flags are lowercase long names with kebab-case multiword names. No
  nonstandard short aliases were found; only command help and root version use
  `-h` and `-v`.
- Cobra/pflag supplies both `--name value` and `--name=value` forms. No custom
  quote stripping was found. There are currently no list-valued public flags,
  so the `StringArray` rule is not applicable yet.
- Standard help is available throughout the command tree and does not perform
  repository work. Ordinary output and diagnostics use separate streams.
- JSON output is centralized as one document, and exit mapping is centralized.
- The required root shapes `init`, `agent`, `upgrade`, `upgrade check`,
  `upgrade list`, and `uninstall` exist. There is no public Cobra `install`
  command, as required.
- Existing project YAML decoding uses `go.yaml.in/yaml/v3`,
  `KnownFields(true)`, single-document enforcement, and semantic path
  validation (`internal/config/config.go:99-170`).
- Existing repository initialization validates containment, stages writes,
  protects immutable wiki identity, and is idempotent for its current narrower
  scope (`internal/repository/initialize.go:30-171`).
- Bounded instruction reconciliation preserves unrelated content and
  deduplicates its managed block (`internal/agent/instruction.go:242-275`).
- Skill projection replacement is internally staged and rollback-aware for its
  current limited destinations (`internal/agent/skill.go:53-82,136-269`).
- The CLI release catalog paginates GitHub releases, rejects drafts/malformed
  canonical tags, and sorts SemVer (`internal/selfmanage/catalog.go:129-190`).
- The eight payload targets and architecture tuning are correct and pure Go
  with `CGO_ENABLED=0` (`devops/build-binaries.go:34-43,104-114`). Foreign
  targets remain build-only unless executed on native runners.
- The current obsolete artifact set is sorted and checksummed deterministically
  (`devops/build-binaries.go:94-96,196-227`).
- The repository uses native Go formatting, tidy, test, vet, and build tooling.
  The audit's live Go tests passed; that is useful regression evidence even
  though it is not convention-compliance evidence.

## Required remediation sequence

The gaps are coupled. Fixing them as isolated flags or scripts would create
another partial lifecycle. The dependency-safe order is:

1. **Durable requirements and ownership:** confirm the main setup prompt and
   any agent definitions; define the complete artifact and state inventories;
   classify replaceable, persistent, disposable, shared, and project-owned
   paths.
2. **Version/config foundation:** update Go; make Mirror authoritative across
   all versioned surfaces; implement global/project configs, merge semantics,
   evolution policy, schemas, examples, pinned schema URLs, and embedded parity.
3. **Common `init`:** implement the interactive/noninteractive reconciliation,
   mandatory `AGENTS.md`, canonical artifact source, policy questions, final
   validation, and absolute-path summary.
4. **CLI surface:** switch to raw SemVer, correct developer-help behavior,
   rename both agent `update` commands to `upgrade`, and add the required agent
   skill evolution guidance and setup prompt.
5. **Release architecture:** add platform launchers, immutable payload layout,
   strict `current.json`, `artifacts.json`, installed manifest, complete agent
   resources, schemas/examples, checksums, and manifest-derived release gates.
6. **Installers:** rebuild both installers around complete paginated selection,
   exact raw SemVer/channels, canonical `.guiho` paths, candidate validation,
   hidden self-test, transactional activation, projection rollback,
   preservation, PATH idempotency, and final `init`.
7. **Upgrade:** implement recovery-first output, exclusive locking, transaction
   journal, stale recovery, peer-instance verification, synchronous pointer
   activation, complete resource replacement, full rollback, and post-upgrade
   `init`. Remove all scheduled/detached success behavior.
8. **Uninstall:** implement the two missing scripts and rewrite Cobra uninstall
   around one manifest-derived grouped plan, confirmation/`--yes`, config/data
   preservation, complete default removal, path-ownership validation, and
   rollback-safe execution.
9. **Documentation and tests:** update README's final uninstall section, current
   installation guidance, XDocs topology/descriptors, and create a
   requirement-to-test matrix. Replace every test/workflow assertion that
   protects the obsolete design.
10. **Release gate:** run format, tidy, tests, vet, strict XDocs checks, RunX
    checks, Mirror config/current/plan, all native lifecycle acceptance, foreign
    cross-builds, artifact/checksum verification, and native launcher/payload
    smoke before any release.

## Validation performed for this audit

| Check | Result |
|---|---|
| `mirror config check` | Passed; this proves the current YAML is valid, not that all version-bearing surfaces are governed. |
| `runx check --format json` | Failed as expected because neither repository nor user scope contains `runx.yaml`; the repository requirement remains unsatisfied. |
| `xdocs meta . --documents --strict --format json` | Passed, including this audit's frontmatter and descriptor registration. |
| `xdocs tree` | Command passed, but its output confirmed the finding: it showed `buda-package` and omitted the disconnected `buda-github` / workflow subtree. |
| `xdocs doctor . --warnings-as-errors --format json` | Passed with zero errors and zero warnings; current doctor does not diagnose the second-root topology. |
| `gofmt -l` over tracked Go files | Passed with no unformatted files. |
| `go mod tidy -diff` | Passed with no module-file drift. |
| `go test -count=1 ./...` | Passed across all packages using an isolated workspace-local Go build cache. |
| `go vet ./...` | Passed across all packages using the same isolated cache. |
| `git diff --check` | Passed for the audit document and descriptor update. |

The validation result does not change the verdict. It shows that the repository
is internally healthy against its current tests and metadata rules while those
tests and rules still encode an older CLI convention.

## Final determination

The answer is **no**: Buda is not compliant with GUIHO CLI Convention 0001 at
the audited commit. Its core Go/Cobra engineering is reusable, but compliance
requires a coordinated redesign of configuration, managed agent resources,
installation layout, release artifacts, upgrade, uninstall, documentation, and
automation. Passing the current suite must not be used as a compliance claim,
because the suite and workflows still codify several behaviors the convention
explicitly rejects.

## Planning follow-up

Implementation is decomposed in the
[Convention 0001 implementation plan](GUIHO_CLI_CONVENTION_0001_IMPLEMENTATION_PLAN.md).
Every finding in this audit maps to a work unit and exact acceptance evidence
in the [acceptance matrix](GUIHO_CLI_CONVENTION_0001_ACCEPTANCE_MATRIX.md), with
repository status tracked in [`TODO.md`](../TODO.md).
