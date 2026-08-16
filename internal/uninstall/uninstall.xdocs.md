---
subject: buda-internal-uninstall
description: Manifest-driven ownership-safe uninstall planning and application without recursive removal of unmanaged CLI-home content.
parent: buda-internal
children: []
files:
  plan.go: Manifest-proven removal and preservation plan generation with global/local skill and bounded instruction cleanup.
  plan_test.go: Ownership, preservation, and corrupt-manifest safety tests.
documents: {}
tags: [uninstall, filesystem]
keywords: [preserve-config, preserve-data, dry-run, ownership]
flags: []
status: draft
---

Canonical OKF content and raw evidence are never uninstall targets. The CLI-home
container is never removed recursively; only manifest-proven or explicitly
classified Buda-owned leaves are planned.
