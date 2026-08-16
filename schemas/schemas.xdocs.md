---
subject: buda-schemas
description: Version-pinned JSON Schemas for project and global Buda configuration.
parent: buda-package
children: []
files:
  embed.go: Embedded filesystem exporting the version-pinned JSON Schemas.
  embed_test.go: Validation tests for embedded JSON schemas and runtime configuration parity.
  buda.schema.json: Offline and editor-discoverable schema for explicit-wiki project configuration.
  buda.global.schema.json: Offline and editor-discoverable schema for user-wide Buda configuration.
documents: {}
tags: [schemas, configuration]
keywords: [buda.yaml, buda.global.yaml, JSON Schema]
flags: []
status: draft
---

Schemas are embedded and published in every complete release. Runtime
validation never depends on network access.
