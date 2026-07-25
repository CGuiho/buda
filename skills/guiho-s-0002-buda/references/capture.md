---
name: buda-skill-capture
purpose: Define the explicit capture trigger and validated persistence workflow.
description: Capture guidance for provenance-preserving concept and evidence writes through Buda.
created: 2026-07-26
owner: buda-skill-0002-references
flags: []
tags:
  - capture
  - provenance
keywords:
  - explicit capture
  - OKF evidence
---

# Capture

Trigger only on an explicit request such as “remember this,” “save this
knowledge,” or “record this decision.” Confirm the intended wiki and call
`buda capture --wiki <path> --target <concept.md> --title <title> --actor <actor> --text <text>`
or provide text on stdin while keeping the other required flags. Preserve the actor,
time, provenance, Buda metadata, and claim evidence. Report the written concept
and evidence after lint and index succeed.
