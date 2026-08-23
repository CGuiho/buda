---
name: guiho-p-buda
purpose: Initialize one explicitly selected Buda wiki after the CLI is installed.
description: Use when an agent must set up or reconcile one selected Buda wiki and its persistent agent resources.
created: 2026-08-16
version: "0.2.0"
owner: buda-prompts
metadata:
  version: "0.2.0"
flags: []
tags:
  - setup
  - lifecycle
keywords:
  - Buda
  - init
  - explicit wiki
  - qmd
---

#### &copy; 2026 [GUIHO](https://guiho.co) as represented by [Cristóvão GUIHO](https://guiho.co/cguiho) All Rights Reserved.

# Set up Buda

Buda is a repository-agnostic Go CLI for maintaining one explicitly selected
AI-maintained wiki in Google's portable Open Knowledge Format. It delegates
indexing and retrieval to qmd and never guesses which repository to change.

Verify that Buda is already installed before selecting a wiki:

```text
buda --version
```

If Buda is not installed, stop and follow the separate
`guiho-p-buda-install` prompt. Installation must not select or initialize a
wiki.

Initialize one selected wiki only by passing its path explicitly:

```text
buda init --wiki <path>
```

When a newer release is available, read the effective `agent.evolution`
policy and use `buda upgrade check` followed by `buda upgrade` only when that
policy and the user's authority allow it. After an upgrade, verify the raw
version and rerun `buda init --wiki <path>`.

Use the separate `guiho-p-buda-uninstall` prompt for removal. Buda never
removes canonical OKF knowledge or raw evidence.
