package cmd

import (
	"fmt"

	"github.com/CGuiho/buda/internal/selfmanage"
	"github.com/spf13/cobra"
)

func newUninstallCommand(deps Dependencies) *cobra.Command {
	var dryRun, keepResources bool
	command := &cobra.Command{
		Use: "uninstall", Short: "Remove the installed Buda binary and agent resources.", Args: NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			wiki, err := optionalSelectedWiki(deps.Options.Wiki)
			if err != nil {
				return err
			}
			executable, err := absoluteExecutable(deps)
			if err != nil {
				return MutationError("determine Buda executable", err)
			}
			result := map[string]any{"command": "buda uninstall", "executable": executable, "dry_run": dryRun, "keep_agent_resources": keepResources, "scheduled": false}
			if dryRun {
				if JSONRequested(deps) {
					return WriteJSON(command, result)
				}
				fmt.Fprintf(command.OutOrStdout(), "Would remove executable: %s\n", executable)
				if !keepResources {
					fmt.Fprintln(command.OutOrStdout(), "Would remove both global Buda skills.")
					if wiki != "" {
						fmt.Fprintf(command.OutOrStdout(), "Would remove Buda instruction blocks from: %s\n", wiki)
					}
				}
				return nil
			}
			var skills any = []any{}
			var instructions any = []any{}
			if !keepResources {
				skillResults, err := deps.Agents.UninstallSkill(false, "")
				if err != nil {
					return MutationError("remove global Buda skills", err)
				}
				skills = skillResults
				if wiki != "" {
					instructionResults, err := deps.Agents.RemoveInstructions(wiki)
					if err != nil {
						return MutationError("remove Buda instruction blocks", err)
					}
					instructions = instructionResults
				}
			}
			scheduled, err := deps.RemoveExecutable(executable)
			if err != nil {
				return MutationError("remove Buda executable", err)
			}
			result["scheduled"] = scheduled
			result["removed_skills"] = skills
			result["instruction_results"] = instructions
			if JSONRequested(deps) {
				return WriteJSON(command, result)
			}
			if scheduled {
				fmt.Fprintf(command.OutOrStdout(), "Scheduled executable removal: %s\n", executable)
			} else {
				fmt.Fprintf(command.OutOrStdout(), "Removed executable: %s\n", executable)
			}
			return nil
		},
	}
	command.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without deleting")
	command.Flags().BoolVar(&keepResources, "keep-agent-resources", false, "Keep global skills and explicit-wiki instruction blocks")
	command.AddCommand(newWindowsRemovalCommand())
	return command
}

func newWindowsRemovalCommand() *cobra.Command {
	var pid int
	var executable, helper string
	command := &cobra.Command{
		Use: "__remove-windows", Hidden: true, Args: NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return selfmanage.CompleteWindowsRemoval(executable, helper, pid)
		},
	}
	command.Flags().IntVar(&pid, "pid", 0, "Internal parent process ID")
	command.Flags().StringVar(&executable, "executable", "", "Internal executable path")
	command.Flags().StringVar(&helper, "helper", "", "Internal helper path")
	return command
}
