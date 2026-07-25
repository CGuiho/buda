package cmd

import (
	"fmt"
	"strings"

	"github.com/CGuiho/buda/internal/health"
	"github.com/CGuiho/buda/internal/ingest"
	"github.com/CGuiho/buda/internal/qmd"
	"github.com/spf13/cobra"
)

func NewIngestCommand(deps Dependencies, factories ...QMDFactory) *cobra.Command {
	var explicitSource, title, actor string
	command := &cobra.Command{
		Use:     "ingest",
		Short:   "Register one explicit durable source and create an agent work item.",
		Example: "  buda ingest --wiki ./wiki --source ./research.pdf --actor human:owner",
		Args:    NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			if explicitSource == "" || actor == "" {
				return UsageError("--source and --actor are required")
			}
			repo, client, err := openQMD(deps, factories)
			if err != nil {
				return err
			}
			deps.Options.Wiki = repo.Root
			if err := client.EnsureProject(command.Context()); err != nil {
				return externalError("validate qmd project and collection", err)
			}
			query := title
			if query == "" {
				query = explicitSource
			}
			matches, err := client.Search(command.Context(), qmd.SearchOptions{Text: query, Mode: qmd.ModeHybrid, Limit: 10})
			if err != nil {
				return externalError("retrieve existing qmd evidence for ingest", err)
			}
			candidates := make([]ingest.Evidence, 0, len(matches))
			for _, match := range matches {
				normalized := normalizeEvidence(repo, qmd.ModeHybrid, match)
				sources := make([]ingest.EvidenceSource, 0, len(normalized.Sources))
				for _, source := range normalized.Sources {
					sources = append(sources, ingest.EvidenceSource{ID: source.ID, Resource: source.Resource, Title: source.Title, Joined: source.Joined})
				}
				candidates = append(candidates, ingest.Evidence{
					Path: normalized.Path, DocumentID: normalized.DocumentID, Rank: normalized.Rank,
					Score: normalized.Score, Snippet: normalized.Snippet, Metadata: normalized.Metadata,
					Sources: sources, Warnings: normalized.Warnings,
				})
			}
			result, err := ingest.Run(command.Context(), repo, ingest.Request{
				Source: explicitSource, Title: title, Actor: actor, Now: dependencyNow(deps), HTTPClient: deps.HTTPClient, Candidates: candidates,
			})
			if err != nil {
				if strings.HasPrefix(strings.ToLower(explicitSource), "http://") || strings.HasPrefix(strings.ToLower(explicitSource), "https://") {
					return externalError("ingest explicit URL source", err)
				}
				return MutationError("ingest explicit local source", err)
			}
			report, err := health.Scan(repo.Bundle, repo.Config.WikiID, dependencyNow(deps))
			if err != nil {
				return MutationError("validate ingested source record", err)
			}
			indexResult, err := client.Update(command.Context())
			if err != nil {
				return externalError("refresh qmd index after ingest", err)
			}
			output := map[string]any{
				"command": "ingest", "wiki": repo.Root, "bundle": repo.Bundle,
				"qmd_project_directory": repo.QMDProject, "collection": repo.Collection,
				"ingest": result, "existing_candidates": candidates,
				"validation": report, "index": indexResult,
			}
			if JSONRequested(deps) {
				return WriteJSON(command, output)
			}
			fmt.Fprintf(command.OutOrStdout(), "wiki: %s\nbundle: %s\nqmd project: %s\ncollection: %s\nsource: %s\ndigest: %s\nartifact: %s\nsource concept: %s\nwork item: %s\n",
				repo.Root, repo.Bundle, repo.QMDProject, repo.Collection, result.SourceID, result.Digest, result.Artifact, result.SourceConcept, result.WorkItem)
			fmt.Fprintf(command.OutOrStdout(), "existing qmd candidates: %d\nvalidation: conformant=%t healthy=%t\nqmd index: refreshed\n",
				len(candidates), report.Conformant, report.Healthy)
			return nil
		},
	}
	command.Flags().StringVar(&explicitSource, "source", "", "Explicit local file path or http(s) URL")
	command.Flags().StringVar(&title, "title", "", "Human-readable source title")
	command.Flags().StringVar(&actor, "actor", "", "OKF actor registering the source")
	return command
}
