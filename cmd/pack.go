package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	packdomain "github.com/CGuiho/buda/internal/pack"
	"github.com/CGuiho/buda/internal/repository"
	"github.com/spf13/cobra"
)

func NewPackCommand(deps Dependencies) *cobra.Command {
	var output string
	command := &cobra.Command{
		Use:     "pack",
		Short:   "Create a deterministic archive of the entire canonical OKF bundle.",
		Example: "  buda pack --wiki ./wiki --output ./wiki.zip",
		Args:    NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			if strings.TrimSpace(output) == "" {
				return UsageError("--output is required")
			}
			repo, err := repository.Open(deps.Options.Wiki)
			if err != nil {
				return RepositoryError("open selected wiki", err)
			}
			deps.Options.Wiki = repo.Root
			absoluteOutput, err := filepath.Abs(output)
			if err != nil {
				return MutationError("resolve pack output", err)
			}
			result, err := packdomain.Create(packdomain.Options{WikiRoot: repo.Root, BundleRoot: repo.Bundle, BundleName: repo.Config.Bundle, WikiID: repo.Config.WikiID, Output: absoluteOutput})
			if err != nil {
				return MutationError("create deterministic pack", err)
			}
			if JSONRequested(deps) {
				return WriteJSON(command, map[string]any{
					"command": "pack", "wiki": repo.Root, "bundle": repo.Bundle,
					"qmd_project_directory": repo.QMDProject, "collection": repo.Collection,
					"pack": result, "sharing_safety_certified": false,
				})
			}
			fmt.Fprintf(command.OutOrStdout(), "wiki: %s\nbundle: %s\nqmd project: %s\ncollection: %s\noutput: %s\nfiles: %d\nsha256: %s\n", repo.Root, repo.Bundle, repo.QMDProject, repo.Collection, result.Output, result.Files, result.ArchiveSHA256)
			fmt.Fprintln(command.OutOrStdout(), "notice: this pack does not certify that the wiki is safe to share")
			return nil
		},
	}
	command.Flags().StringVar(&output, "output", "", "Destination .zip archive path")
	return command
}
