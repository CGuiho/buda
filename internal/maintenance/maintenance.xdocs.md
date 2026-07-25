---
subject: buda-internal-maintenance
description: Failure-isolated first-success agent-resource reconciliation and platform-specific detached process setup.
parent: buda-internal
children: []
files:
  detach_unix.go: Unix detached-worker process attributes.
  detach_windows.go: Windows detached-worker process attributes.
  maintenance.go: Route eligibility, wiki-scoped lease, detached scheduling, and local reconciliation.
  maintenance_test.go: Eligibility, isolation, and lease behavior tests.
documents: {}
tags:
  - maintenance
  - background-worker
keywords:
  - reconciliation
  - failure isolation
  - local resources
flags: []
status: draft
---
