package cmd

import (
	"fmt"
	"os"

	"github.com/CGuiho/buda/internal/uninstall"
	"github.com/spf13/cobra"
)

func newUninstallCommand(deps Dependencies) *cobra.Command {
	var dryRun, preserveConfig, preserveData, yes bool
	command := &cobra.Command{Use: "uninstall", Short: "Remove Buda-owned installation artifacts safely.", Args: NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			wiki, err := optionalSelectedWiki(deps.Options.Wiki)
			if err != nil {
				return err
			}
			layout, err := deps.InstallLayout()
			if err != nil {
				return MutationError("resolve Buda installation layout", err)
			}
			plan, err := uninstall.BuildPlan(layout, wiki, preserveConfig, preserveData)
			if err != nil {
				return MutationError("build uninstall plan", err)
			}
			if JSONRequested(deps) && dryRun {
				return WriteJSON(command, map[string]any{"command": "buda uninstall", "dry_run": true, "plan": plan})
			}
			if dryRun {
				for _, item := range plan.Items {
					fmt.Fprintf(command.OutOrStdout(), "%s %s (%s)\n", item.Action, item.Path, item.Owner)
				}
				return nil
			}
			if !yes && !interactiveReader(deps.In) {
				return UsageError("uninstall is destructive; provide --yes when no interactive terminal is available")
			}
			if !yes {
				fmt.Fprint(command.OutOrStdout(), "Remove all Buda-owned artifacts listed above? [y/N] ")
				var answer string
				if _, err := fmt.Fscan(deps.In, &answer); err != nil || (answer != "y" && answer != "Y" && answer != "yes" && answer != "YES") {
					return UsageError("uninstall cancelled")
				}
			}
			if err := uninstall.Apply(plan); err != nil {
				return MutationError("apply Buda uninstall plan", err)
			}
			executable, err := absoluteExecutable(deps)
			if err == nil && executable != layout.Launcher {
				if deferred, removeErr := deps.RemoveExecutable(executable); removeErr != nil {
					return MutationError("remove legacy Buda executable", removeErr)
				} else if deferred {
					return MutationError("remove Buda executable", fmt.Errorf("removal was not synchronous"))
				}
			}
			if JSONRequested(deps) {
				return WriteJSON(command, map[string]any{"command": "buda uninstall", "outcome": "succeeded", "plan": plan})
			}
			fmt.Fprintln(command.OutOrStdout(), "Buda uninstall completed synchronously.")
			return nil
		},
	}
	command.Flags().BoolVar(&dryRun, "dry-run", false, "Preview the manifest-driven removal plan without changing files")
	command.Flags().BoolVar(&preserveConfig, "preserve-config", false, "Preserve global and selected-project configuration")
	command.Flags().BoolVar(&preserveData, "preserve-data", false, "Preserve persistent Buda data and databases")
	command.Flags().BoolVar(&yes, "yes", false, "Confirm destructive removal without prompting")
	return command
}

func interactiveReader(reader any) bool {
	file, ok := reader.(*os.File)
	if !ok || file == nil {
		return false
	}
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}
