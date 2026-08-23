package welcome

import (
	"strings"
	"testing"
)

func TestRenderBorderlessHelloWindow(t *testing.T) {
	output := Render("windows", "amd64", "1.2.3")
	for _, required := range []string{
		"██████╗ ██╗   ██╗██████╗", "AI-maintained OKF wiki",
		"GUIHO \u00b7 Crist\u00f3v\u00e3o GUIHO",
		"platform   Windows x64", "version    v1.2.3",
		"Run buda --help to see available commands.",
	} {
		if !strings.Contains(output, required) {
			t.Fatalf("welcome output omits %q:\n%s", required, output)
		}
	}
	if !strings.HasPrefix(output, "\n\n") {
		t.Fatal("welcome output does not start with two blank lines")
	}
	if !strings.HasSuffix(output, "\n\n") {
		t.Fatal("welcome output does not end with two blank lines")
	}
	// Should be borderless and deterministic
	for _, forbidden := range []string{"\x1b[", "╭", "╰", "Hello Windows - buda"} {
		if strings.Contains(output, forbidden) {
			t.Fatalf("welcome output contains %q:\n%s", forbidden, output)
		}
	}
	if strings.Count(output, "platform") != 1 || strings.Count(output, "version") != 1 {
		t.Fatalf("welcome output has duplicate detail lines: %q", output)
	}
}

func TestRenderWithColorUsesCompleteBudaPalette(t *testing.T) {
	output := RenderWithColor("linux", "arm64", "v2.0.0", true)
	for _, color := range []string{ansiMist, ansiSky, ansiOcean, ansiOceanDark, ansiNavy} {
		if !strings.Contains(output, color) {
			t.Fatalf("colored welcome omits palette code %q", color)
		}
	}
	if !strings.Contains(output, "Linux arm64") || !strings.Contains(output, "v2.0.0") {
		t.Fatalf("colored welcome omits normalized runtime details: %q", output)
	}
	// Ensure plain render has no ANSI
	plain := Render("linux", "arm64", "2.0.0")
	if strings.Contains(plain, "\x1b[") {
		t.Fatalf("plain welcome contains ANSI: %q", plain)
	}
}

func TestShouldUseColorHonorsEnvironment(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	t.Setenv("TERM", "xterm-256color")
	if !ShouldUseColor(true) || ShouldUseColor(false) {
		t.Fatal("terminal color detection is incorrect")
	}
	t.Setenv("NO_COLOR", "1")
	if ShouldUseColor(true) {
		t.Fatal("NO_COLOR did not disable color")
	}
	t.Setenv("NO_COLOR", "")
	t.Setenv("TERM", "dumb")
	if ShouldUseColor(true) {
		t.Fatal("TERM=dumb did not disable color")
	}
}

func TestRenderIsDeterministicAcrossPlatforms(t *testing.T) {
	a := Render("linux", "amd64", "0.2.0")
	b := Render("linux", "amd64", "0.2.0")
	if a != b {
		t.Fatalf("render not deterministic")
	}
	c := Render("darwin", "arm64", "0.2.0")
	if a == c {
		t.Fatalf("different platforms should produce different output")
	}
}
