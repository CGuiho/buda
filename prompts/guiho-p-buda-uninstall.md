---
name: guiho-p-buda-uninstall
purpose: Uninstall Buda through its ownership-safe lifecycle interface.
description: Use when an agent must preview and perform a Buda uninstall while preserving canonical wiki knowledge and honoring requested preservation options.
created: 2026-08-23
version: "0.2.0"
owner: buda-prompts
flags: []
tags:
  - uninstallation
  - lifecycle
keywords:
  - Buda
  - uninstall
  - preserve-config
  - preserve-data
metadata:
  version: "0.2.0"
---

#### &copy; 2026 [GUIHO](https://guiho.co) as represented by [Cristóvão GUIHO](https://guiho.co/cguiho) All Rights Reserved.

# Uninstall Buda

Confirm the one selected wiki path whose Buda project configuration and managed
instruction block should be included. Never infer a wiki from the working
directory or search multiple repositories.

Explain that default uninstallation removes all Buda-owned installation data,
configuration, state, agent resources, and the selected wiki's Buda
configuration. It never removes canonical OKF knowledge, raw evidence,
qmd-owned state, shared `.guiho` directories, or another CLI's files.

Run exactly one platform uninstaller first as a dry run:

```powershell
& ([scriptblock]::Create((Invoke-RestMethod 'https://raw.githubusercontent.com/CGuiho/buda/main/devops/uninstall.ps1'))) -Wiki <path> -DryRun
```

```sh
curl -fsSL https://raw.githubusercontent.com/CGuiho/buda/main/devops/uninstall.sh | sh -s -- --wiki <path> --dry-run
```

Show the complete `REMOVE` and `PRESERVE` plan. Ask for confirmation before the
destructive invocation. When approved, rerun the same remote script with
`-Yes` on PowerShell or `--yes` on macOS/Linux. If requested, combine
`-PreserveConfig -PreserveData` or `--preserve-config --preserve-data` to keep
both configuration and persistent data. Do not treat agent resources, caches,
launchers, or versioned payloads as preservable data.

Report the selected wiki, preservation choices, command result, and remaining
preserved paths. If uninstallation fails, preserve the reported diagnostics and
do not bypass checksum, ownership, confirmation, or containment safeguards.
