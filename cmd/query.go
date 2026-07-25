package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/CGuiho/buda/internal/okf"
	"github.com/CGuiho/buda/internal/qmd"
	"github.com/CGuiho/buda/internal/repository"
	"github.com/spf13/cobra"
)

type EvidenceSource struct {
	ID       string `json:"id"`
	Resource string `json:"resource"`
	Title    string `json:"title,omitempty"`
	Joined   bool   `json:"claim_footnote_joined"`
}

type Evidence struct {
	Rank        int              `json:"rank"`
	Path        string           `json:"path"`
	DocumentID  string           `json:"document_id"`
	Score       float64          `json:"score"`
	Mode        qmd.SearchMode   `json:"mode"`
	Title       string           `json:"title,omitempty"`
	Snippet     string           `json:"snippet,omitempty"`
	Line        int              `json:"line,omitempty"`
	Explanation any              `json:"explanation,omitempty"`
	Metadata    map[string]any   `json:"metadata,omitempty"`
	Sources     []EvidenceSource `json:"sources,omitempty"`
	Warnings    []string         `json:"warnings,omitempty"`
}

type QueryResult struct {
	Command             string         `json:"command"`
	Wiki                string         `json:"wiki"`
	Bundle              string         `json:"bundle"`
	QMDProjectDirectory string         `json:"qmd_project_directory"`
	Collection          string         `json:"collection"`
	Mode                qmd.SearchMode `json:"mode"`
	Query               string         `json:"query"`
	Results             []Evidence     `json:"results"`
}

func NewQueryCommand(deps Dependencies, factories ...QMDFactory) *cobra.Command {
	var text, mode string
	var limit int
	var explain bool
	command := &cobra.Command{
		Use:   "query",
		Short: "Retrieve normalized OKF evidence through one explicit qmd collection.",
		Args:  NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			if strings.TrimSpace(text) == "" {
				return UsageError("--text is required")
			}
			searchMode := qmd.SearchMode(mode)
			switch searchMode {
			case qmd.ModeLexical, qmd.ModeSemantic, qmd.ModeHybrid:
			default:
				return UsageError("--mode must be lexical, semantic, or hybrid")
			}
			if limit < 1 {
				return UsageError("--limit must be a positive integer")
			}
			repo, client, err := openQMD(deps, factories)
			if err != nil {
				return err
			}
			if _, err := client.Ready(command.Context()); err != nil {
				return externalError("validate qmd readiness", err)
			}
			matches, err := client.Search(command.Context(), qmd.SearchOptions{Mode: searchMode, Text: text, Limit: limit, Explain: explain})
			if err != nil {
				return externalError("query qmd", err)
			}
			evidence := make([]Evidence, 0, len(matches))
			for _, match := range matches {
				evidence = append(evidence, normalizeEvidence(repo, searchMode, match))
			}
			result := QueryResult{Command: "query", Wiki: repo.Root, Bundle: repo.Bundle, QMDProjectDirectory: repo.QMDProject, Collection: repo.Collection, Mode: searchMode, Query: text, Results: evidence}
			if JSONRequested(deps) {
				return WriteJSON(command, result)
			}
			fmt.Fprintf(command.OutOrStdout(), "wiki: %s\nbundle: %s\nqmd project: %s\ncollection: %s\nmode: %s\nresults: %d\n", repo.Root, repo.Bundle, repo.QMDProject, repo.Collection, searchMode, len(evidence))
			for _, item := range evidence {
				fmt.Fprintf(command.OutOrStdout(), "%d. %s score=%g id=%s\n", item.Rank, item.Path, item.Score, item.DocumentID)
				if item.Title != "" {
					fmt.Fprintf(command.OutOrStdout(), "   title: %s\n", item.Title)
				}
				if item.Snippet != "" {
					fmt.Fprintf(command.OutOrStdout(), "   snippet: %s\n", strings.ReplaceAll(item.Snippet, "\n", " "))
				}
				for _, warning := range item.Warnings {
					fmt.Fprintf(command.OutOrStdout(), "   warning: %s\n", warning)
				}
			}
			return nil
		},
	}
	command.Flags().StringVar(&text, "text", "", "Query text delegated to qmd")
	command.Flags().StringVar(&mode, "mode", string(qmd.ModeHybrid), "qmd retrieval mode: lexical, semantic, or hybrid")
	command.Flags().IntVar(&limit, "limit", 20, "Maximum qmd candidates to normalize")
	command.Flags().BoolVar(&explain, "explain", false, "Include qmd hybrid retrieval explanation when available")
	return command
}

func normalizeEvidence(repo repository.Repository, mode qmd.SearchMode, match qmd.Match) Evidence {
	evidence := Evidence{Rank: match.Rank, Path: match.Path, DocumentID: match.DocumentID, Score: match.Score, Mode: mode, Title: match.Title, Snippet: match.Snippet, Line: match.Line, Explanation: match.Explanation}
	absolute, err := repository.ResolveContained(repo.Bundle, filepath.FromSlash(match.Path))
	if err != nil {
		evidence.Warnings = append(evidence.Warnings, "reject matched canonical path: "+err.Error())
		return evidence
	}
	data, err := os.ReadFile(absolute)
	if err != nil {
		evidence.Warnings = append(evidence.Warnings, "open matched canonical file: "+err.Error())
		return evidence
	}
	document, err := okf.ParseConcept(match.Path, data)
	if err != nil {
		evidence.Warnings = append(evidence.Warnings, "matched file is not a valid OKF concept: "+err.Error())
		return evidence
	}
	metadata, err := document.MetadataCopy()
	if err != nil {
		evidence.Warnings = append(evidence.Warnings, "decode OKF metadata: "+err.Error())
	} else {
		evidence.Metadata = metadata
	}
	sources, err := document.Sources()
	if err != nil {
		evidence.Warnings = append(evidence.Warnings, "decode OKF sources: "+err.Error())
	} else {
		body := string(document.Body)
		for _, source := range sources {
			joined := claimFootnoteJoined(body, source.ID)
			evidence.Sources = append(evidence.Sources, EvidenceSource{ID: source.ID, Resource: source.Resource, Title: source.Title, Joined: joined})
			if !joined {
				evidence.Warnings = append(evidence.Warnings, fmt.Sprintf("source %q is not joined to both a claim and footnote definition", source.ID))
			}
		}
	}
	if _, exists, err := document.Buda(); err != nil {
		evidence.Warnings = append(evidence.Warnings, "decode Buda metadata: "+err.Error())
	} else if !exists {
		evidence.Warnings = append(evidence.Warnings, "Buda extension metadata is missing")
	}
	return evidence
}

func claimFootnoteJoined(body, id string) bool {
	if id == "" {
		return false
	}
	escaped := regexp.QuoteMeta(id)
	definition := regexp.MustCompile(`(?m)^\[\^` + escaped + `\]:`).FindStringIndex(body)
	if definition == nil {
		return false
	}
	withoutDefinition := body[:definition[0]] + body[definition[1]:]
	return strings.Contains(withoutDefinition, "[^"+id+"]")
}
