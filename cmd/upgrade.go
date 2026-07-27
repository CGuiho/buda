package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/CGuiho/buda/internal/repository"
	"github.com/CGuiho/buda/internal/selfmanage"
	"github.com/spf13/cobra"
)

func newUpgradeCommand(deps Dependencies, info BuildInfo) *cobra.Command {
	var requested string
	var dryRun bool
	command := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade the installed Buda native binary.",
		Args:  NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			wiki, err := optionalSelectedWiki(deps.Options.Wiki)
			if err != nil {
				return err
			}
			release, asset, manifest, err := resolveUpgrade(command.Context(), deps, requested, info.Target)
			if err != nil {
				return externalError("resolve Buda release", err)
			}
			if selfmanage.CompareVersions(release.Version, info.Version) == 0 {
				result := map[string]any{"command": "buda upgrade", "current_version": info.Version, "target_version": release.Version, "outcome": "already_current"}
				if JSONRequested(deps) {
					return WriteJSON(command, result)
				}
				fmt.Fprintf(command.OutOrStdout(), "Buda %s is already current; no replacement was performed.\n", info.Version)
				return nil
			}
			preview := map[string]any{
				"command": "buda upgrade", "current_version": info.Version,
				"target_version": release.Version, "asset": asset.Name,
				"download_url": asset.DownloadURL, "checksums_url": manifest.DownloadURL,
				"dry_run": true,
			}
			if dryRun {
				if JSONRequested(deps) {
					return WriteJSON(command, preview)
				}
				fmt.Fprintf(command.OutOrStdout(), "Current: %s\nTarget: %s\nAsset: %s\nURL: %s\nChecksums: %s\nDry run: true\n",
					info.Version, release.Version, asset.Name, asset.DownloadURL, manifest.DownloadURL)
				return nil
			}
			if !JSONRequested(deps) {
				fmt.Fprintf(command.OutOrStdout(), "Target: %s\nAsset: %s\nURL: %s\nChecksums: %s\n", release.Version, asset.Name, asset.DownloadURL, manifest.DownloadURL)
			}
			checksum, err := selfmanage.FetchChecksum(command.Context(), deps.HTTPClient, manifest.DownloadURL, asset.Name)
			if err != nil {
				return externalError("verify Buda release manifest", err)
			}
			executable, err := deps.Executable()
			if err != nil {
				return MutationError("determine Buda executable", err)
			}
			if !JSONRequested(deps) {
				fmt.Fprintf(command.OutOrStdout(), "Checksum verified: sha256:%s\nInstallation path: %s\n", checksum, executable)
			}
			var progress func(selfmanage.DownloadProgress)
			if !JSONRequested(deps) {
				progress = func(event selfmanage.DownloadProgress) {
					if event.Total > 0 {
						fmt.Fprintf(command.OutOrStdout(), "Download progress: %.1f%% (%d/%d bytes)\n", event.Percent, event.Bytes, event.Total)
					} else {
						fmt.Fprintf(command.OutOrStdout(), "Download progress: %d bytes\n", event.Bytes)
					}
				}
			}
			result, err := deps.UpgradeBinary(command.Context(), selfmanage.UpgradeOptions{
				Executable: executable, TargetVersion: release.Version, DownloadURL: asset.DownloadURL,
				ExpectedChecksum: checksum, Client: deps.HTTPClient, Wiki: wiki, Progress: progress,
			})
			result.CurrentVersion = info.Version
			result.Asset = asset.Name
			if err != nil {
				return MutationError("upgrade Buda", err)
			}
			if !result.Scheduled {
				if err := deps.ReconcileInstalled(result.Executable, wiki); err != nil {
					scheduledRollback, rollbackErr := deps.RollbackExecutable(result.Executable)
					if rollbackErr != nil {
						return MutationError("agent-resource reconciliation failed and automatic binary rollback failed", fmt.Errorf("reconcile: %w; rollback: %v", err, rollbackErr))
					}
					if scheduledRollback {
						return MutationError("agent-resource reconciliation failed; automatic binary rollback was scheduled", err)
					}
					return MutationError("agent-resource reconciliation failed; binary automatically rolled back", err)
				}
				if !JSONRequested(deps) {
					fmt.Fprintln(command.OutOrStdout(), "Agent resources: reconciled")
					fmt.Fprintf(command.OutOrStdout(), "Final version verified: buda v%s\n", release.Version)
				}
			}
			if JSONRequested(deps) {
				return WriteJSON(command, map[string]any{"command": "buda upgrade", "outcome": "succeeded", "result": result})
			}
			if result.Scheduled {
				fmt.Fprintf(command.OutOrStdout(), "Buda %s upgrade scheduled after this Windows process exits.\nRecovery: %s\n", release.Version, result.Recovery)
			} else {
				fmt.Fprintf(command.OutOrStdout(), "Buda upgraded from %s to %s.\nBackup: %s\n", info.Version, release.Version, result.Backup)
			}
			return nil
		},
	}
	command.Flags().StringVar(&requested, "version", "", "Select an exact semantic version")
	command.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without replacing the executable")
	command.AddCommand(newUpgradeCheckCommand(deps, info))
	command.AddCommand(newUpgradeListCommand(deps))
	command.AddCommand(newUpgradeRollbackCommand(deps))
	command.AddCommand(newWindowsReplacementCommand())
	command.AddCommand(newWindowsRollbackCommand())
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
			scheduled, err := deps.RollbackExecutable(executable)
			if err != nil {
				return MutationError("rollback Buda executable", err)
			}
			result := map[string]any{"command": "buda upgrade rollback", "executable": executable, "scheduled": scheduled, "outcome": "succeeded"}
			if JSONRequested(deps) {
				return WriteJSON(command, result)
			}
			if scheduled {
				fmt.Fprintln(command.OutOrStdout(), "Buda rollback scheduled after this Windows process exits.")
			} else {
				fmt.Fprintln(command.OutOrStdout(), "Restored the previous Buda executable.")
			}
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

