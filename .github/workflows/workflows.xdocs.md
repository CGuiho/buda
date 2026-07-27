---
subject: buda-github-workflows
description: Go CI, native smoke coverage, canonical-tag GitHub Release publication, exact-asset reconciliation, and public installer acceptance.
parent: buda-github
children: []
files:
  ci.yml: Go formatting, module, test, vet, exact eleven-artifact build, checksum, installer syntax, hosted native-platform smoke, and real qmd 2.5.3 project-local lexical integration validation.
  publish.yml: Canonical buda/v* GitHub Release publication with exact-version notes, exactly eleven assets, checksum verification, pre-publication Linux and Windows upgrade-installer gates, real qmd lexical retrieval, and post-publication installer acceptance.
documents: {}
tags:
  - github-actions
  - ci
  - release
keywords:
  - buda/vX.Y.Z
  - eleven artifacts
  - public installers
  - GitHub Release
  - no package publication
flags: []
status: stable
---

CI builds all eight pure-Go release targets and executes compatible artifacts
on hosted Linux, macOS, and Windows AMD64 and ARM64 runners. ARMv6 and ARMv7
remain cross-build-only because GitHub does not provide matching hosted runners.
It also installs the separate tested qmd 2.5.3 runtime and exercises Buda init,
capture, project-local indexing, lexical query, get, status, and doctor on
Linux, plus the npm `qmd.cmd` launcher path on Windows, without downloading
models or adding a Buda package-manager publication.

Publication runs only for canonical `buda/v*` tags. It reconciles one GitHub
Release to the exact eleven checked artifacts and verifies the tag-pinned Bash
and PowerShell installers before publication, then repeats public installer
acceptance after publication, without publishing to a package manager or
registry.
