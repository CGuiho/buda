package examples

import "embed"

// FS contains Buda's configuration example files.
//
//go:embed *.yaml
var FS embed.FS
