// Package releasecatalog provides one release selection contract for Cobra
// upgrade and the shell installers' fixture-compatible behavior.
package releasecatalog

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/CGuiho/buda/internal/selfmanage"
)

type Selector struct {
	Version string
	Channel string
}

var semverPattern = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?(?:\+[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?$`)

func (s Selector) Validate() error {
	if strings.TrimSpace(s.Version) != "" && strings.TrimSpace(s.Channel) != "" {
		return fmt.Errorf("--version and --channel are mutually exclusive")
	}
	if s.Version != "" {
		version := strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(s.Version), "buda/"), "v")
		if !IsSemver(version) {
			return fmt.Errorf("invalid exact version %q", s.Version)
		}
	}
	if strings.ContainsAny(s.Channel, " \t\r\n/\\") {
		return fmt.Errorf("invalid release channel %q", s.Channel)
	}
	return nil
}

func IsSemver(value string) bool {
	match := semverPattern.FindStringSubmatch(strings.TrimPrefix(strings.TrimSpace(value), "v"))
	if match == nil {
		return false
	}
	for _, identifier := range strings.Split(match[4], ".") {
		if len(identifier) > 1 && identifier[0] == '0' {
			if _, err := strconv.ParseUint(identifier, 10, 64); err == nil {
				return false
			}
		}
	}
	return true
}

func Channel(version string) string {
	parts := strings.SplitN(version, "-", 2)
	if len(parts) == 1 {
		return "stable"
	}
	identifier := strings.SplitN(parts[1], ".", 2)[0]
	if identifier == "" {
		return "stable"
	}
	return identifier
}

type Result struct {
	Release  selfmanage.Release
	Binary   selfmanage.Asset
	Launcher selfmanage.Asset
	Manifest selfmanage.Asset
	Channel  string
}

var requiredReleaseAssets = []string{
	"buda-linux-amd64", "buda-linux-arm64", "buda-linux-armv7", "buda-linux-armv6",
	"buda-darwin-amd64", "buda-darwin-arm64", "buda-windows-amd64.exe", "buda-windows-arm64.exe",
	"buda-launcher-linux-amd64", "buda-launcher-linux-arm64", "buda-launcher-linux-armv7", "buda-launcher-linux-armv6",
	"buda-launcher-darwin-amd64", "buda-launcher-darwin-arm64", "buda-launcher-windows-amd64.exe", "buda-launcher-windows-arm64.exe",
	"buda.schema.json", "buda.global.schema.json",
	"buda.example.yaml", "buda.global.example.yaml",
	"guiho-s-0002-buda.zip", "guiho-i-buda.md", "guiho-p-buda.md",
	"artifacts.json", "checksums.txt",
}

func Select(ctx context.Context, catalog selfmanage.Catalog, selector Selector, buildTarget string) (Result, error) {
	if err := selector.Validate(); err != nil {
		return Result{}, err
	}
	assetName, err := selfmanage.AssetName(buildTarget)
	if err != nil {
		return Result{}, err
	}
	launcherName, err := selfmanage.LauncherAssetName(buildTarget)
	if err != nil {
		return Result{}, err
	}
	releases, err := catalog.Releases(ctx)
	if err != nil {
		return Result{}, err
	}
	requested := strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(selector.Version), "buda/"), "v")
	channel := strings.TrimSpace(selector.Channel)
	for _, release := range releases {
		if requested != "" && release.Version != requested {
			continue
		}
		if requested == "" {
			wanted := channel
			if wanted == "" {
				wanted = "stable"
			}
			if Channel(release.Version) != wanted {
				continue
			}
		}
		var binary, launcher, manifest selfmanage.Asset
		present := map[string]bool{}
		hasDuplicates := false
		for _, asset := range release.Assets {
			if strings.TrimSpace(asset.Name) == "" || strings.TrimSpace(asset.DownloadURL) == "" {
				continue
			}
			if present[asset.Name] {
				hasDuplicates = true
				break
			}
			present[asset.Name] = true
			if asset.Name == assetName {
				binary = asset
			}
			if asset.Name == launcherName {
				launcher = asset
			}
			if asset.Name == "checksums.txt" {
				manifest = asset
			}
		}
		if hasDuplicates {
			if requested != "" {
				return Result{}, fmt.Errorf("release %s contains duplicate asset names", release.Version)
			}
			continue
		}
		if binary.Name == "" || launcher.Name == "" || manifest.Name == "" {
			if requested != "" {
				return Result{}, fmt.Errorf("release %s lacks compatible payload, launcher, or checksums.txt", release.Version)
			}
			continue
		}
		if release.HTMLURL == "" || release.Draft {
			if requested != "" {
				return Result{}, fmt.Errorf("release %s is not a publishable GitHub release", release.Version)
			}
			continue
		}
		complete := true
		for _, name := range requiredReleaseAssets {
			if !present[name] {
				complete = false
				break
			}
		}
		if !complete {
			if requested != "" {
				return Result{}, fmt.Errorf("release %s is incomplete: required manifest, payload, launcher, schema, example, or agent artifacts are missing", release.Version)
			}
			continue
		}
		return Result{Release: release, Binary: binary, Launcher: launcher, Manifest: manifest, Channel: Channel(release.Version)}, nil
	}
	return Result{}, fmt.Errorf("no compatible complete Buda release found")
}
