---
subject: buda-package
description: Public Go/Cobra Buda CLI, embedded agent resources, OKF wiki services, qmd adapter, documentation, and GitHub Release-only pure-Go distribution tooling.
parent: null
children:
  - buda-cmd
  - buda-internal
  - buda-devops
  - buda-skills
  - buda-prompts
  - buda-docs
  - buda-schemas
  - buda-examples
files:
  .gitignore: Git ignore rules for local executables, test output, and release artifacts.
  go.mod: Go module definition and pinned Cobra and YAML dependencies.
  go.sum: Reproducible Go dependency checksums.
  main.go: Thin executable entrypoint with linker-provided build metadata and centralized exit handling.
  mirror.yaml: Mirror semantic-version, canonical-tag, changelog, commit, clean-worktree, and push configuration.
  xdocs.yaml: XDocs YAML configuration using automatic documentation updates and generated-directory exclusions.
  runx.yaml: Complete RunX catalog for repeatable development, lifecycle, tooling, and release commands.
documents:
  AGENTS.md: Repository engineering, product-boundary, documentation, validation, and release instructions.
  CHANGELOG.md: Chronological Buda release scope, compatibility boundaries, canonical tags, GitHub Release-only policy, and manifest-derived distribution contract.
  README.md: Public overview of Buda, verified latest and exact installation, installer and native self-upgrade/rollback/uninstall operations, explicit-wiki model, OKF and qmd boundaries, commands, and development workflow.
  TODO.md: Convention 0001 implementation-unit, acceptance-gate, human-confirmation, and release-authorization tracking.
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
  - GitHub Releases
  - buda/v0.1.1
flags: []
status: draft
---

Buda operates on exactly one explicitly selected wiki repository per command.
Canonical knowledge remains portable OKF Markdown and evidence; qmd and Buda
runtime state are disposable derived material.
