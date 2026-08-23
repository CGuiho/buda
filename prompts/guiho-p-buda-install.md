---
name: guiho-p-buda-install
purpose: Install the Buda CLI without initializing or selecting a wiki.
description: Use when an agent must install and verify Buda while keeping wiki initialization as a separate explicit task.
created: 2026-08-23
version: "0.2.0"
owner: buda-prompts
flags: []
tags:
  - installation
  - lifecycle
keywords:
  - Buda
  - install
  - checksums
  - stable launcher
metadata:
  version: "0.2.0"
---

#### &copy; 2026 [GUIHO](https://guiho.co) as represented by [Cristóvão GUIHO](https://guiho.co/cguiho) All Rights Reserved.

# Install Buda

Install the GUIHO Buda CLI by running exactly one command chosen for the host
operating system:

- Windows PowerShell:

  ```powershell
  irm https://raw.githubusercontent.com/CGuiho/buda/main/devops/install.ps1 | iex
  ```

- macOS or Linux:

  ```sh
  curl -fsSL https://raw.githubusercontent.com/CGuiho/buda/main/devops/install.sh | sh
  ```

To install an exact release or channel, use the same remote scripts with one
full-name selector. The selectors are mutually exclusive:

```powershell
& ([scriptblock]::Create((Invoke-RestMethod 'https://raw.githubusercontent.com/CGuiho/buda/main/devops/install.ps1'))) -Version 'X.Y.Z'
& ([scriptblock]::Create((Invoke-RestMethod 'https://raw.githubusercontent.com/CGuiho/buda/main/devops/install.ps1'))) -Channel '<name>'
```

```sh
curl -fsSL https://raw.githubusercontent.com/CGuiho/buda/main/devops/install.sh | sh -s -- --version 'X.Y.Z'
curl -fsSL https://raw.githubusercontent.com/CGuiho/buda/main/devops/install.sh | sh -s -- --channel '<name>'
```

Never install Buda through npm, Bun, pip, another package manager, or a direct
package registry. No package-manager distribution is supported.

When the installer finishes, verify the raw version:

```text
buda --version
```

Stop after version verification. Do not run `buda init`, select a wiki, create
wiki files, invoke qmd, or apply project instructions as part of this
installation task. Wiki initialization is a separate operation that requires
one explicitly selected `--wiki <path>`.

If installation fails, report the failed command and diagnostic. Ask before
creating an issue at <https://github.com/CGuiho/buda/issues/new>, and provide
the resulting issue URL only after creation succeeds.

For a manual fallback, resolve the matching GitHub Release and host target,
download `artifacts.json`, `checksums.txt`, the platform payload and launcher,
and every artifact declared by the manifest. Verify every SHA-256 digest before
execution. Run the staged payload with `--version` and `__self-test`, install it
under `$HOME/.guiho/buda/versions/<version>/`, install the stable launcher at
`$HOME/.guiho/bin/buda` (`buda.exe` on Windows), write the manifest-owned
activation files atomically, install the bundled global skill, ensure
`$HOME/.guiho/bin` is on `PATH`, and verify again with `buda --version`. Do not
initialize a wiki during the manual fallback.
