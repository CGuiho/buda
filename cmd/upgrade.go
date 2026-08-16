package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/CGuiho/buda/internal/releasecatalog"
	"github.com/CGuiho/buda/internal/repository"
	"github.com/CGuiho/buda/internal/selfmanage"
	"github.com/CGuiho/buda/internal/upgrade"
	"github.com/spf13/cobra"
)

func newUpgradeCommand(deps Dependencies, info BuildInfo) *cobra.Command {
	var requested, channel string
	var dryRun bool
	command := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade the installed Buda native binary.",
		Args:  NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			selector := releasecatalog.Selector{Version: requested, Channel: channel}
			if err := selector.Validate(); err != nil {
				return UsageError("%v", err)
			}
			wiki, err := optionalSelectedWiki(deps.Options.Wiki)
			if err != nil {
				return err
			}
			if wiki == "" {
				return RepositoryError("--wiki is required for a complete upgrade; Buda never selects a wiki implicitly", nil)
			}
			if !JSONRequested(deps) {
				printRecovery(command, requested, channel, wiki)
			}
			selection, err := resolveUpgrade(command.Context(), deps, requested, channel, info.Target)
			if err != nil {
				if !JSONRequested(deps) {
					printRecovery(command, requested, channel, wiki)
				}
				return externalError("resolve Buda release", err)
			}
			release, asset, manifest := selection.Release, selection.Binary, selection.Manifest
			if selfmanage.CompareVersions(release.Version, info.Version) == 0 {
				result := map[string]any{"command": "buda upgrade", "current_version": info.Version, "target_version": release.Version, "outcome": "up-to-date", "recovery": recoveryCommand(release.Version, "", wiki)}
				if JSONRequested(deps) {
					return WriteJSON(command, result)
				}
				fmt.Fprintf(command.OutOrStdout(), "Buda %s is already current; no replacement was performed.\n", info.Version)
				if !JSONRequested(deps) {
					printRecovery(command, release.Version, "", wiki)
				}
				return nil
			}
			preview := map[string]any{
				"command": "buda upgrade", "current_version": info.Version,
				"target_version": release.Version, "asset": asset.Name,
				"download_url": asset.DownloadURL, "checksums_url": manifest.DownloadURL,
				"dry_run": true, "recovery": recoveryCommand(release.Version, "", wiki),
			}
			if dryRun {
				if JSONRequested(deps) {
					return WriteJSON(command, preview)
				}
				fmt.Fprintf(command.OutOrStdout(), "Current: %s\nTarget: %s\nAsset: %s\nURL: %s\nChecksums: %s\nDry run: true\n",
					info.Version, release.Version, asset.Name, asset.DownloadURL, manifest.DownloadURL)
				if !JSONRequested(deps) {
					printRecovery(command, release.Version, "", wiki)
				}
				return nil
			}
			layout, layoutErr := deps.InstallLayout()
			if layoutErr != nil {
				if !JSONRequested(deps) {
					printRecovery(command, release.Version, "", wiki)
				}
				return MutationError("resolve Buda installation layout", layoutErr)
			}
			if deps.UpgradeRelease == nil {
				if !JSONRequested(deps) {
					printRecovery(command, release.Version, "", wiki)
				}
				return MutationError("upgrade Buda synchronously", errors.New("no upgrade release transaction handler configured"))
			}
			transaction, txErr := deps.UpgradeRelease(command.Context(), upgrade.Options{
				Layout:         layout,
				Selection:      selection,
				Client:         deps.HTTPClient,
				CurrentVersion: info.Version,
				Target:         info.Target,
				OS:             runtime.GOOS,
				Arch:           runtime.GOARCH,
				CurrentPID:     os.Getpid(),
			})
			if txErr != nil {
				if !JSONRequested(deps) {
					printRecovery(command, release.Version, "", wiki)
				}
				return MutationError("upgrade Buda synchronously", txErr)
			}
			initErr := error(nil)
			if deps.ReconcileInstalled != nil {
				initErr = deps.ReconcileInstalled(transaction.PayloadPath, wiki)
			}
			if JSONRequested(deps) {
				outcome := "succeeded"
				if initErr != nil {
					outcome = "succeeded-init-failed"
				}
				if writeErr := WriteJSON(command, map[string]any{
					"command":    "buda upgrade",
					"outcome":    outcome,
					"result":     transaction,
					"init_error": errorMessage(initErr),
					"recovery":   recoveryCommand(transaction.InstalledVersion, "", wiki),
				}); writeErr != nil {
					return writeErr
				}
				if initErr != nil {
					return MutationError("Buda upgraded but explicit-wiki init failed", initErr)
				}
				return nil
			}
			fmt.Fprintf(command.OutOrStdout(), "Buda upgraded from %s to %s.\nLauncher: %s\nPayload: %s\n", info.Version, transaction.InstalledVersion, transaction.LauncherPath, transaction.PayloadPath)
			if initErr != nil {
				fmt.Fprintf(command.OutOrStdout(), "Post-upgrade explicit-wiki init failed; the verified release remains active: %v\n", initErr)
				printRecovery(command, transaction.InstalledVersion, "", wiki)
				return MutationError("Buda upgraded but explicit-wiki init failed", initErr)
			}
			fmt.Fprintln(command.OutOrStdout(), "Post-upgrade explicit-wiki init: verified")
			printRecovery(command, transaction.InstalledVersion, "", wiki)
			return nil
		},
	}
	command.Flags().StringVar(&requested, "version", "", "Select an exact semantic version")
	command.Flags().StringVar(&channel, "channel", "", "Select the highest published release in a channel")
	command.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without replacing the executable")
	command.AddCommand(newUpgradeCheckCommand(deps, info))
	command.AddCommand(newUpgradeListCommand(deps))
	command.AddCommand(newUpgradeRollbackCommand(deps))
	return command
}

