---
subject: buda-internal-qmd
description: Thin project-local qmd process adapter, compatibility gate, containment checks, normalized types, and versioned fixtures.
parent: buda-internal
children:
  - buda-internal-qmd-fixtures
files:
  adapter.go: Supported qmd command mapping, project setup, indexing, retrieval, status, and diagnostics.
  adapter_test.go: Adapter invocation, capability, error, and fixture tests.
  error.go: Structured external-process error type and stable diagnostics.
  path.go: Project-local path and collection containment validation.
  runner.go: Injectable process execution boundary.
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
