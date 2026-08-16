// Package installlayout defines Buda's platform-native, shared GUIHO home.
// All lifecycle operations use this package so installers, upgrade, launcher,
// and uninstall cannot silently drift to different locations.
package installlayout

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type Layout struct {
	Home              string `json:"home"`
	GUIHO             string `json:"guiho"`
	CLIHome           string `json:"cli_home"`
	Bin               string `json:"bin"`
	Temp              string `json:"temp"`
	Versions          string `json:"versions"`
	State             string `json:"state"`
	Instances         string `json:"instances"`
	Current           string `json:"current"`
	InstalledManifest string `json:"installed_manifest"`
	GlobalConfig      string `json:"global_config"`
	Launcher          string `json:"launcher"`
}

func Current() (Layout, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Layout{}, fmt.Errorf("resolve user home: %w", err)
	}
	return ForHome(home)
}

func ForHome(home string) (Layout, error) {
	if strings.TrimSpace(home) == "" {
		return Layout{}, errors.New("home directory must not be empty")
	}
	absolute, err := filepath.Abs(home)
	if err != nil {
		return Layout{}, fmt.Errorf("resolve home directory: %w", err)
	}
	abs := filepath.Clean(absolute)
	guiho := filepath.Join(abs, ".guiho")
	cli := filepath.Join(guiho, "buda")
	bin := filepath.Join(guiho, "bin")
	launcher := filepath.Join(bin, "buda")
	if runtime.GOOS == "windows" {
		launcher += ".exe"
	}
	return Layout{
		Home: abs, GUIHO: guiho, CLIHome: cli, Bin: bin, Temp: filepath.Join(guiho, ".temp"),
		Versions: filepath.Join(cli, "versions"), State: filepath.Join(cli, "state"),
		Instances: filepath.Join(cli, "state", "instances"), Current: filepath.Join(cli, "current.json"),
		InstalledManifest: filepath.Join(cli, "installed-artifacts.json"),
		GlobalConfig:      filepath.Join(cli, "buda.global.yaml"), Launcher: launcher,
	}, nil
}

func (l Layout) Validate() error {
	paths := map[string]string{"guiho": l.GUIHO, "cli home": l.CLIHome, "bin": l.Bin, "temp": l.Temp, "versions": l.Versions, "state": l.State}
	for name, path := range paths {
		if filepath.IsAbs(path) == false || strings.TrimSpace(path) == "" {
			return fmt.Errorf("%s path must be absolute", name)
		}
	}
	if !within(l.GUIHO, l.CLIHome) || !within(l.GUIHO, l.Bin) || !within(l.GUIHO, l.Temp) {
		return errors.New("CLI paths must remain beneath shared .guiho")
	}
	if !within(l.CLIHome, l.Versions) || !within(l.CLIHome, l.State) {
		return errors.New("Buda state paths must remain beneath CLI home")
	}
	return nil
}

func within(parent, child string) bool {
	rel, err := filepath.Rel(parent, child)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return false
	}
	return true
}

func (l Layout) Operation(prefix string) (string, error) {
	if err := l.Validate(); err != nil {
		return "", err
	}
	prefix = strings.TrimSpace(prefix)
	if prefix == "" || strings.ContainsAny(prefix, `\/:*?"<>|`) {
		return "", errors.New("operation prefix is invalid")
	}
	if err := os.MkdirAll(l.Temp, 0o755); err != nil {
		return "", fmt.Errorf("create GUIHO temp directory: %w", err)
	}
	dir, err := os.MkdirTemp(l.Temp, "buda-"+prefix+"-")
	if err != nil {
		return "", fmt.Errorf("create operation staging directory: %w", err)
	}
	return dir, nil
}
