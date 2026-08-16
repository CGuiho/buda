package releasecatalog

import "testing"

func TestIsSemverAndChannel(t *testing.T) {
	for _, value := range []string{"0.2.0", "1.2.3-canary.4", "1.2.3+build.7"} {
		if !IsSemver(value) {
			t.Fatalf("IsSemver(%q) = false", value)
		}
	}
	for _, value := range []string{"1.2", "01.2.3", "1.2.3-01", "1.2.3-"} {
		if IsSemver(value) {
			t.Fatalf("IsSemver(%q) = true", value)
		}
	}
	for value, want := range map[string]string{
		"1.2.3": "stable", "1.2.3-canary.4": "canary", "1.2.3-alpha.2": "alpha",
	} {
		if got := Channel(value); got != want {
			t.Fatalf("Channel(%q) = %q, want %q", value, got, want)
		}
	}
}
