package cmd

import (
	"fmt"

	"github.com/CGuiho/buda/internal/health"
	"github.com/spf13/cobra"
)

func NewStatusCommand(deps Dependencies, factories ...QMDFactory) *cobra.Command {
	return &cobra.Command{
		Use:     "status",
		Short:   "Show canonical Buda health and qmd state separately.",
		Example: "  buda status --wiki ./wiki --json",
		Args:    NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			repo, client, err := openQMD(deps, factories)
			if err != nil {
				return err
			}
			deps.Options.Wiki = repo.Root
			report, err := health.Scan(repo.Bundle, repo.Config.WikiID, dependencyNow(deps))
			if err != nil {
				return RepositoryError("scan canonical OKF bundle", err)
			}
			qmdStatus, err := client.Status(command.Context())
			output := map[string]any{
				"command": "status", "wiki": repo.Root, "bundle": repo.Bundle,
				"qmd_project_directory": repo.QMDProject, "collection": repo.Collection,
				"canonical": report,
			}
			if err != nil {
				output["qmd"] = map[string]any{"state": "unavailable", "error": err.Error(), "detail": qmdErrorDetail(err)}
			} else {
				output["qmd"] = qmdStatus
			}
			if JSONRequested(deps) {
				if writeErr := WriteJSON(command, output); writeErr != nil {
					return writeErr
				}
				if err != nil {
					return &renderedError{cause: externalError("read qmd status", err)}
				}
				return nil
			}
			fmt.Fprintf(command.OutOrStdout(), "wiki: %s\nbundle: %s\nqmd project: %s\ncollection: %s\ncanonical OKF: conformant=%t\ncanonical Buda health: healthy=%t concepts=%d sources=%d findings=%d\n",
				repo.Root, repo.Bundle, repo.QMDProject, repo.Collection, report.Conformant, report.Healthy, report.Counts.Concepts, report.Counts.Sources, len(report.Findings))
			if err != nil {
				fmt.Fprintf(command.OutOrStdout(), "qmd: unavailable\nrepair: install a supported qmd, then run: buda index --wiki <path>\n")
				return externalError("read qmd status", err)
			}
			fmt.Fprintf(command.OutOrStdout(), "qmd: version=%s state=%s\n", qmdStatus.Version, qmdStatus.State)
			return nil
		},
	}
}
