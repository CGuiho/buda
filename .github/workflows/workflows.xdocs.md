---
subject: buda-github-workflows
description: Go quality, native smoke coverage, complete manifest-derived GitHub Release publication, and public lifecycle acceptance.
parent: buda-github
children: []
files:
  ci.yml: Go formatting, module, test, vet, RunX, Mirror, strict XDocs, manifest build, checksums, installer syntax, and hosted native-platform smoke.
  publish.yml: Canonical buda/v* GitHub Release publication with manifest-derived assets and public checksum reconciliation.
documents: {}
tags:
  - github-actions
  - ci
  - release
keywords:
  - buda/vX.Y.Z
  - artifacts.json
  - public installers
  - GitHub Release
  - no package publication
flags: []
status: stable
---

CI builds the manifest-derived complete release unit, validates checksums, and
executes compatible native payload smoke checks on hosted Linux, macOS, and
Windows AMD64 and ARM64 runners. ARMv6 and ARMv7 remain cross-build-only
because GitHub does not provide matching hosted runners. RunX, Mirror, strict
XDocs, Go quality, and both installer-language syntax checks are required.

Publication runs for canonical `buda/v*` tags or a manual recovery dispatch.
It reconciles one GitHub Release to every manifest-declared artifact, verifies
the public asset names against the generated release directory, and never
publishes to a package manager or registry.
