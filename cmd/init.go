package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/CGuiho/buda/internal/config"
	"github.com/CGuiho/buda/internal/repository"
	"github.com/spf13/cobra"
)

func NewInitCommand(deps Dependencies, factories ...QMDFactory) *cobra.Command {
	var wikiID string
	command := &cobra.Command{
		Use:     "init",
		Short:   "Initialize one explicit OKF-compatible Buda wiki.",
		Example: "  buda init --wiki ./wiki --wiki-id team-knowledge",
		Args:    NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			if wikiID == "" {
				return UsageError("--wiki-id is required")
			}
			root, err := repository.ResolveTarget(deps.Options.Wiki)
			if err != nil {
				return RepositoryError("resolve initialization target", err)
			}
			configuration := config.Default(wikiID)
			candidate := repository.Repository{
				Root: root, ConfigPath: filepath.Join(root, config.FileName),
				Bundle:     filepath.Join(root, configuration.Bundle),
				QMDProject: filepath.Join(root, configuration.QMD.ProjectDirectory),
				Derived:    filepath.Join(root, configuration.Derived),
				Collection: configuration.QMD.Collection, Config: configuration,
			}
			client, err := qmdFactory(factories)(candidate)
			if err != nil {
				return RepositoryError("configure qmd adapter", err)
			}
			if _, err := client.CheckCompatibility(command.Context()); err != nil {
				return externalError("validate qmd compatibility before initialization", err)
			}
			result, err := repository.Initialize(root, repository.InitOptions{
				WikiID: wikiID,
				Now:    dependencyNow(deps),
				BeforeCommit: func(candidate repository.Repository) error {
					if err := client.EnsureProject(command.Context()); err != nil {
						return externalError("validate qmd and establish project-local collection", err)
					}
					return nil
				},
			})
			if err != nil {
				if _, ok := err.(exitCoder); ok {
					return err
				}
				return MutationError("initialize selected wiki", err)
			}
			deps.Options.Wiki = result.Repository.Root
			skills, err := deps.Agents.InstallSkill(false, "")
			if err != nil {
				return MutationError("install embedded Buda skill globally", err)
			}
			instructions, err := deps.Agents.ApplyInstructions(result.Repository.Root)
			if err != nil {
				return MutationError("apply embedded Buda repository instructions", err)
			}
			output := map[string]any{
				"command": "init", "wiki": result.Repository.Root,
				"bundle": result.Repository.Bundle, "qmd_project_directory": result.Repository.QMDProject,
				"collection": result.Repository.Collection, "wiki_id": result.Repository.Config.WikiID,
				"created": result.Created, "unchanged": result.Unchanged,
				"skills": skills, "instructions": instructions,
			}
			if JSONRequested(deps) {
				return WriteJSON(command, output)
			}
			fmt.Fprintf(command.OutOrStdout(), "wiki: %s\nbundle: %s\nqmd project: %s\ncollection: %s\nwiki id: %s\n",
				result.Repository.Root, result.Repository.Bundle, result.Repository.QMDProject,
				result.Repository.Collection, result.Repository.Config.WikiID)
			if result.Unchanged {
				fmt.Fprintln(command.OutOrStdout(), "initialization: already current")
			} else {
				fmt.Fprintf(command.OutOrStdout(), "initialized: %d paths\n", len(result.Created))
			}
			fmt.Fprintln(command.OutOrStdout(), "qmd: compatible project-local collection established")
			fmt.Fprintln(command.OutOrStdout(), "agent resources: reconciled")
			return nil
		},
	}
	command.Flags().StringVar(&wikiID, "wiki-id", "", "Immutable user-assigned identifier for this wiki")
	return command
}
