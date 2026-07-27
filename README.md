---
name: buda-readme
purpose: Introduce Buda and document its repository-agnostic wiki workflow.
description: Public overview of the Go/Cobra CLI, explicit-wiki contract, OKF knowledge model, qmd boundary, commands, and agent resources.
created: 2026-07-26
owner: buda-package
flags: []
tags:
  - readme
  - cli
  - knowledge
keywords:
  - Buda
  - Open Knowledge Format
  - qmd
  - LLM Wiki
---

# GUIHO Buda

GUIHO Buda is a CLI and agent skill collection to manage AI-maintained wiki. Buda help
agents save, navigate, read, validate, and maintain any one explicitly selected
AI-maintained wiki. Bult using public Go and Cobra. It follows the persistent knowledge operating pattern in
[Karpathy's LLM Wiki](https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f)
and stores portable knowledge according to Google's canonical
[Open Knowledge Format specification](https://github.com/GoogleCloudPlatform/knowledge-catalog/blob/main/okf/SPEC.md).
Google Cloud's
[official OKF overview](https://cloud.google.com/blog/products/data-analytics/how-the-open-knowledge-format-can-improve-data-sharing)
is useful explanatory context.

The name honors José António Ernesto, a former high-school philosophy professor
known as Buda. The note explains the name; it does not affect the technical
contract.

## Boundaries

Every repository command requires `--wiki <path>`. Buda never guesses a wiki,
stores a default wiki, searches several wikis, moves knowledge between
repositories, or publishes a wiki. Personal knowledge, research, team or
project knowledge, company knowledge, and application documentation are only
examples; Buda does not encode fixed use cases, visibility classes, or corpus
roles.

[qmd](https://github.com/tobi/qmd) is the external project-local indexing and
retrieval engine. Buda delegates lexical, semantic, and hybrid retrieval,
embeddings, ranking, reranking, document lookup, and index diagnostics to qmd.
Buda owns explicit repository selection, OKF-aware initialization, capture and
ingest orchestration, provenance, citations, health checks, stable output,
agent skills, instructions, and deterministic packaging.

## Commands

```text
buda init --wiki <path> --wiki-id <id>
buda capture --wiki <path> ...
buda ingest --wiki <path> --source <path-or-url> ...
buda lint --wiki <path>
buda index --wiki <path> [--embed]
buda query --wiki <path> --text <query> [--mode lexical|semantic|hybrid]
buda get --wiki <path> <concept-path-or-result-id>
buda status --wiki <path>
buda pack --wiki <path> --output <path>
buda doctor --wiki <path>
buda agent ...
```

Every command supports `--help`, `--help-tree`, positive
`--help-tree-depth`, and `--help-docs`. Root also supports `--version` and
`-v`. Repository commands provide stable text and one-document JSON output.

## Agent resource bootstrap

A successful bare `buda` invocation may schedule a hidden, non-blocking
reconciler for the embedded Buda skill in both global agent-skill locations. It
has no wiki context and never writes repository instructions.

A successful `buda --wiki <path>` invocation or successful repository command
may schedule the same global reconciliation and, because it carries an
explicitly resolved wiki, reconcile Buda's bounded instruction block in only
that selected wiki. The block identifies the selected wiki and bundle, declares
Buda the required tool for maintaining and retrieving it, requires every
repository operation to pass `--wiki`, and requires agents to evaluate evidence
and cite concept paths plus source IDs or resources. Agents maintain the wiki
through Buda rather than bypassing it or invoking qmd directly. The block also
requires weak evidence to be surfaced and says that repository instructions are
behavior rather than access control.

Bootstrap performs no qmd or network operation, does not discover repositories,
and never reads or writes another wiki. Worker startup or reconciliation failure
cannot change foreground output or exit status. Explicit `buda agent ...`, help,
version, and hidden-worker routes do not schedule reconciliation.

## Wiki layout

An initialized wiki contains strict `buda.yaml`, canonical OKF files under
`knowledge/`, disposable project-local qmd state under `.qmd/`, and disposable
Buda staging/report state under `.buda/`. Canonical Markdown and source evidence
remain readable without Buda or qmd.

## Development

Buda requires a supported Go toolchain and builds with `CGO_ENABLED=0`.

```powershell
go mod tidy
go test ./...
go vet ./...
go run . --help-tree
```

qmd is a separate runtime dependency. Buda does not vendor or silently install
it. The repository currently implements the approved MVP contract; releases,
tags, and publication are separate explicitly authorized operations.
