package schemas

import "embed"

// FS contains Buda's version-pinned configuration JSON schemas.
//
//go:embed *.json
var FS embed.FS
