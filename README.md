#### &copy; 2026 [GUIHO](https://guiho.co) as represented by [Cristóvão GUIHO](https://guiho.co/cguiho) All Rights Reserved.

# GUIHO Buda

Buda is a repository-agnostic Go/Cobra CLI for maintaining one explicitly
selected AI-maintained wiki in Google's portable Open Knowledge Format. qmd is
the sole external indexing and retrieval engine; Buda does not guess a wiki,
federate repositories, publish knowledge, or implement a retrieval fallback.

## Install

Windows PowerShell:

```powershell
irm https://raw.githubusercontent.com/CGuiho/buda/main/devops/install.ps1 | iex
```

macOS or Linux:

```sh
curl -fsSL https://raw.githubusercontent.com/CGuiho/buda/main/devops/install.sh | sh
```

AI agent:

```text
Load the `guiho-p-buda-install` prompt and follow it in order to install the Buda CLI. Find it here: https://raw.githubusercontent.com/CGuiho/buda/main/prompts/guiho-p-buda-install.md
```

Verify the raw installed version:

```text
buda --version
```

Installation is global-only. It installs the stable launcher, immutable
payload, release resources, and global Buda skill; it does not select,
initialize, or modify a wiki and does not invoke qmd. Run `buda init` separately
only when one explicit wiki should be set up.

## Uninstall

Uninstallation removes all Buda-owned installation data by default, including
the launcher, payloads, configuration, persistent data, caches, manifests, and
agent resources. It removes the selected wiki's Buda configuration and managed
instruction block, but never canonical OKF knowledge, raw evidence, qmd-owned
state, shared `.guiho/` infrastructure, or another CLI's files.

Windows PowerShell:

```powershell
& ([scriptblock]::Create((Invoke-RestMethod 'https://raw.githubusercontent.com/CGuiho/buda/main/devops/uninstall.ps1'))) -Wiki C:\path\to\wiki -Yes
```

macOS or Linux:

```sh
curl -fsSL https://raw.githubusercontent.com/CGuiho/buda/main/devops/uninstall.sh | sh -s -- --wiki /path/to/wiki --yes
```

AI agent:

```text
Load the `guiho-p-buda-uninstall` prompt and follow it in order to uninstall the Buda CLI. Find it here: https://raw.githubusercontent.com/CGuiho/buda/main/prompts/guiho-p-buda-uninstall.md
```

Preview the exact `REMOVE` and `PRESERVE` plan without changing files:

```text
buda uninstall --wiki <path> --dry-run
```

The destructive default and the combined configuration/data preservation form
are:

```text
buda uninstall --wiki <path> --yes
buda uninstall --wiki <path> --preserve-config --preserve-data --yes
```

The remote scripts expose the equivalent `--dry-run`, `--preserve-config`, and
`--preserve-data` options on macOS/Linux and `-DryRun`, `-PreserveConfig`, and
`-PreserveData` on PowerShell.

## Migrate a 0.1.x direct-binary installation

A verified 0.1.x direct binary is migrated only through the new launcher
transaction: the installer activates and verifies the new launcher and
immutable payload first, then removes the old binary from its exact historical
path (`~/.local/bin/buda` on Unix, `%LOCALAPPDATA%\GUIHO\bin\buda.exe` on
Windows). Installation does not migrate or initialize a wiki. A later explicit
`buda init --wiki <path>` strictly validates and maps the preserved legacy
global `buda.yaml` into `buda.global.yaml` and carries its `wiki_id` into the
selected wiki. Canonical OKF knowledge and configuration are never deleted by
installation.

## Upgrade

```text
buda upgrade check
buda upgrade list
buda upgrade --channel stable --wiki <path>
buda upgrade --version 0.2.0 --wiki <path>
```

Upgrade is synchronous and verifies the complete manifest, checksums, payload,
launcher, agent resources, schemas, examples, and managed projections before
activating `current.json`. If it fails, the command prints a full reinstall
command pinned to the resolved selector. Read the effective
`agent.evolution` policy before an agent performs an upgrade or creates an
issue.

## Agent resources

The main skill is `guiho-s-buda`; the setup prompt is `guiho-p-buda`;
the lifecycle prompts are `guiho-p-buda-install` and
`guiho-p-buda-uninstall`; and the managed instruction is `guiho-i-buda`.
Inspect them without modifying files:

```text
buda agent skill list
buda agent skill show guiho-s-buda
buda agent prompt list
buda agent prompt show guiho-p-buda-install
buda agent prompt show guiho-p-buda-uninstall
buda agent instruction show --wiki <path>
```

Every agent-resource command uses `upgrade`, never the prohibited `update`
name. `buda init --wiki <path>` reconciles all supported global skill
destinations and the bounded `AGENTS.md` instruction block while preserving
unmanaged bytes.

## Development

Repeatable development, lifecycle, documentation, and release commands are
cataloged in `runx.yaml`. The required gates are:

```text
mirror config check
runx check --format json
xdocs meta . --documents --strict
xdocs tree
xdocs doctor .
gofmt -l main.go cmd devops internal prompts schemas skills
go mod tidy -diff
go test -count=1 ./...
go vet ./...
```

`devops/build-binaries.go` produces the eight pure-Go payloads, eight stable
launchers, typed resources, schemas, examples, `artifacts.json`, and
`checksums.txt`. Release completeness is derived from the manifest rather than
from a fixed asset count. CI additionally runs native POSIX and Windows
lifecycle jobs that install, repair, migrate a synthetic 0.1.1 layout, and
uninstall in disposable homes, plus interruption, rollback, and locked-file
acceptance in the Go suite.
