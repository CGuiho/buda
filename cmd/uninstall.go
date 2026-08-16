package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/CGuiho/buda/internal/repository"
	"github.com/CGuiho/buda/internal/uninstall"
	"github.com/spf13/cobra"
)

func newUninstallCommand(deps Dependencies) *cobra.Command {
	var dryRun, preserveConfig, preserveData, yes bool
	command := &cobra.Command{Use: "uninstall", Short: "Remove Buda-owned installation artifacts safely.", Args: NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			if strings.TrimSpace(deps.Options.Wiki) == "" {
				return RepositoryError("--wiki is required for uninstall; Buda never selects a wiki implicitly", nil)
			}
			wiki, err := repository.ResolveTarget(deps.Options.Wiki)
			if err != nil {
				return RepositoryError("resolve selected wiki", err)
			}
			layout, err := deps.InstallLayout()
			if err != nil {
				return MutationError("resolve Buda installation layout", err)
			}
			plan, err := uninstall.BuildPlan(layout, wiki, preserveConfig, preserveData)
			if err != nil {
				return MutationError("build uninstall plan", err)
			}
			if !JSONRequested(deps) {
				printGroupedPlan(command.OutOrStdout(), plan)
			}
			if dryRun {
				if JSONRequested(deps) {
					return WriteJSON(command, map[string]any{"command": "buda uninstall", "dry_run": true, "plan": plan})
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
			if err := uninstall.Apply(plan, layout.Temp); err != nil {
				return MutationError("apply Buda uninstall plan", err)
			}
			// The manifest-driven plan already removed the immutable payload.
			// This secondary removal covers only a verified legacy direct-binary
			// executable that still exists outside the managed layout.
			executable, err := absoluteExecutable(deps)
			if err == nil && executable != layout.Launcher {
				if _, statErr := os.Stat(executable); errors.Is(statErr, os.ErrNotExist) {
					// Already removed by the plan; nothing legacy remains.
				} else if statErr != nil {
					return MutationError("inspect legacy Buda executable", statErr)
				} else if deferred, removeErr := deps.RemoveExecutable(executable); removeErr != nil {
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

func printGroupedPlan(out io.Writer, plan uninstall.Plan) {
	fmt.Fprintln(out, "Buda Uninstall Plan:")
	var removeItems, preserveItems []uninstall.Item
	for _, item := range plan.Items {
		if item.Action == "REMOVE" {
			removeItems = append(removeItems, item)
		} else {
			preserveItems = append(preserveItems, item)
		}
	}
	fmt.Fprintln(out, "REMOVE:")
	if len(removeItems) == 0 {
		fmt.Fprintln(out, "  (none)")
	} else {
		for _, item := range removeItems {
			fmt.Fprintf(out, "  - %s (%s)\n", item.Path, item.Owner)
		}
	}
	fmt.Fprintln(out, "PRESERVE:")
	if len(preserveItems) == 0 {
		fmt.Fprintln(out, "  (none)")
	} else {
		for _, item := range preserveItems {
			fmt.Fprintf(out, "  - %s (%s)\n", item.Path, item.Owner)
		}
	}
}

func interactiveReader(reader any) bool {
	file, ok := reader.(*os.File)
	if !ok || file == nil {
		return false
	}
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}
