---
subject: buda-devops
description: Pure-Go cross-build matrix, manifest-derived complete release layout, and checksum-verifying lifecycle scripts.
parent: buda-package
children: []
files:
  build-binaries.go: Eight-target CGO-disabled payload and launcher build with manifest-derived assembly.
  build-binaries_test.go: Target matrix, naming, checksum, and archive contract tests.
  install.ps1: Global-only PowerShell exact/channel complete-release installer with Windows-PowerShell-5.1-safe release discovery, staging, checksums, immutable activation, PATH repair, and global skill reconciliation without wiki initialization.
  install.sh: Global-only POSIX exact/channel complete-release installer with staging, checksums, immutable activation, PATH repair, and global skill reconciliation without wiki initialization.
  uninstall.ps1: PowerShell uninstaller with ownership preservation and explicit confirmation options.
  uninstall.sh: POSIX uninstaller with ownership preservation and explicit confirmation options.
  installers_test.go: Installer/uninstaller selector, install-only, README lifecycle-section, checksum, staging, ownership, safety-contract, and PowerShell 5.1 release-discovery regression tests.
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
  - Invoke-RestMethod
flags: []
status: draft
---

#### &copy; 2026 [GUIHO](https://guiho.co) as represented by [Cristóvão GUIHO](https://guiho.co/cguiho) All Rights Reserved.
