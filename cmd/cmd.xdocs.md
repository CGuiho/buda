---
subject: buda-cmd
description: Cobra command construction, application assembly, stable output, exit mapping, and command-level integration tests for Buda.
parent: buda-package
children: []
files:
  agent.go: Agent skill, instruction, and prompt command tree.
  application.go: Assembly of the public repository command set.
  capture.go: Explicit-wiki capture command orchestration.
  doctor.go: Read-only repository and qmd diagnostic command.
  domain_commands_test.go: Repository-command behavior and output tests.
  get.go: qmd-backed concept and raw-file retrieval command.
  index.go: qmd project initialization and indexing command.
  ingest.go: Source registration and qmd-backed ingest work-item command.
  init.go: OKF-aware wiki initialization command.
  lint.go: Base OKF conformance and Buda health command.
  pack.go: Deterministic portable-bundle packaging command.
  pack_test.go: Packaging command integration tests.
  query.go: qmd-backed query and normalized evidence command.
  query_test.go: Query command normalization and failure tests.
  root.go: Fresh Cobra root, persistent flags, help routes, dependency injection, bare global bootstrap, explicit-wiki reconciliation routing, and exit handling.
  root_test.go: Root command, help, JSON-error, and bootstrap-scheduling tests.
  status.go: Repository, health, and qmd readiness status command.
documents: {}
tags:
  - cobra
  - cli
  - commands
keywords:
  - command tree
  - stable output
  - explicit wiki
flags: []
status: draft
---

This module owns Buda's user-facing Cobra surface and composes domain services
without owning canonical wiki persistence or retrieval algorithms.
