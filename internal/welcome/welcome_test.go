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
		"Platform  Windows x64", "Version   v1.2.3", "Author    GUIHO (www.guiho.co)",
		"Run buda --help to see available commands.",
	} {
		if !strings.Contains(output, required) {
			t.Fatalf("welcome output omits %q:\n%s", required, output)
		}
	}
	if !strings.HasPrefix(output, "\n") || strings.HasPrefix(output, "\n\n") {
		t.Fatal("welcome output does not start with exactly one blank line")
	}
	if !strings.HasSuffix(output, "commands.\n\n") {
		t.Fatal("welcome output does not end with exactly one blank line")
	}
	// Should be borderless and deterministic
	for _, forbidden := range []string{"\x1b[", "╭", "╰", "Hello Windows - buda"} {
		if strings.Contains(output, forbidden) {
			t.Fatalf("welcome output contains %q:\n%s", forbidden, output)
		}
	}
	if strings.Count(output, "Platform") != 1 || strings.Count(output, "Version") != 1 || strings.Count(output, "Author") != 1 {
		t.Fatalf("welcome output has duplicate detail lines: %q", output)
	}
}

func TestRenderIsLeftAlignedWithoutCreatorLine(t *testing.T) {
	output := Render("linux", "amd64", "1.2.3")
	var expected strings.Builder
	expected.WriteString("\n")
	for _, line := range budaLogo {
		expected.WriteString(line + "\n")
	}
	expected.WriteString("AI-maintained OKF wiki\n")
	expected.WriteString("GUIHO \u00b7 Crist\u00f3v\u00e3o GUIHO\n")
	expected.WriteString("\n")
	expected.WriteString("Platform  Linux x64\n")
	expected.WriteString("Version   v1.2.3\n")
	expected.WriteString("Author    GUIHO (www.guiho.co)\n")
	expected.WriteString("\n")
	expected.WriteString("Run buda --help to see available commands.\n")
	expected.WriteString("\n")
	if output != expected.String() {
		t.Fatalf("welcome output mismatch:\n got: %q\nwant: %q", output, expected.String())
	}
	if strings.Contains(output, "Creator") {
		t.Fatalf("welcome output contains a Creator line:\n%s", output)
	}
	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(line, " ") {
			t.Fatalf("welcome line is not left-aligned: %q", line)
		}
	}
}

func TestRenderWithColorUsesCompleteBudaPalette(t *testing.T) {
	output := RenderWithColor("linux", "arm64", "v2.0.0", true)
	for _, color := range []string{ansiMist, ansiSky, ansiOcean, ansiOceanDark, ansiNavy} {
		if !strings.Contains(output, color) {
			t.Fatalf("colored welcome omits palette code %q", color)
		}
	}
	if !strings.Contains(output, ansiOceanDark+"Platform"+ansiReset+"  ") || !strings.Contains(output, ansiOceanDark+"Version "+ansiReset+"  ") || !strings.Contains(output, ansiOceanDark+"Author  "+ansiReset+"  ") {
		t.Fatalf("colored welcome detail labels are not left-aligned: %q", output)
	}
	authorValue := ansiMist + "GUIHO (" + ansiReset + ansiAccent + "www.guiho.co" + ansiReset + ansiMist + ")" + ansiReset
	if !strings.Contains(output, authorValue) {
		t.Fatalf("colored welcome does not paint the website in the accent colour: %q", output)
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
