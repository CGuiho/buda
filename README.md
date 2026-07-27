---
name: buda-readme
purpose: Introduce Buda and document its repository-agnostic wiki workflow.
description: Public overview, verified installation and upgrade instructions, explicit-wiki contract, OKF knowledge model, qmd boundary, commands, and agent resources.
created: 2026-07-26
owner: buda-package
flags: []
tags:
  - readme
  - cli
  - knowledge
keywords:
  - Buda
  - Open Knowledge Format
  - qmd
  - LLM Wiki
  - install Buda
  - buda/v0.0.2
---

# GUIHO Buda

GUIHO Buda (B-U-D-A) is a public Go/Cobra CLI and agent-skill collection for
managing AI-maintained wikis. Buda helps agents save, navigate, read, validate,
and maintain any one explicitly selected AI-maintained wiki. It follows the
persistent knowledge operating pattern in
[Karpathy's LLM Wiki](https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f)
and stores portable knowledge according to Google's canonical
[Open Knowledge Format specification](https://github.com/GoogleCloudPlatform/knowledge-catalog/blob/main/okf/SPEC.md).
Google Cloud's
[official OKF overview](https://cloud.google.com/blog/products/data-analytics/how-the-open-knowledge-format-can-improve-data-sharing)
is useful explanatory context.

The name honors José António Ernesto, a former high-school philosophy professor
known as Buda. The note explains the name; it does not affect the technical
contract.

## Boundaries

Every repository command requires `--wiki <path>`. Buda never guesses a wiki,
stores a default wiki, searches several wikis, moves knowledge between
repositories, or publishes a wiki. Personal knowledge, research, team or
project knowledge, company knowledge, and application documentation are only
examples; Buda does not encode fixed use cases, visibility classes, or corpus
roles.

[qmd](https://github.com/tobi/qmd) is the external project-local indexing and
retrieval engine. Buda delegates lexical, semantic, and hybrid retrieval,
embeddings, ranking, reranking, document lookup, and index diagnostics to qmd.
Buda owns explicit repository selection, OKF-aware initialization, capture and
ingest orchestration, provenance, citations, health checks, stable output,
agent skills, instructions, and deterministic packaging.

## Install

Buda requires [qmd](https://github.com/tobi/qmd) as its external indexing and
retrieval engine. Install the tested qmd 2.5.3 release with Node.js 22 or later,
then confirm that `qmd` is on `PATH`:

```powershell
npm install --global @tobilu/qmd@2.5.3
qmd --version
```

The npm command installs qmd only. Buda itself is distributed exclusively as
checksum-verified binaries in GitHub Releases; it is not published to npm or
another package registry.

Install the latest Buda release on Linux or macOS:

```sh
curl -fsSL https://raw.githubusercontent.com/CGuiho/buda/main/devops/install.sh | sh
```

Install the latest release on Windows PowerShell:

```powershell
irm https://raw.githubusercontent.com/CGuiho/buda/main/devops/install.ps1 | iex
```

For a reproducible install, pass the exact canonical release tag. Buda never
adds or guesses a tag prefix:

```sh
curl -fsSL -H 'Accept: application/vnd.github.raw+json' \
  'https://api.github.com/repos/CGuiho/buda/contents/devops/install.sh?ref=buda%2Fv0.0.2' |
  sh -s -- buda/v0.0.2
```

```powershell
$installerSource = Invoke-RestMethod `
  -Headers @{ Accept = 'application/vnd.github.raw+json' } `
  -Uri 'https://api.github.com/repos/CGuiho/buda/contents/devops/install.ps1?ref=buda%2Fv0.0.2'
$installBuda = [scriptblock]::Create($installerSource)
& $installBuda -Version 'buda/v0.0.2'
```

The installers select the native release binary, verify both the binary and
embedded-skill archive against `checksums.txt`, replace an existing binary with
rollback protection, install Buda's canonical skill in both global agent
locations, and verify the installed version. They do not install qmd or write
wiki instructions without an explicit wiki. The default binary destinations
are `~/.local/bin/buda` on Linux and macOS and
`%LOCALAPPDATA%\GUIHO\bin\buda.exe` on Windows.

Verify the installation and initialize one explicitly selected wiki:

```text
buda --version
buda init --wiki <path> --wiki-id <id>
buda doctor --wiki <path>
```

### Upgrade from 0.0.1

Run the exact `buda/v0.0.2` installer command above. The installer stages and
verifies 0.0.2 before replacing the existing 0.0.1 executable, restores the
previous executable if installation fails, and refreshes both embedded global
skill copies. Canonical wiki content and explicit-wiki instructions are not
rewritten by the installer.

Once Buda is installed, its native self-management routes use only canonical
`buda/v<semver>` GitHub Releases and checksum-verified platform assets:

```text
buda upgrade check
buda upgrade list
buda upgrade --dry-run
buda upgrade
buda upgrade --version 0.0.2
buda upgrade rollback
buda uninstall --dry-run
buda uninstall
```

`buda upgrade` reports release URLs, download progress, checksum validation,
the installation path, agent-resource reconciliation, and final version
verification. On Unix it replaces the executable on the same filesystem and
automatically rolls back on verification or resource-reconciliation failure.
On Windows a detached helper waits for the running process to exit before
performing the same verified replacement. `upgrade`, `uninstall`, and their
hidden helpers never select a wiki implicitly, run qmd, or trigger Buda's
background resource reconciler. Pass an explicit `--wiki <path>` only when an
upgrade or uninstall should also reconcile that wiki's bounded instruction
block. `buda uninstall --keep-agent-resources` removes only the executable.

## Commands

```text
buda init --wiki <path> --wiki-id <id>
buda capture --wiki <path> ...
buda ingest --wiki <path> --source <path-or-url> ...
buda lint --wiki <path>
buda index --wiki <path> [--embed]
buda query --wiki <path> --text <query> [--mode lexical|semantic|hybrid]
buda get --wiki <path> <concept-path-or-result-id>
buda status --wiki <path>
buda pack --wiki <path> --output <path>
buda doctor --wiki <path>
buda agent ...
buda upgrade [check|list|rollback]
buda uninstall
```

Every command supports `--help`, `--help-tree`, positive
`--help-tree-depth`, and `--help-docs`. Root also supports `--version` and
`-v`. Repository commands provide stable text and one-document JSON output.

## Agent resource bootstrap

A successful bare `buda` invocation may schedule a hidden, non-blocking
reconciler for the embedded Buda skill in both global agent-skill locations. It
has no wiki context and never writes repository instructions.

A successful `buda --wiki <path>` invocation or successful repository command
may schedule the same global reconciliation and, because it carries an
explicitly resolved wiki, reconcile Buda's bounded instruction block in only
that selected wiki. The block identifies the selected wiki and bundle, declares
Buda the required tool for maintaining and retrieving it, requires every
repository operation to pass `--wiki`, and requires agents to evaluate evidence
and cite concept paths plus source IDs or resources. Agents maintain the wiki
through Buda rather than bypassing it or invoking qmd directly. The block also
requires weak evidence to be surfaced and says that repository instructions are
behavior rather than access control.

Bootstrap performs no qmd or network operation, does not discover repositories,
and never reads or writes another wiki. Worker startup or reconciliation failure
cannot change foreground output or exit status. Explicit `buda agent ...`, help,
version, and hidden-worker routes do not schedule reconciliation.

## Wiki layout

An initialized wiki contains strict `buda.yaml`, canonical OKF files under
`knowledge/`, disposable project-local qmd state under `.qmd/`, and disposable
Buda staging/report state under `.buda/`. Canonical Markdown and source evidence
remain readable without Buda or qmd.

## Development

Buda requires a supported Go toolchain and builds with `CGO_ENABLED=0`.

```powershell
go mod tidy
go test ./...
go vet ./...
go run . --help-tree
```

qmd is a separate runtime dependency. Buda does not vendor or silently install
it. CI validates the Go CLI and exact eleven-artifact build contract. Canonical
`buda/v*` tags drive GitHub Release publication of those eleven artifacts only;
no Buda package is published to npm or another package manager.
