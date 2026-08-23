---
subject: buda-internal-releasecatalog
description: Complete paginated release selection, channel matching, and mandatory convention-named lifecycle prompt validation.
parent: buda-internal
children: []
files:
  catalog.go: Exact SemVer, channel, complete-artifact, payload, launcher, and target selection.
documents: {}
tags: [release, lifecycle]
keywords: [stable, canary, alpha, beta, nightly]
flags: []
status: draft
---

#### &copy; 2026 [GUIHO](https://guiho.co) as represented by [Cristóvão GUIHO](https://guiho.co/cguiho) All Rights Reserved.

The selector rejects incomplete releases before installation or activation,
including releases missing either lifecycle prompt artifact.
