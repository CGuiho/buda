---
name: buda-architecture
purpose: Define Buda's implemented component and ownership boundaries.
description: Architecture for explicit wiki resolution, OKF canonical files, qmd delegation, agent resources, deterministic packaging, and pure-Go distribution.
created: 2026-07-26
owner: buda-docs
flags: []
tags:
  - architecture
  - go
  - knowledge
keywords:
  - OKF
  - qmd adapter
  - Cobra
  - provenance
---

# Buda architecture

Buda opens exactly one explicit wiki per repository-facing command. Repository
resolution validates one `buda.yaml`, loads strict typed configuration, and
ensures the canonical bundle, qmd project directory, and Buda derived directory
remain contained within that wiki root.

## Ownership

- `internal/config` owns strict `buda.yaml` decoding and semantics.
- `internal/repository` owns explicit wiki resolution and path containment.
- `internal/okf` owns forward-compatible OKF parsing and Buda extension checks.
- `internal/source`, `internal/capture`, and `internal/ingest` own durable
  evidence registration and deterministic orchestration primitives.
- `internal/health` distinguishes base OKF conformance from stricter Buda
  health findings.
- `internal/qmd` is a process-only adapter. It scopes every operation to the
  selected collection and never implements retrieval itself.
- `internal/agent` owns embedded, versioned skill/instruction/prompt bytes and
  transactional installation.
- `internal/maintenance` owns the failure-isolated first-success reconciler.
- `internal/pack` owns deterministic canonical-bundle archives.
- `internal/help` renders command-tree and Markdown help from the live Cobra
  tree.

The `cmd` package assembles one new Cobra tree for every invocation. Command
handlers translate stable CLI input/output to focused service APIs. `main.go`
only supplies linker metadata, invokes the command tree, and maps exit codes.

## Canonical and derived state

Canonical state is the OKF bundle under `knowledge/`, including concept
frontmatter and bodies, reserved index/log files, source records, and immutable
or content-addressed raw evidence. `.qmd/` and `.buda/` are derived and may be
deleted without losing canonical knowledge.

Base OKF consumes unknown concept types and extension keys permissively. Buda's
strict maintenance profile lives under the `buda` frontmatter mapping and may
report health failures that are not base OKF conformance failures.

## qmd boundary

Buda launches the configured qmd executable without shell interpolation using
an injected runner and bounded context. The adapter preserves qmd rank, score,
snippet, and retrieval semantics; Buda only validates containment and adds OKF
evidence. Missing or incompatible qmd is an actionable dependency failure.
There is no fallback index or search implementation.

## Distribution boundary

Release tooling builds eight pure-Go executables for Linux AMD64/ARM64/ARMv7/
ARMv6, Darwin AMD64/ARM64, and Windows AMD64/ARM64. The embedded skill ZIP,
instruction Markdown, and checksum manifest complete the eleven-artifact
layout. Building these artifacts is not publishing them.