func newUpgradeRollbackCommand(deps Dependencies) *cobra.Command {
	return &cobra.Command{
		Use: "rollback", Short: "Restore the previous Buda executable backup.", Args: NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			executable, err := absoluteExecutable(deps)
			if err != nil {
				return MutationError("determine Buda executable", err)
			}
			deferred, err := deps.RollbackExecutable(executable)
			if err != nil {
				return MutationError("rollback Buda executable", err)
			}
			if deferred {
				return MutationError("rollback Buda executable", fmt.Errorf("rollback did not complete synchronously"))
			}
			result := map[string]any{"command": "buda upgrade rollback", "executable": executable, "outcome": "succeeded"}
			if JSONRequested(deps) {
				return WriteJSON(command, result)
			}
			fmt.Fprintln(command.OutOrStdout(), "Restored the previous Buda executable.")
			return nil
		},
	}
}

func newUpgradeCheckCommand(deps Dependencies, info BuildInfo) *cobra.Command {
	return &cobra.Command{
		Use: "check", Short: "Check whether a newer stable Buda release exists.", Args: NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			releases, err := (selfmanage.Catalog{Client: deps.HTTPClient}).Releases(command.Context())
			if err != nil {
				return externalError("fetch Buda releases", err)
			}
			var latest string
			for _, release := range releases {
				if !release.Prerelease {
					latest = release.Version
					break
				}
			}
			if latest == "" {
				return externalError("fetch Buda releases", fmt.Errorf("no stable canonical Buda release found"))
			}
			available := selfmanage.CompareVersions(latest, info.Version) > 0
			result := map[string]any{"command": "buda upgrade check", "current_version": info.Version, "latest_version": latest, "update_available": available}
			if JSONRequested(deps) {
				return WriteJSON(command, result)
			}
			fmt.Fprintf(command.OutOrStdout(), "Current: %s\nLatest: %s\nUpdate available: %t\n", info.Version, latest, available)
			return nil
		},
	}
}

func newUpgradeListCommand(deps Dependencies) *cobra.Command {
	var page, perPage int
	var prereleases bool
	command := &cobra.Command{
		Use: "list", Short: "List canonical Buda GitHub releases.", Args: NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			if page < 1 || perPage < 1 || perPage > 100 {
				return UsageError("--page must be positive and --per-page must be between 1 and 100")
			}
			releases, err := (selfmanage.Catalog{Client: deps.HTTPClient}).Releases(command.Context())
			if err != nil {
				return externalError("fetch Buda releases", err)
			}
			filtered := make([]selfmanage.Release, 0, len(releases))
			for _, release := range releases {
				if prereleases || !release.Prerelease {
					filtered = append(filtered, release)
				}
			}
			start := (page - 1) * perPage
			if start > len(filtered) {
				start = len(filtered)
			}
			end := start + perPage
			if end > len(filtered) {
				end = len(filtered)
			}
			selected := filtered[start:end]
			if JSONRequested(deps) {
				return WriteJSON(command, map[string]any{"command": "buda upgrade list", "page": page, "per_page": perPage, "total": len(filtered), "releases": selected})
			}
			fmt.Fprintln(command.OutOrStdout(), "VERSION\tTAG\tPRERELEASE\tPUBLISHED\tRELEASE")
			for _, release := range selected {
				fmt.Fprintf(command.OutOrStdout(), "%s\t%s\t%t\t%s\t%s\n", release.Version, release.TagName, release.Prerelease, release.PublishedAt.UTC().Format("2006-01-02T15:04:05Z"), release.HTMLURL)
			}
			fmt.Fprintf(command.OutOrStdout(), "Page: %d\nTotal: %d\n", page, len(filtered))
			return nil
		},
	}
	command.Flags().IntVar(&page, "page", 1, "Select a positive result page")
	command.Flags().IntVar(&perPage, "per-page", 8, "Select 1 to 100 results per page")
	command.Flags().BoolVar(&prereleases, "pre-releases", false, "Include prerelease versions")
	return command
}

