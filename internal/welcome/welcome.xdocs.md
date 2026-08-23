---
subject: buda-internal-welcome
description: Borderless Buda CLI hello-window rendering and terminal color policy.
parent: buda-internal
children: []
files:
  welcome.go: Deterministic borderless hello window using the approved Buda palette 2F6690, 3A7CA5, D9DCD6, 16425B, 81C3D7 with platform/version normalization and NO_COLOR/TERM handling.
  welcome_test.go: Exact hello-window output, complete palette, and no-color contract tests.
documents: {}
tags:
  - welcome
  - cli
keywords:
  - hello window
  - Buda welcome
  - ANSI palette
  - NO_COLOR
flags: []
status: draft
---

#### &copy; 2026 [GUIHO](https://guiho.co) as represented by [Cristóvão GUIHO](https://guiho.co/cguiho) All Rights Reserved.

This package owns the deterministic borderless hello window shown only by a bare `buda` invocation. It prints the Buda block logo, the AI-maintained OKF wiki tagline, and the GUIHO attribution, then lists platform, version, and the help hint. Color uses the approved palette when the terminal is compatible; NO_COLOR and TERM=dumb always disable color. Help, version, and explicit subcommands never print the hello window.
