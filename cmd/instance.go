package cmd

import (
	"os"

	"github.com/CGuiho/buda/internal/installlayout"
	"github.com/CGuiho/buda/internal/lifecycle"
)

func registerCurrentInstance(info BuildInfo) (func(), error) {
	if info.Target == "development" || info.Version == "dev" {
		return func() {}, nil
	}
	layout, err := installlayout.Current()
	if err != nil {
		return nil, err
	}
	executable, err := os.Executable()
	if err != nil {
		return nil, err
	}
	cleanup, err := lifecycle.RegisterInstance(layout.Instances, executable, info.Version)
	if err != nil {
		return nil, err
	}
	return func() { _ = cleanup() }, nil
}
