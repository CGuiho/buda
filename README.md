#### &copy; 2026 [GUIHO](https://guiho.co) as represented by [Cristóvão GUIHO](https://guiho.co/cguiho) All Rights Reserved.

# GUIHO Buda

Buda is a repository-agnostic Go/Cobra CLI for maintaining one explicitly
selected AI-maintained wiki in Google's portable Open Knowledge Format. qmd is
the sole external indexing and retrieval engine; Buda does not guess a wiki,
federate repositories, publish knowledge, or implement a retrieval fallback.

## Install

Install qmd separately according to its upstream documentation. Buda installers
require an explicit wiki path because no lifecycle action may select a project
implicitly.

For a new wiki directory, also pass the immutable identifier with
`--wiki-id <id>` (or `-WikiId <id>` in PowerShell); an existing valid
`buda.yaml` supplies it automatically.

Linux or macOS (latest stable):

```sh
curl -fsSL https://raw.githubusercontent.com/CGuiho/buda/main/devops/install.sh | sh -s -- --wiki /path/to/wiki
```

Windows PowerShell (latest stable):

```powershell
& ([scriptblock]::Create((Invoke-RestMethod 'https://raw.githubusercontent.com/CGuiho/buda/main/devops/install.ps1'))) -Wiki C:\path\to\wiki
```

Exact version or channel selectors are full-name options and mutually
exclusive:

```sh
curl -fsSL https://raw.githubusercontent.com/CGuiho/buda/main/devops/install.sh | sh -s -- --version 0.2.0 --wiki /path/to/wiki
curl -fsSL https://raw.githubusercontent.com/CGuiho/buda/main/devops/install.sh | sh -s -- --channel canary --wiki /path/to/wiki
```

```powershell
& ([scriptblock]::Create((Invoke-RestMethod 'https://raw.githubusercontent.com/CGuiho/buda/main/devops/install.ps1'))) -Version 0.2.0 -Wiki C:\path\to\wiki
& ([scriptblock]::Create((Invoke-RestMethod 'https://raw.githubusercontent.com/CGuiho/buda/main/devops/install.ps1'))) -Channel canary -Wiki C:\path\to\wiki
```

Verify the raw release version and initialize the selected wiki:

```text
buda --version
buda init --wiki <path>
```

The stable launcher is installed at `$HOME/.guiho/bin/buda` (or `buda.exe`),
and immutable payloads and manifest-owned resources live under
`$HOME/.guiho/buda/versions/<version>/`. Foreign ARMv6/ARMv7 targets are
cross-build-only unless run on their native platform.

### Migrating a 0.1.x direct-binary installation

A verified 0.1.x direct binary is migrated only through the new launcher
transaction: the installer activates and verifies the new launcher and
immutable payload first, then removes the old binary from its exact historical
path (`~/.local/bin/buda` on Unix, `%LOCALAPPDATA%\GUIHO\bin\buda.exe` on
Windows). The legacy global `buda.yaml` is strictly validated, mapped into
`buda.global.yaml`, and left in place; its `wiki_id` is carried into the
explicitly selected wiki so migration works without a terminal and keeps the
wiki's identity. Canonical OKF knowledge and configuration are never deleted
by installation.

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

The main skill is `guiho-s-0002-buda`; the setup prompt is `guiho-p-buda`; the
managed instruction is `guiho-i-buda`. Inspect them without modifying files:

```text
buda agent skill list
buda agent skill show guiho-s-0002-buda
buda agent prompt list
buda agent prompt show guiho-p-buda
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
gofmt -l .
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

## Uninstall

Uninstallation removes all Buda-owned installation data by default, including
immutable versions, the stable launcher, caches, manifests, configuration, and
the bounded instruction block for the explicitly selected wiki. It never
removes canonical OKF knowledge, raw evidence, qmd-owned state, shared
`.guiho/` infrastructure, or another CLI's files.

Preview first, then confirm explicitly in unattended environments:

```sh
curl -fsSL https://raw.githubusercontent.com/CGuiho/buda/main/devops/uninstall.sh | sh -s -- --wiki /path/to/wiki --dry-run
curl -fsSL https://raw.githubusercontent.com/CGuiho/buda/main/devops/uninstall.sh | sh -s -- --wiki /path/to/wiki --yes
```

```powershell
& ([scriptblock]::Create((Invoke-RestMethod 'https://raw.githubusercontent.com/CGuiho/buda/main/devops/uninstall.ps1'))) -Wiki C:\path\to\wiki -DryRun
& ([scriptblock]::Create((Invoke-RestMethod 'https://raw.githubusercontent.com/CGuiho/buda/main/devops/uninstall.ps1'))) -Wiki C:\path\to\wiki -Yes
```

Use `--preserve-config`/`-PreserveConfig` to retain global and selected-project
configuration, and `--preserve-data`/`-PreserveData` to retain persistent Buda
data and databases. `--dry-run`/`-DryRun` never mutates the filesystem.
