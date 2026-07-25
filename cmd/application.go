package cmd

import "github.com/spf13/cobra"

// NewApplicationCommands is the single composition point for Buda's public
// repository commands. Domain command implementations are added here during
// integration; keeping composition separate preserves a fresh testable tree.
func NewApplicationCommands(_ Dependencies) []*cobra.Command {
	return nil
}
