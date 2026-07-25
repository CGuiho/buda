---
name: buda-skill-ingest
purpose: Define source registration and affected-concept review through Buda.
description: Ingest guidance for provenance, work-item review, linting, and qmd indexing.
created: 2026-07-26
owner: buda-skill-0002-references
flags: []
tags:
  - ingest
  - provenance
keywords:
  - source registration
  - review work items
---

# Ingest

Register one explicit source with
`buda ingest --wiki <path> --source <value> --actor <actor> [--title <title>]`.
Use Buda query results to find materially affected concepts, distinguish the
immutable source from agent synthesis, preserve contradictions, attach source
IDs and claim footnotes, and update every affected concept. Run lint and index,
then present the repository diff for review. Never call qmd directly.
