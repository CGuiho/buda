---
name: buda-qmd-compatibility
purpose: Record the external qmd process-adapter compatibility contract.
description: Supported qmd range, upstream revision, project-local scoping, capability mapping, fixture provenance, and runtime-validation limits.
created: 2026-07-26
owner: buda-docs
flags: []
tags:
  - qmd
  - compatibility
keywords:
  - qmd 2.x
  - process adapter
  - project-local index
  - JSON fixtures
---

# qmd compatibility

Buda supports qmd versions `>=2.5.0` and `<3.0.0`. Its CLI argument mapping and
structured fixtures are pinned to official qmd upstream revision
`e428df76bc0274d9e93eb7ca3e95673315c42e90`. qmd is discovered from the
`qmd.executable` value in strict `buda.yaml`; the default is `qmd`. Buda invokes
the executable directly with an argument array and never through a shell.

The adapter follows qmd's documented project-local model: execution starts in
the selected wiki, `qmd init` owns `.qmd/index.yml` and `.qmd/index.sqlite`, and
one configured collection maps exactly to the canonical bundle. Buda never
reads qmd's SQLite files. Search, vector search, hybrid query, embed, and get
always carry the explicit collection name or a collection-qualified `qmd://`
URI. Project-local status/update/doctor operations are accepted only after the
single collection mapping has been verified.

| Buda capability | qmd CLI capability |
| --- | --- |
| compatibility | `qmd --version` and public help capability checks |
| initialize | `qmd init`, `qmd collection add/show` |
| index | `qmd update` in the validated single-collection project |
| embed | `qmd embed -c <collection>` |
| lexical query | `qmd search --format json -c <collection>` |
| semantic query | `qmd vsearch --format json -c <collection>` |
| hybrid query | `qmd query --format json -c <collection>` |
| get | `qmd multi-get qmd://<collection>/... --format json` |
| diagnostics | `qmd status`, `qmd doctor` after mapping validation |

Recorded `2.6.3` fixtures cover search ordering/scores, multi-get document
fields, status counts, and doctor checks including advisory CPU-only warnings.
Injected-runner tests assert exact arguments, working
directory, containment, version rejection, malformed output, child exit
context, and absence of fallback retrieval.

The current Windows development host does not have qmd installed. Therefore,
these fixtures and process-boundary tests are implemented validation, while
native qmd setup/model/index/query execution remains an explicit unverified
runtime check. Buda must not convert that absence into an internal retrieval
fallback or claim qmd runtime readiness.

Upstream references:

- [qmd README](https://github.com/tobi/qmd/blob/main/README.md)
- [qmd changelog](https://github.com/tobi/qmd/blob/main/CHANGELOG.md)
- [qmd package metadata](https://github.com/tobi/qmd/blob/main/package.json)
