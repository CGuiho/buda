package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/CGuiho/buda/internal/qmd"
	"github.com/CGuiho/buda/internal/repository"
	"github.com/spf13/cobra"
)

type QMDClient interface {
	CheckCompatibility(context.Context) (qmd.Compatibility, error)
	EnsureProject(context.Context) error
	Ready(context.Context) (qmd.Compatibility, error)
	Update(context.Context) (qmd.IndexResult, error)
	Embed(context.Context) (qmd.IndexResult, error)
	Search(context.Context, qmd.SearchOptions) ([]qmd.Match, error)
	Get(context.Context, string) (qmd.Document, error)
	Status(context.Context) (qmd.Diagnostic, error)
	Doctor(context.Context) (qmd.Diagnostic, error)
}

type QMDFactory func(repository.Repository) (QMDClient, error)

func DefaultQMDFactory(repo repository.Repository) (QMDClient, error) {
	return qmd.New(qmd.Config{
		Executable:       repo.Config.QMD.Executable,
		WikiRoot:         repo.Root,
		BundleRoot:       repo.Bundle,
		ProjectDirectory: repo.QMDProject,
		Collection:       repo.Collection,
	}, qmd.OSRunner{})
}

func qmdFactory(factories []QMDFactory) QMDFactory {
	if len(factories) > 0 && factories[0] != nil {
		return factories[0]
	}
	return DefaultQMDFactory
}

func openQMD(deps Dependencies, factories []QMDFactory) (repository.Repository, QMDClient, error) {
	repo, err := repository.Open(deps.Options.Wiki)
	if err != nil {
		return repository.Repository{}, nil, RepositoryError("open selected wiki", err)
	}
	deps.Options.Wiki = repo.Root
	client, err := qmdFactory(factories)(repo)
	if err != nil {
		return repository.Repository{}, nil, RepositoryError("configure qmd adapter", err)
	}
	return repo, client, nil
}

func externalError(operation string, err error) error {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return &codedError{code: 130, message: operation, cause: err}
	}
	return &codedError{code: 4, message: operation, cause: err}
}

func NewIndexCommand(deps Dependencies, factories ...QMDFactory) *cobra.Command {
	var embed bool
	command := &cobra.Command{
		Use:   "index",
		Short: "Refresh the explicit wiki collection through qmd.",
		Args:  NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			repo, client, err := openQMD(deps, factories)
			if err != nil {
				return err
			}
			if err := client.EnsureProject(command.Context()); err != nil {
				return externalError("validate qmd project and collection", err)
			}
			indexResult, err := client.Update(command.Context())
			if err != nil {
				return externalError("refresh qmd index", err)
			}
			var embedResult *qmd.IndexResult
			if embed {
				result, err := client.Embed(command.Context())
				if err != nil {
					return externalError("refresh qmd embeddings", err)
				}
				embedResult = &result
			}
			output := map[string]any{
				"command": "index", "wiki": repo.Root, "bundle": repo.Bundle,
				"qmd_project_directory": repo.QMDProject, "collection": repo.Collection,
				"index": indexResult, "embed": embedResult,
			}
			if JSONRequested(deps) {
				return WriteJSON(command, output)
			}
			fmt.Fprintf(command.OutOrStdout(), "wiki: %s\nbundle: %s\nqmd project: %s\ncollection: %s\nindex: refreshed\n", repo.Root, repo.Bundle, repo.QMDProject, repo.Collection)
			if embedResult != nil {
				fmt.Fprintln(command.OutOrStdout(), "embeddings: refreshed by qmd")
			}
			return nil
		},
	}
	command.Flags().BoolVar(&embed, "embed", false, "Also delegate embedding refresh to qmd")
	return command
}
