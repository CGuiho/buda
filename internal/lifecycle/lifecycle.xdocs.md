---
subject: buda-internal-lifecycle
description: Crash-safe locks, journals, atomic files, process liveness, and containment primitives.
parent: buda-internal
children: []
files:
  transaction.go: Token-owned lock, transaction journal, atomic write, and safe removal primitives.
  instances.go: Tokenized current-user payload registry and verified process termination boundary.
  instances_test.go: Registry ownership, cleanup, and invoking-process protection tests.
  process_windows.go: Windows process liveness adapter.
  process_unix.go: Unix process adapter boundary.
documents: {}
tags: [lifecycle, recovery]
keywords: [upgrade.lock, transaction.json, rollback]
flags: []
status: draft
---

These primitives keep installers, upgrade, and uninstall rollback-capable.
