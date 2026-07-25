package cmd

import "github.com/spf13/cobra"

// NewApplicationCommands is the single composition point for Buda's public
// repository commands. Every call constructs fresh commands against the same
// injected dependency set; no mutable global command registry exists.
func NewApplicationCommands(deps Dependencies) []*cobra.Command {
	return []*cobra.Command{
		NewInitCommand(deps),
		NewCaptureCommand(deps),
		NewIngestCommand(deps),
		NewLintCommand(deps),
		NewIndexCommand(deps),
		NewQueryCommand(deps),
		NewGetCommand(deps),
		NewStatusCommand(deps),
		NewPackCommand(deps),
		NewDoctorCommand(deps),
	}
}
