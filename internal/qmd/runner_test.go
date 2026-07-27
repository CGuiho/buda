package qmd

import (
	"strings"
	"testing"
)

func TestEnvironmentWithPWDReplacesInheritedVariants(t *testing.T) {
	directory := `C:\selected\wiki`
	environment := environmentWithPWD([]string{
		"Path=C:\\tools",
		"PWD=C:\\caller",
		"Pwd=C:\\another",
		"pwd=C:\\last",
		"BUDA_TEST=value",
	}, directory)

	count := 0
	for _, entry := range environment {
		key, value, _ := strings.Cut(entry, "=")
		if strings.EqualFold(key, "PWD") {
			count++
			if value != directory {
				t.Fatalf("PWD = %q, want %q", value, directory)
			}
		}
	}
	if count != 1 {
		t.Fatalf("effective PWD entries = %d, environment = %#v", count, environment)
	}
}
