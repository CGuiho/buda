---
subject: buda-internal-qmd
description: Thin project-local qmd process adapter, official-package compatibility gate, containment checks, normalized types, and versioned fixtures.
parent: buda-internal
children:
  - buda-internal-qmd-fixtures
files:
  adapter.go: Supported qmd command mapping, project setup, indexing, retrieval, status, and diagnostics.
  adapter_test.go: Adapter invocation, capability, error, and fixture tests.
  e2e_test.go: Opt-in native process test covering project-local initialization, indexing, lexical retrieval, get, status, and doctor.
  error.go: Structured external-process error type and stable diagnostics.
  path.go: Project-local path and collection containment validation.
  runner.go: Injectable process execution boundary.
  runner_test.go: Case-insensitive PWD replacement tests for deterministic qmd project selection.
  types.go: Adapter request, response, and normalized result types.
  version.go: Supported qmd semantic-version parsing and range checks.
documents: {}
tags:
  - qmd
  - process-adapter
keywords:
  - external retrieval
  - compatibility fixtures
  - project local
flags: []
status: draft
---
