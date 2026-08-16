---
subject: buda-internal-upgrade
description: Synchronous manifest-based release transaction with staging, checksum verification, immutable payload activation, journals, and rollback.
parent: buda-internal
children: []
files:
  transaction.go: Complete release download, strict manifest/checksum validation, candidate self-test, immutable version staging, atomic activation, and rollback transaction.
  transaction_test.go: Checksum completeness, manifest publication, and transaction safety tests.
documents: {}
tags:
  - go
  - lifecycle
  - upgrade
keywords:
  - synchronous upgrade
  - artifacts.json
  - checksums
flags: []
status: draft
---

This module owns the installed stable-launcher upgrade transaction. It never
replaces the running payload in place and leaves a recoverable journal when a
mutation fails.
