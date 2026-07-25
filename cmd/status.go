package cmd

import (
	"fmt"
	"time"

	"github.com/CGuiho/buda/internal/health"
	"github.com/spf13/cobra"
)

func NewStatusCommand(deps Dependencies, factories ...QMDFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show canonical Buda health and qmd state separately.",
		Args:  NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			repo, client, err := openQMD(deps, factories)
			if err != nil {
				return err
			}
			deps.Options.Wiki = repo.Root
			report, err := health.Scan(repo.Bundle, repo.Config.WikiID, time.Now().UTC())
			if err != nil {
				return RepositoryError("scan canonical OKF bundle", err)
			}
			qmdStatus, err := client.Status(command.Context())
			if err != nil {
				return externalError("read qmd status", err)
			}
			output := map[string]any{
				"command": "status", "wiki": repo.Root, "bundle": repo.Bundle,
				"qmd_project_directory": repo.QMDProject, "collection": repo.Collection,
				"canonical": report, "qmd": qmdStatus,
			}
			if JSONRequested(deps) {
				return WriteJSON(command, output)
			}
			fmt.Fprintf(command.OutOrStdout(), "wiki: %s\nbundle: %s\ncanonical OKF: conformant=%t\ncanonical Buda health: healthy=%t concepts=%d sources=%d findings=%d\n",
				repo.Root, repo.Bundle, report.Conformant, report.Healthy, report.Counts.Concepts, report.Counts.Sources, len(report.Findings))
			fmt.Fprintf(command.OutOrStdout(), "qmd: version=%s state=%s\n", qmdStatus.Version, qmdStatus.Output)
			return nil
		},
	}
}
