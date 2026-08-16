---
subject: buda-internal-selfmanage
description: Canonical GitHub Release discovery plus checksum-verified native upgrade and platform-safe executable removal.
parent: buda-internal
children: []
files:
  catalog.go: Bounded canonical buda/v semantic-release catalog parsing, filtering, and ordering.
  catalog_test.go: Catalog, semantic version, target selection, checksum, progress, backup rotation, rollback, and injected replacement tests.
  replace_unix.go: Same-filesystem Unix executable replacement, verification rollback, explicit rollback, and direct removal.
  replace_windows.go: Fail-closed Windows direct-replacement boundaries; canonical lifecycle operations use the stable launcher and immutable payload transaction.
  upgrade.go: Release target selection, bounded progress-reporting downloads, SHA-256 validation, safe backup rotation, candidate staging, rollback primitives, and executable verification.
documents: {}
tags:
  - self-management
  - upgrade
  - github-releases
keywords:
  - checksum
  - atomic replacement
  - uninstall
flags: []
status: draft
---

This module keeps Buda self-management separate from wiki operations. It only
accepts canonical `buda/v<semver>` releases and never performs qmd, wiki
discovery, cross-repository, or package-manager publication work.
