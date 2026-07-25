package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewDoctorCommand(deps Dependencies, factories ...QMDFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose Buda repository and qmd readiness without repair.",
		Args:  NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			repo, client, err := openQMD(deps, factories)
			if err != nil {
				return err
			}
			diagnostic, err := client.Doctor(command.Context())
			if err != nil {
				return externalError("diagnose qmd", err)
			}
			result := map[string]any{
				"command": "doctor", "wiki": repo.Root, "bundle": repo.Bundle,
				"qmd_project_directory": repo.QMDProject, "collection": repo.Collection,
				"configuration": "valid", "repository": "resolved", "qmd": diagnostic,
				"repair_performed": false,
			}
			if JSONRequested(deps) {
				return WriteJSON(command, result)
			}
			fmt.Fprintf(command.OutOrStdout(), "wiki: %s\nbundle: %s\nqmd project: %s\ncollection: %s\nconfiguration: valid\nrepository: resolved\nqmd version: %s\n", repo.Root, repo.Bundle, repo.QMDProject, repo.Collection, diagnostic.Version)
			if diagnostic.Output != "" {
				fmt.Fprintln(command.OutOrStdout(), diagnostic.Output)
			}
			fmt.Fprintln(command.OutOrStdout(), "repair performed: no")
			return nil
		},
	}
}
