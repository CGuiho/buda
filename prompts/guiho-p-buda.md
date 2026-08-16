---
name: guiho-p-buda
purpose: Install and initialize the Buda repository-agnostic OKF wiki CLI.
description: Guide an agent through installing, verifying, initializing, upgrading, and safely uninstalling Buda.
created: 2026-08-16
version: "0.2.0"
owner: buda-prompts
flags: []
tags:
  - setup
  - lifecycle
keywords:
  - Buda
  - install
  - init
  - upgrade
  - uninstall
---

# Set up Buda

Buda is a repository-agnostic Go CLI for maintaining one explicitly selected
AI-maintained wiki in Google's portable Open Knowledge Format. It delegates
indexing and retrieval to qmd and never guesses which repository to change.

Install the latest stable release for the current platform with the canonical
remote installer, always providing the selected wiki path:

```sh
curl -fsSL https://raw.githubusercontent.com/CGuiho/buda/main/devops/install.sh | sh -s -- --wiki <path>
```

```powershell
& ([scriptblock]::Create((Invoke-RestMethod 'https://raw.githubusercontent.com/CGuiho/buda/main/devops/install.ps1'))) -Wiki <path>
```

Then verify the installed CLI with:

```text
buda --version
```

Initialize one selected wiki only by passing its path explicitly:

```text
buda init --wiki <path>
```

When a newer release is available, read the effective `agent.evolution`
policy and use `buda upgrade check` followed by `buda upgrade` only when that
policy and the user's authority allow it. After an upgrade, verify the raw
version and rerun `buda init --wiki <path>`.

Use `buda uninstall --dry-run --wiki <path>` to inspect ownership before a
removal. Buda never removes canonical OKF knowledge or raw evidence.
