---
subject: buda-internal-maintenance
description: Failure-isolated first-success reconciliation split between bare global skills and explicitly selected wiki instructions.
parent: buda-internal
children: []
files:
  detach_unix.go: Unix detached-worker process attributes.
  detach_windows.go: Windows detached-worker process attributes.
  maintenance.go: Bare global-only and explicit-wiki reconciliation routing, scoped leases, detached scheduling, and local execution.
  maintenance_test.go: Bootstrap routing, eligibility, isolation, and lease behavior tests.
documents: {}
tags:
  - maintenance
  - background-worker
keywords:
  - reconciliation
  - failure isolation
  - local resources
  - global bootstrap
  - explicit wiki
flags: []
status: draft
---
