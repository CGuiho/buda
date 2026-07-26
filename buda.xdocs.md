---
subject: buda-package
description: Public Go/Cobra Buda CLI, embedded agent resources, OKF wiki services, qmd adapter, documentation, and pure-Go distribution tooling.
parent: null
children:
  - buda-cmd
  - buda-internal
  - buda-devops
  - buda-skills
  - buda-prompts
  - buda-docs
files:
  .gitignore: Git ignore rules for local executables, test output, and release artifacts.
  go.mod: Go module definition and pinned Cobra and YAML dependencies.
  go.sum: Reproducible Go dependency checksums.
  main.go: Thin executable entrypoint with linker-provided build metadata and centralized exit handling.
  xdocs.yaml: XDocs YAML configuration using automatic documentation updates and generated-directory exclusions.
documents:
  AGENTS.md: Repository engineering, product-boundary, documentation, validation, and release instructions.
  CHANGELOG.md: Chronological Buda release scope, compatibility boundaries, and exact distribution artifact contract.
  README.md: Public overview of Buda, its explicit-wiki model, OKF and qmd boundaries, agent bootstrap, commands, and development workflow.
tags:
  - go
  - cli
  - knowledge
  - agent-skills
keywords:
  - Buda
  - Cobra
  - Open Knowledge Format
  - qmd
  - LLM Wiki
flags: []
status: draft
---

Buda operates on exactly one explicitly selected wiki repository per command.
Canonical knowledge remains portable OKF Markdown and evidence; qmd and Buda
runtime state are disposable derived material.