func resolveUpgrade(ctx context.Context, deps Dependencies, requested, channel, buildTarget string) (releasecatalog.Result, error) {
	result, err := releasecatalog.Select(ctx, selfmanage.Catalog{Client: deps.HTTPClient}, releasecatalog.Selector{Version: requested, Channel: channel}, buildTarget)
	if err != nil {
		return releasecatalog.Result{}, err
	}
	return result, nil
}

func optionalSelectedWiki(value string) (string, error) {
	if strings.TrimSpace(value) == "" {
		return "", nil
	}
	repo, err := repository.Open(value)
	if err != nil {
		return "", RepositoryError("open explicit --wiki", err)
	}
	return repo.Root, nil
}

func reconcileInstalledResources(executable, wiki string) error {
	commands := [][]string{{"agent", "skill", "upgrade"}}
	if wiki != "" {
		// init owns the complete explicit-wiki projection: it reconciles the
		// managed instruction and the project configuration in one operation.
		commands = append(commands, []string{"init", "--wiki", wiki})
	}
	for _, arguments := range commands {
		command := exec.Command(executable, arguments...)
		command.Env = append(os.Environ(), "BUDA_DISABLE_MAINTENANCE=1")
		output, err := command.CombinedOutput()
		if err != nil {
			return fmt.Errorf("run %s: %w (%s)", strings.Join(arguments, " "), err, strings.TrimSpace(string(output)))
		}
	}
	return nil
}

func recoveryCommand(version, channel, wiki string) string {
	version = strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(version), "buda/"), "v")
	selectorPS, selectorSH := "", ""
	if version != "" {
		selectorPS = "-Version '" + strings.ReplaceAll(version, "'", "''") + "'"
		selectorSH = "--version '" + strings.ReplaceAll(version, "'", "'\\''") + "'"
	} else if strings.TrimSpace(channel) != "" {
		selectorPS = "-Channel '" + strings.ReplaceAll(strings.TrimSpace(channel), "'", "''") + "'"
		selectorSH = "--channel '" + strings.ReplaceAll(strings.TrimSpace(channel), "'", "'\\''") + "'"
	}
	psWiki := "'" + strings.ReplaceAll(wiki, "'", "''") + "'"
	shWiki := "'" + strings.ReplaceAll(wiki, "'", "'\\''") + "'"
	return fmt.Sprintf("powershell.exe -NoProfile -ExecutionPolicy Bypass -Command \"& ([scriptblock]::Create((Invoke-RestMethod 'https://raw.githubusercontent.com/CGuiho/buda/main/devops/install.ps1'))) %s -Wiki %s\"\ncurl -fsSL https://raw.githubusercontent.com/CGuiho/buda/main/devops/install.sh | sh -s -- %s --wiki %s", selectorPS, psWiki, selectorSH, shWiki)
}

func printRecovery(command *cobra.Command, version, channel, wiki string) {
	fmt.Fprintf(command.OutOrStdout(), "If the upgrade fails, reinstall buda with this command:\n%s\n", recoveryCommand(version, channel, wiki))
}

func fileExists(path string) bool { _, err := os.Stat(path); return err == nil }

func errorMessage(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func absoluteExecutable(deps Dependencies) (string, error) {
	path, err := deps.Executable()
	if err != nil {
		return "", err
	}
	path, err = filepath.Abs(path)
	if err != nil {
		return "", err
	}
	if resolved, resolveErr := filepath.EvalSymlinks(path); resolveErr == nil {
		path = resolved
	}
	return path, nil
}
