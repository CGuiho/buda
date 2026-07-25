package cmd

import (
	"fmt"

	"github.com/CGuiho/buda/internal/qmd"
	"github.com/spf13/cobra"
)

type GetResult struct {
	Command             string       `json:"command"`
	Wiki                string       `json:"wiki"`
	Bundle              string       `json:"bundle"`
	QMDProjectDirectory string       `json:"qmd_project_directory"`
	Collection          string       `json:"collection"`
	Document            qmd.Document `json:"document"`
	Evidence            Evidence     `json:"evidence"`
}

func NewGetCommand(deps Dependencies, factories ...QMDFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "get <concept-path-or-result-id>",
		Short: "Retrieve one contained concept through qmd.",
		Args:  ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			repo, client, err := openQMD(deps, factories)
			if err != nil {
				return err
			}
			if _, err := client.Ready(command.Context()); err != nil {
				return externalError("validate qmd readiness", err)
			}
			document, err := client.Get(command.Context(), args[0])
			if err != nil {
				return externalError("retrieve qmd document", err)
			}
			evidence := normalizeEvidence(repo, "", qmd.Match{Rank: 1, Path: document.Path, Title: document.Title})
			result := GetResult{Command: "get", Wiki: repo.Root, Bundle: repo.Bundle, QMDProjectDirectory: repo.QMDProject, Collection: repo.Collection, Document: document, Evidence: evidence}
			if JSONRequested(deps) {
				return WriteJSON(command, result)
			}
			fmt.Fprintf(command.OutOrStdout(), "wiki: %s\nbundle: %s\nqmd project: %s\ncollection: %s\npath: %s\n", repo.Root, repo.Bundle, repo.QMDProject, repo.Collection, document.Path)
			fmt.Fprint(command.OutOrStdout(), document.Body)
			if document.Body == "" || document.Body[len(document.Body)-1] != '\n' {
				fmt.Fprintln(command.OutOrStdout())
			}
			return nil
		},
	}
}
