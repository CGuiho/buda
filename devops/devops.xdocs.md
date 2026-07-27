---
subject: buda-devops
description: Pure-Go cross-build matrix, exact eleven-artifact release layout, and checksum-verifying canonical-tag installers with transactional upgrade support.
parent: buda-package
children: []
files:
  build-binaries.go: Eight-target CGO-disabled binary build and eleven-artifact assembly.
  build-binaries_test.go: Target matrix, naming, checksum, and archive contract tests.
  install.ps1: PowerShell latest or exact canonical-tag binary installer with checksum verification, rollback, execution-policy-safe qmd.cmd compatibility checks, and embedded global-skill refresh.
  install.sh: POSIX latest or exact canonical-tag binary installer with checksum verification, rollback, qmd compatibility checks, and embedded global-skill refresh.
  installers_test.go: Installer canonical-tag extraction, qmd pinning, checksum, rollback, local-asset, and destination contract tests.
  workflows_test.go: GitHub Actions syntax, canonical tag, exact asset, real qmd integration, and no package or production publication contract tests.
documents: {}
tags:
  - devops
  - distribution
keywords:
  - CGO_ENABLED
  - eleven artifacts
  - installers
  - buda/v0.0.2
  - transactional upgrade
flags: []
status: draft
---
