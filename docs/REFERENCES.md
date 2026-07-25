---
name: buda-governing-references
purpose: Record Buda's normative and informative governing references.
description: Reference authority for the LLM Wiki operating model, canonical Open Knowledge Format, official Google Cloud overview, and approved GUIHO RFC revision.
created: 2026-07-26
owner: buda-docs
flags: []
tags:
  - references
  - okf
  - llm-wiki
keywords:
  - SPEC.md
  - Google Cloud
  - Karpathy
  - RFC 0002
---

# Governing references

1. [Karpathy's LLM Wiki gist](https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f)
   governs the operating pattern: agents compile sources into persistent,
   maintained knowledge and navigate that knowledge later.
2. [Google's canonical Open Knowledge Format `SPEC.md`](https://github.com/GoogleCloudPlatform/knowledge-catalog/blob/main/okf/SPEC.md)
   is normative for Buda's portable base OKF conformance contract.
3. [Google Cloud's official OKF announcement and overview](https://cloud.google.com/blog/products/data-analytics/how-the-open-knowledge-format-can-improve-data-sharing)
   is informative context and cannot override the canonical specification.
4. GUIHO RFC 0002 at commit
   `30698669a2e72f1ded575574b5b8ff7f0b9b5c6e` defines the approved Buda
   integration, CLI, qmd, agent-resource, packaging, and phased implementation
   contract.

Buda-specific rules are extensions and must never be reported as base OKF
requirements. When the operating workflow and portable file representation
overlap, the LLM Wiki pattern governs operating intent, OKF governs base format,
and the approved RFC defines their Buda integration.
