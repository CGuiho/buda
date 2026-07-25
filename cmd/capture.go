package cmd

import (
	"bytes"
	"fmt"
	"io"

	"github.com/CGuiho/buda/internal/capture"
	"github.com/CGuiho/buda/internal/health"
	"github.com/CGuiho/buda/internal/repository"
	"github.com/spf13/cobra"
)

func NewCaptureCommand(deps Dependencies, factories ...QMDFactory) *cobra.Command {
	var target, title, description, conceptType, text, actor string
	var replace bool
	command := &cobra.Command{
		Use:     "capture",
		Short:   "Save explicit user-directed text as one cited OKF concept.",
		Example: "  buda capture --wiki ./wiki --target concepts/decision.md --title Decision --actor human:owner --text \"Use stable IDs.\"",
		Args:    NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			if target == "" || title == "" || actor == "" {
				return UsageError("--target, --title, and --actor are required")
			}
			input := []byte(text)
			if text == "" {
				data, err := io.ReadAll(io.LimitReader(deps.In, 8<<20+1))
				if err != nil {
					return MutationError("read capture stdin", err)
				}
				if len(data) > 8<<20 {
					return UsageError("capture stdin exceeds 8 MiB")
				}
				input = data
			}
			if len(bytes.TrimSpace(input)) == 0 {
				return UsageError("provide --text or non-empty stdin")
			}
			repo, err := repository.Open(deps.Options.Wiki)
			if err != nil {
				return RepositoryError("open selected wiki", err)
			}
			deps.Options.Wiki = repo.Root
			result, err := capture.Run(repo, capture.Request{
				Target: target, Title: title, Description: description, Type: conceptType,
				Text: input, Actor: actor, Now: dependencyNow(deps), Replace: replace,
			})
			if err != nil {
				return MutationError("capture concept", err)
			}
			report, err := health.Scan(repo.Bundle, repo.Config.WikiID, dependencyNow(deps))
			if err != nil {
				return MutationError("validate captured concept", err)
			}
			client, err := qmdFactory(factories)(repo)
			if err != nil {
				return RepositoryError("configure qmd adapter", err)
			}
			if err := client.EnsureProject(command.Context()); err != nil {
				return externalError("validate qmd project and collection", err)
			}
			indexResult, err := client.Update(command.Context())
			if err != nil {
				return externalError("refresh qmd index after capture", err)
			}
			output := map[string]any{
				"command": "capture", "wiki": repo.Root, "bundle": repo.Bundle,
				"qmd_project_directory": repo.QMDProject, "collection": repo.Collection,
				"capture": result, "validation": report, "index": indexResult,
			}
			if JSONRequested(deps) {
				return WriteJSON(command, output)
			}
			fmt.Fprintf(command.OutOrStdout(), "wiki: %s\nbundle: %s\nqmd project: %s\ncollection: %s\nconcept: %s\nuid: %s\nsource: %s (%s)\n",
				repo.Root, repo.Bundle, repo.QMDProject, repo.Collection, result.Path, result.UID, result.SourceID, result.Digest)
			fmt.Fprintf(command.OutOrStdout(), "validation: conformant=%t healthy=%t\nqmd index: refreshed\n", report.Conformant, report.Healthy)
			return nil
		},
	}
	flags := command.Flags()
	flags.StringVar(&target, "target", "", "Explicit bundle-relative concept path")
	flags.StringVar(&title, "title", "", "Concept title")
	flags.StringVar(&description, "description", "", "One-sentence concept description")
	flags.StringVar(&conceptType, "type", "Note", "OKF concept type")
	flags.StringVar(&text, "text", "", "Text to save; omit to read stdin")
	flags.StringVar(&actor, "actor", "", "OKF actor that directed the capture")
	flags.BoolVar(&replace, "replace", false, "Explicitly approve replacing different target content")
	return command
}
