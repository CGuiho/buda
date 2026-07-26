---
name: buda
purpose: Provide the bounded repository instruction block for one selected Buda wiki.
description: Universal repository instructions for one explicitly selected Buda wiki.
created: 2026-07-26
version: "0.1.0"
owner: buda-prompts
flags: []
tags:
  - repository-instructions
  - knowledge
keywords:
  - Buda
  - explicit wiki
  - OKF
---

## Buda Wiki

This repository is the Buda wiki `{{WIKI_ID}}`. Its canonical OKF bundle is
`{{BUNDLE}}`.

Buda is the required tool for maintaining and retrieving this wiki. Load the
`guiho-s-0002-buda` skill for explicit requests to remember, save, ingest,
find, cite, curate, or maintain knowledge. Pass this repository path through
`--wiki` on every Buda repository command. Maintain the wiki through Buda;
do not bypass Buda or invoke qmd directly. Never infer another wiki, search
multiple repositories, or copy knowledge between repositories.

For capture and ingest, preserve source provenance and claim citations, then
run `buda lint --wiki <path>` and `buda index --wiki <path>`. For retrieval,
evaluate the returned evidence before answering and cite repository-relative
concept paths plus source IDs or resources. Surface unverified, stale,
deprecated, conflicting, or missing evidence.

Repository instruction files define agent behavior; they are not access
control. Repository-specific policy belongs outside this managed block.
