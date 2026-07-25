package cmd

import (
	"fmt"

	"github.com/CGuiho/buda/internal/health"
	"github.com/CGuiho/buda/internal/repository"
	"github.com/spf13/cobra"
)

func NewLintCommand(deps Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:     "lint",
		Short:   "Report base OKF conformance separately from Buda health.",
		Example: "  buda lint --wiki ./wiki --json",
		Args:    NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			repo, err := repository.Open(deps.Options.Wiki)
			if err != nil {
				return RepositoryError("open selected wiki", err)
			}
			deps.Options.Wiki = repo.Root
			report, err := health.Scan(repo.Bundle, repo.Config.WikiID, dependencyNow(deps))
			if err != nil {
				return RepositoryError("scan canonical OKF bundle", err)
			}
			if JSONRequested(deps) {
				if err := WriteJSON(command, map[string]any{
					"command": "lint", "wiki": repo.Root, "bundle": repo.Bundle,
					"qmd_project_directory": repo.QMDProject, "collection": repo.Collection,
					"report": report,
				}); err != nil {
					return err
				}
			} else {
				fmt.Fprintf(command.OutOrStdout(), "wiki: %s\nbundle: %s\nqmd project: %s\ncollection: %s\nOKF conformant: %t\nBuda healthy: %t\nconcepts: %d\nsources: %d\nfindings: %d\n",
					repo.Root, repo.Bundle, repo.QMDProject, repo.Collection, report.Conformant, report.Healthy,
					report.Counts.Concepts, report.Counts.Sources, len(report.Findings))
				for _, finding := range report.Findings {
					fmt.Fprintf(command.OutOrStdout(), "%s %s %s %s: %s\n", finding.Category, finding.Severity, finding.Code, finding.Path, finding.Message)
				}
			}
			if !report.Conformant || !report.Healthy {
				return &codedError{code: 1, message: "wiki lint reported conformance or health findings"}
			}
			return nil
		},
	}
}