func resolveUpgrade(ctx context.Context, deps Dependencies, requested, buildTarget string) (selfmanage.Release, selfmanage.Asset, selfmanage.Asset, error) {
	releases, err := (selfmanage.Catalog{Client: deps.HTTPClient}).Releases(ctx)
	if err != nil {
		return selfmanage.Release{}, selfmanage.Asset{}, selfmanage.Asset{}, err
	}
	assetName, err := selfmanage.AssetName(buildTarget)
	if err != nil {
		return selfmanage.Release{}, selfmanage.Asset{}, selfmanage.Asset{}, err
	}
	requested = strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(requested), "buda/"), "v")
	for _, release := range releases {
		if requested != "" && release.Version != requested {
			continue
		}
		if requested == "" && release.Prerelease {
			continue
		}
		var binary, manifest selfmanage.Asset
		for _, asset := range release.Assets {
			switch asset.Name {
			case assetName:
				binary = asset
			case "checksums.txt":
				manifest = asset
			}
		}
		if binary.Name != "" && manifest.Name != "" {
			return release, binary, manifest, nil
		}
		if requested != "" {
			return release, binary, manifest, fmt.Errorf("release %s lacks %s or checksums.txt", release.Version, assetName)
		}
	}
	return selfmanage.Release{}, selfmanage.Asset{}, selfmanage.Asset{}, fmt.Errorf("no compatible canonical Buda release found")
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
	commands := [][]string{{"agent", "skill", "update"}}
	if wiki != "" {
		commands = append(commands, []string{"agent", "instruction", "update", "--wiki", wiki})
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

func newWindowsReplacementCommand() *cobra.Command {
	var pid int
	var executable, candidate, backup, version, checksum, helper, wiki string
	command := &cobra.Command{
		Use: "__replace-windows", Hidden: true, Args: NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return selfmanage.CompleteWindowsReplacement(executable, candidate, backup, version, checksum, helper, wiki, pid)
		},
	}
	command.Flags().IntVar(&pid, "pid", 0, "Internal parent process ID")
	command.Flags().StringVar(&executable, "executable", "", "Internal executable path")
	command.Flags().StringVar(&candidate, "candidate", "", "Internal candidate path")
	command.Flags().StringVar(&backup, "backup", "", "Internal backup path")
	command.Flags().StringVar(&version, "target-version", "", "Internal target version")
	command.Flags().StringVar(&checksum, "checksum", "", "Internal checksum")
	command.Flags().StringVar(&helper, "helper", "", "Internal helper path")
	command.Flags().StringVar(&wiki, "selected-wiki", "", "Internal explicit wiki path")
	return command
}

func newWindowsRollbackCommand() *cobra.Command {
	var pid int
	var executable, helper string
	command := &cobra.Command{
		Use: "__rollback-windows", Hidden: true, Args: NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return selfmanage.CompleteWindowsRollback(executable, helper, pid)
		},
	}
	command.Flags().IntVar(&pid, "pid", 0, "Internal parent process ID")
	command.Flags().StringVar(&executable, "executable", "", "Internal executable path")
	command.Flags().StringVar(&helper, "helper", "", "Internal helper path")
	return command
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
