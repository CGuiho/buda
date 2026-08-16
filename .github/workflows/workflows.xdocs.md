---
subject: buda-github-workflows
description: Go quality, native launcher and payload smoke, complete manifest-derived GitHub Release publication, and public lifecycle acceptance.
parent: buda-github
children: []
files:
  ci.yml: Go formatting, module, test, vet, RunX, Mirror, strict XDocs, manifest build, checksums, installer syntax, hosted native launcher/payload smoke, POSIX and Windows lifecycle acceptance, legacy 0.1.1 migration, and foreign build-only targets.
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

CI builds the manifest-derived complete release unit and validates checksums.
Hosted Linux, macOS, and Windows AMD64 and ARM64 runners execute native
payload smokes and launcher smokes through a disposable pointer fixture.
ARMv6 and ARMv7 remain cross-build-only because GitHub does not provide
matching hosted runners. RunX, Mirror, strict XDocs, Go quality, lifecycle
transaction and interruption acceptance, and both installer-language syntax
checks are required.

Native lifecycle jobs on Linux and Windows install the release from a local
asset directory into a disposable home with a qmd process stub, verify the
stable launcher, prove same-version reinstall repair, migrate a synthetic
0.1.1 direct-binary layout through the launcher transaction (carrying the
legacy wiki_id into the newly selected wiki), and run a synchronous default
uninstall while asserting that canonical OKF content and shared GUIHO
sentinels survive.

Publication runs for canonical `buda/v*` tags or a manual recovery dispatch.
It reconciles one GitHub Release to every manifest-declared asset, verifies
the public asset names against the generated release directory, and never
publishes to a package manager or registry.
