# Buda structured documentation

## Root module

- [buda.xdocs.md](buda.xdocs.md): Buda repository and Go module descriptor.

## Modules

- [.github/](.github/): GitHub Actions CI and canonical-tag GitHub Release
  automation with no package-manager publication.
- [cmd/](cmd/): Cobra command construction and application assembly.
- [internal/](internal/): Repository, OKF, qmd, health, workflow, agent, help,
  maintenance, and packaging services.
- [devops/](devops/): Pure-Go cross-build and installation tooling.
- [skills/](skills/): Embedded `guiho-s-0002-buda` agent skill family.
- [prompts/](prompts/): Embedded Buda instruction and prompt resources.
- [docs/](docs/): Architecture, governing references, Convention 0001 audit,
  implementation plan, acceptance matrix, historical delivery record, and
  validation records.

All named module descriptors use the `buda-*` subject namespace and model
containment rather than runtime dependencies.
