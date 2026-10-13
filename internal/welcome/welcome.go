package welcome

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

const (
	ansiReset = "\x1b[0m"
	ansiBold  = "\x1b[1m"
	ansiDim   = "\x1b[2m"

	// Buda palette derived from user image: 2F6690, 3A7CA5, D9DCD6, 16425B, 81C3D7
	ansiNavy      = "\x1b[38;2;22;66;91m"    // 16425B deep navy
	ansiOceanDark = "\x1b[38;2;47;102;144m"  // 2F6690
	ansiOcean     = "\x1b[38;2;58;124;165m"  // 3A7CA5
	ansiMist      = "\x1b[38;2;217;220;214m" // D9DCD6 light
	ansiSky       = "\x1b[38;2;129;195;215m" // 81C3D7

	// ansiAccent is Buda's primary brand colour, used for the GUIHO
	// attribution and the website in the Author detail.
	ansiAccent = ansiOcean
)

var platformLabels = map[string]string{
	"darwin":  "macOS",
	"linux":   "Linux",
	"windows": "Windows",
	"win32":   "Windows",
}

var archLabels = map[string]string{
	"amd64": "x64",
	"386":   "x86",
	"arm64": "arm64",
	"arm":   "arm",
}

var budaLogo = []string{
	"██████╗ ██╗   ██╗██████╗  █████╗",
	"██╔══██╗██║   ██║██╔══██╗██╔══██╗",
	"██████╔╝██║   ██║██║  ██║███████║",
	"██╔══██╗██║   ██║██║  ██║██╔══██║",
	"██████╔╝╚██████╔╝██████╔╝██║  ██║",
	"╚═════╝  ╚═════╝ ╚═════╝ ╚═╝  ╚═╝",
}

var logoColors = []string{ansiMist, ansiSky, ansiOcean, ansiOceanDark, ansiNavy, ansiOcean}

// Render returns deterministic borderless welcome output without ANSI codes.
func Render(platform, architecture, version string) string {
	return render(platform, architecture, version, false)
}

// RenderWithColor returns the borderless welcome output with the Buda palette
// when withColor is true.
func RenderWithColor(platform, architecture, version string, withColor bool) string {
	return render(platform, architecture, version, withColor)
}

// ShouldUseColor reports whether ANSI output is appropriate for the current
// terminal. NO_COLOR and TERM=dumb always disable color.
func ShouldUseColor(isTerminal bool) bool {
	if os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" {
		return false
	}
	return isTerminal
}

func render(platform, architecture, version string, withColor bool) string {
	if platform == "" {
		platform = runtime.GOOS
	}
	if architecture == "" {
		architecture = runtime.GOARCH
	}
	platform = displayValue(platformLabels, platform)
	architecture = displayValue(archLabels, architecture)
	version = strings.TrimPrefix(strings.TrimSpace(version), "v")

	var output strings.Builder
	output.WriteString("\n")
	for index, line := range budaLogo {
		output.WriteString(colorize(withColor, line, logoColors[index]+ansiBold))
		output.WriteByte('\n')
	}
	output.WriteString(colorize(withColor, "AI-maintained OKF wiki", ansiSky+ansiDim))
	output.WriteByte('\n')
	output.WriteString(colorize(withColor, "GUIHO \u00b7 Crist\u00f3v\u00e3o GUIHO", ansiAccent))
	output.WriteString("\n\n")

	writeDetail(&output, withColor, "Platform", colorize(withColor, platform+" "+architecture, ansiMist))
	writeDetail(&output, withColor, "Version", colorize(withColor, "v"+version, ansiSky+ansiBold))
	writeDetail(&output, withColor, "Author",
		colorize(withColor, "GUIHO (", ansiMist)+
			colorize(withColor, "www.guiho.co", ansiAccent)+
			colorize(withColor, ")", ansiMist))
	output.WriteByte('\n')
	output.WriteString("Run ")
	output.WriteString(colorize(withColor, "buda --help", ansiMist+ansiBold))
	output.WriteString(" to see available commands.\n")
	output.WriteString("\n")
	return output.String()
}

func displayValue(labels map[string]string, value string) string {
	if label, ok := labels[value]; ok {
		return label
	}
	return value
}

func colorize(enabled bool, value, codes string) string {
	if !enabled {
		return value
	}
	return codes + value + ansiReset
}

// writeDetail writes one left-aligned detail line. value is already
// colorized by the caller when colour is enabled.
func writeDetail(output *strings.Builder, withColor bool, label, value string) {
	output.WriteString(colorize(withColor, fmt.Sprintf("%-8s", label), ansiOceanDark))
	output.WriteString("  ")
	output.WriteString(value)
	output.WriteByte('\n')
}
