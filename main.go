package main

import (
	"fmt"
	"os"

	"github.com/CGuiho/buda/cmd"
)

var (
	// Keep source builds SemVer-compatible. Release builds replace this value
	// with the exact tag version through ldflags, while local `go run` and
	// developer binaries still obey the raw-version CLI contract.
	version     = "0.2.0-dev.0"
	commit      = "unknown"
	buildDate   = "unknown"
	buildTarget = "development"
)

func main() {
	info := cmd.BuildInfo{
		Version: version,
		Commit:  commit,
		Date:    buildDate,
		Target:  buildTarget,
	}
	if err := cmd.Execute(info); err != nil {
		if !cmd.IsErrorRendered(err) {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(cmd.ExitCode(err))
	}
}
