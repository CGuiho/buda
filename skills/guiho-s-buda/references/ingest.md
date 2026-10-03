---
name: buda-skill-ingest
purpose: Define source registration and affected-concept review through Buda.
description: Ingest guidance for provenance, work-item review, linting, and qmd indexing.
created: 2026-07-26
owner: buda-skill-references
flags: []
tags:
  - ingest
  - provenance
keywords:
  - source registration
  - review work items
---

#### &copy; 2026 [GUIHO](https://guiho.co) as represented by [Cristóvão GUIHO](https://guiho.co/cguiho) All Rights Reserved.

# Ingest

Register one explicit source with
`buda ingest --wiki <path> --source <value> --actor <actor> [--title <title>] [--mode lexical|semantic|hybrid]`.
The default is hybrid. Explicit `--mode lexical` uses qmd keyword search for
existing evidence without semantic models; semantic and hybrid keep their qmd
model prerequisites. Mode selection does not change source acquisition or perform
agent synthesis. Use ordinary human-readable titles; without a title, the source
value is the retrieval text. Invalid modes fail before qmd setup or source writes.

Inspect the returned `mode`, `ingest`, and normalized `existing_candidates`.
Unchanged source bytes and resource retain their raw artifact, source concept,
and original pending work item; its candidate snapshot is not refreshed on repeat.
Changed bytes create a new digest/source record and preserve earlier raw evidence.
Use the current returned candidates to review affected concepts, and verify
original resources, digests, and claim-footnote joins before citing them.

Use Buda query results to find materially affected concepts, distinguish the
immutable source from agent synthesis, preserve contradictions, attach source
IDs and claim footnotes, and update every affected concept. Run lint and index,
then present the repository diff for review. Never call qmd directly.
Use `buda query --wiki <path> --mode lexical --text <query>` and
`buda get --wiki <path> <result>` for cited model-free discovery. Full doctor still
reports missing semantic models, embeddings, or vectors; never hide those limits.
