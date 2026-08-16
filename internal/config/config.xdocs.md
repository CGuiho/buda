---
subject: buda-internal-config
description: Strict typed global/project YAML decoding, deterministic merge, defaults, and semantic validation.
parent: buda-internal
children: []
files:
  config.go: Effective configuration model, strict YAML decoding, defaults, and semantic checks.
  config_test.go: Strict decoding, precedence, global baseline, and invalid-configuration tests.
  dual.go: Distinct global/project contracts, inheritance, schema URLs, and atomic legacy migration.
  dual_test.go: Global/project merge, policy inheritance, and strict-field tests.
documents: {}
tags:
  - configuration
  - yaml
keywords:
  - buda.yaml
  - KnownFields
  - semantic validation
flags: []
status: draft
---
