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
structured fixtures are verified against the official `@tobilu/qmd@2.5.3`
package, whose registry metadata identifies upstream revision
`53232770867ccb16538c2c6034e7d891dffc9ce3` and integrity
`sha512-wUKc4pSPDbgs7mV7JYE8/Qj1pNXXatJFV8byTT/T3yLaoAXheFtWu0BgSWwoWGhRkMmxl5Qyitt66NHgbMyeBA==`.
qmd 2.5.0 is the lower bound because that series supplies the diagnostic
contract used by `buda doctor`. qmd is discovered from the `qmd.executable`
value in strict `buda.yaml`; the default is `qmd`. Buda invokes the executable
directly with an argument array and never through a shell.

The adapter follows qmd's implemented project-local model: execution starts in
the selected wiki, `qmd init` owns `.qmd/index.yml` and `.qmd/index.sqlite`, and
qmd discovers that local configuration by walking upward from the process
working directory. One configured collection must map exactly to Buda's
canonical bundle. Buda reads the YAML collection mapping for containment
validation but never reads or writes qmd's SQLite files. Search, vector search,
hybrid query, embed, and get always carry the explicit collection name or a
collection-qualified `qmd://` URI. Project-local status/update/doctor operations
are accepted only after the single collection mapping has been verified.

| Buda capability | qmd CLI capability |
| --- | --- |
| compatibility | `qmd --version`, supported-range gating, and public help capability checks |
| initialize | `qmd init`, `qmd collection add/show` |
| index | `qmd update` in the validated single-collection project |
| embed | `qmd embed -c <collection>` |
| lexical query | `qmd search --format json -c <collection>` |
| semantic query | `qmd vsearch --format json -c <collection>` |
| hybrid query | `qmd query --format json -c <collection>` |
| get | `qmd multi-get qmd://<collection>/... --format json` for paths; `qmd get #<docid> --no-line-numbers` for stable result ids |
| diagnostics | `qmd status`, `qmd doctor` after mapping validation; doctor availability is version-gated because root help omits it |

Recorded `2.5.3` fixtures cover canonical `qmd://` search paths,
ordering/scores, multi-get document identifiers and fields, status counts, and
doctor checks. Injected-runner tests assert exact arguments, working directory,
containment, version rejection, malformed output, child exit context, and
absence of fallback retrieval. The opt-in native test is enabled with
`BUDA_QMD_E2E=1` and, when needed, `BUDA_QMD_EXECUTABLE=<path>`.

The official 2.5.3 npm package was exercised on Windows with Node 22 through
the adapter boundary. The native check created only the selected wiki's
`.qmd/index.yml` and `.qmd/index.sqlite`, added the canonical collection,
updated one Markdown document, performed lexical JSON search, retrieved the
document by qmd id, parsed status, and parsed doctor output. Semantic and
hybrid execution require qmd's external model downloads and were not fetched
for this compatibility check. Buda does not replace that external requirement
with fallback retrieval.

Upstream references:

- [qmd README](https://github.com/tobi/qmd/blob/main/README.md)
- [qmd changelog](https://github.com/tobi/qmd/blob/main/CHANGELOG.md)
- [qmd package metadata](https://github.com/tobi/qmd/blob/main/package.json)
- [qmd 2.5.3 npm package](https://www.npmjs.com/package/@tobilu/qmd/v/2.5.3)
