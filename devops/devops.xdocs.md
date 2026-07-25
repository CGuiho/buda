---
subject: buda-devops
description: Pure-Go cross-build matrix, exact eleven-artifact release layout, and checksum-verifying installers.
parent: buda-package
children: []
files:
  build-binaries.go: Eight-target CGO-disabled binary build and eleven-artifact assembly.
  build-binaries_test.go: Target matrix, naming, checksum, and archive contract tests.
  install.ps1: PowerShell binary and embedded-skill installer.
  install.sh: POSIX shell binary and embedded-skill installer.
  installers_test.go: Installer pinning, checksum, and destination contract tests.
documents: {}
tags:
  - devops
  - distribution
keywords:
  - CGO_ENABLED
  - eleven artifacts
  - installers
flags: []
status: draft
---
