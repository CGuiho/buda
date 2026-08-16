---
subject: buda-internal-launcher
description: Stable launcher pointer resolution, payload execution, and verified fallback.
parent: buda-internal
children: []
files:
  launcher.go: Strict active/previous pointer decoding and exact exit-code forwarding.
documents: {}
tags: [launcher, lifecycle]
keywords: [current.json, fallback, immutable payload]
flags: []
status: draft
---

The launcher performs no network work and falls back only before an active
payload successfully starts.
