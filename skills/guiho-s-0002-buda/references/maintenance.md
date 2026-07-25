---
name: buda-skill-maintenance
purpose: Define repository health and qmd readiness maintenance.
description: Maintenance guidance separating base OKF conformance, Buda health, and qmd readiness.
created: 2026-07-26
owner: buda-skill-0002-references
flags: []
tags:
  - maintenance
  - health
keywords:
  - OKF conformance
  - qmd readiness
---

# Maintenance

Run `buda lint`, `buda index`, `buda status`, and `buda doctor` with the one
explicit `--wiki`. Keep base OKF conformance findings separate from stricter
Buda health and qmd readiness. Repair deterministic drift only when authorized;
create review items for semantic contradictions, gaps, and stale concepts.
