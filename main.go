package main

import (
	"fmt"
	"os"

	"github.com/CGuiho/buda/cmd"
)

var (
	version     = "dev"
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
		fmt.Fprintln(os.Stderr, err)
		os.Exit(cmd.ExitCode(err))
	}
}
