---
name: buda-skill-retrieval
purpose: Define cited retrieval from one selected wiki through Buda.
description: Retrieval guidance for qmd-backed discovery, concept reads, citations, and evidence disclosure.
created: 2026-07-26
owner: buda-skill-0002-references
flags: []
tags:
  - retrieval
  - citations
keywords:
  - qmd query
  - cited evidence
---

# Cited retrieval

Call `buda query --wiki <path> --text <query>` for exactly one wiki. Retrieve
selected concepts through `buda get --wiki <path> <result>`. Answer only claims
supported by the returned concepts and cite both repository-relative concept
paths and source IDs or resources. Explicitly disclose stale, unverified,
deprecated, conflicting, or missing evidence.
