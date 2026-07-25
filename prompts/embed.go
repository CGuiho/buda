package prompts

import "embed"

// FS contains Buda's versioned instruction and prompt resources.
//
//go:embed *.md
var FS embed.FS
