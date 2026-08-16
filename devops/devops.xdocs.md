---
subject: buda-devops
description: Pure-Go cross-build matrix, manifest-derived complete release layout, and checksum-verifying lifecycle scripts.
parent: buda-package
children: []
files:
  build-binaries.go: Eight-target CGO-disabled payload and launcher build with manifest-derived assembly.
  build-binaries_test.go: Target matrix, naming, checksum, and archive contract tests.
  install.ps1: PowerShell exact/channel complete-release installer with staging, checksums, immutable payloads, launcher activation, and PATH repair.
  install.sh: POSIX exact/channel complete-release installer with staging, checksums, immutable payloads, launcher activation, and PATH repair.
  uninstall.ps1: PowerShell uninstaller with ownership preservation and explicit confirmation options.
  uninstall.sh: POSIX uninstaller with ownership preservation and explicit confirmation options.
  installers_test.go: Installer and uninstaller selector, checksum, staging, ownership, and safety-contract tests.
  workflows_test.go: GitHub Actions syntax, canonical tag, manifest-derived asset, and no package or production publication contract tests.
documents: {}
tags:
  - devops
  - distribution
keywords:
  - CGO_ENABLED
  - artifacts.json
  - installers
  - buda/v0.2.0
  - transactional upgrade
flags: []
status: draft
---
