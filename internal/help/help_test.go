package help

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestParseDepthUsesConventionGrammar(t *testing.T) {
	for _, value := range []string{"max", "2", "999"} {
		if _, err := ParseDepth(value); err != nil {
			t.Fatalf("ParseDepth(%q) failed: %v", value, err)
		}
	}
	for _, value := range []string{"", "0", "1", "-1", "two"} {
		if _, err := ParseDepth(value); err == nil {
			t.Fatalf("ParseDepth(%q) unexpectedly succeeded", value)
		}
	}
}

func TestTreeScopesGlobalFlagsAndDescendants(t *testing.T) {
	root := &cobra.Command{Use: "cliname", Short: "root"}
	root.PersistentFlags().String("wiki", "", "explicit wiki")
	child := &cobra.Command{Use: "child", Short: "child"}
	child.Flags().String("local", "", "local value")
	leaf := &cobra.Command{Use: "leaf", Short: "leaf"}
	child.AddCommand(leaf)
	root.AddCommand(child)
	without, err := Tree(root, "max", false)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(without, "--wiki") != 1 || !strings.Contains(without, "--local") {
		t.Fatalf("global flags were not concise or local flags were omitted:\n%s", without)
	}
	with, err := Tree(root, "max", true)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(with, "--wiki") != 3 {
		t.Fatalf("global flag was not repeated at every scope:\n%s", with)
	}
}
