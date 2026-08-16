package main

import (
	"os"

	"github.com/CGuiho/buda/internal/installlayout"
	"github.com/CGuiho/buda/internal/launcher"
)

func main() {
	layout, err := installlayout.Current()
	if err != nil {
		panic(err)
	}
	launcher.Exit(launcher.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr, layout))
}
