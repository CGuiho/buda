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

	innerWidth := 56
	for _, line := range budaLogo {
		if width := len([]rune(line)); width > innerWidth {
			innerWidth = width
		}
	}
	innerWidth += 4

	var output strings.Builder
	output.WriteString("\n\n")
	for index, line := range budaLogo {
		output.WriteString(colorize(withColor, center(line, innerWidth), logoColors[index]+ansiBold))
		output.WriteByte('\n')
	}
	output.WriteString(colorize(withColor, center("AI-maintained OKF wiki", innerWidth), ansiSky+ansiDim))
	output.WriteByte('\n')
	output.WriteString(colorize(withColor, center("GUIHO \u00b7 Crist\u00f3v\u00e3o GUIHO", innerWidth), ansiOcean))
	output.WriteString("\n\n\n")

	writeDetail(&output, withColor, "platform", platform+" "+architecture, ansiMist)
	writeDetail(&output, withColor, "version", "v"+version, ansiSky+ansiBold)
	output.WriteByte('\n')
	output.WriteString("Run ")
	output.WriteString(colorize(withColor, "buda --help", ansiMist+ansiBold))
	output.WriteString(" to see available commands.\n")
	output.WriteString("\n\n")
	return output.String()
}

func displayValue(labels map[string]string, value string) string {
	if label, ok := labels[value]; ok {
		return label
	}
	return value
}

func center(value string, width int) string {
	padding := width - len([]rune(value))
	if padding <= 0 {
		return value
	}
	return strings.Repeat(" ", padding/2) + value
}

func colorize(enabled bool, value, codes string) string {
	if !enabled {
		return value
	}
	return codes + value + ansiReset
}

func writeDetail(output *strings.Builder, withColor bool, label, value, valueColor string) {
	if withColor {
		output.WriteString(colorize(true, fmt.Sprintf("  %-9s", label), ansiOceanDark))
		output.WriteString("  ")
		output.WriteString(colorize(true, value, valueColor))
		output.WriteByte('\n')
		return
	}
	fmt.Fprintf(output, "  %-9s  %s\n", label, value)
}
