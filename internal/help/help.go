package help

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"github.com/spf13/pflag"
)

type treeItem struct {
	name        string
	description string
	command     *cobra.Command
	flag        bool
}

// Tree renders the live public subtree rooted at command.
func Tree(command *cobra.Command, depth int) string {
	var output strings.Builder
	output.WriteString("COMMAND TREE\n\n")
	fmt.Fprintf(&output, "%s  %s\n", command.CommandPath(), command.Short)
	renderChildren(&output, command, "", 1, depth)
	return output.String()
}

func renderChildren(output *strings.Builder, command *cobra.Command, prefix string, level, maximum int) {
	if maximum > 0 && level > maximum {
		return
	}
	items := commandTreeItems(command)
	for index, item := range items {
		last := index == len(items)-1
		branch, next := "├── ", prefix+"│   "
		if last {
			branch, next = "└── ", prefix+"    "
		}
		fmt.Fprintf(output, "%s%s%s  %s\n", prefix, branch, item.name, item.description)
		if item.command != nil {
			renderChildren(output, item.command, next, level+1, maximum)
		}
	}
}

func commandTreeItems(command *cobra.Command) []treeItem {
	var items []treeItem
	for _, child := range command.Commands() {
		if !child.Hidden {
			items = append(items, treeItem{name: child.Name(), description: child.Short, command: child})
		}
	}
	seen := map[string]bool{}
	add := func(set *pflag.FlagSet) {
		set.VisitAll(func(flag *pflag.Flag) {
			if flag.Hidden || seen[flag.Name] {
				return
			}
			seen[flag.Name] = true
			name := "--" + flag.Name
			if flag.Shorthand != "" {
				name = "-" + flag.Shorthand + ", " + name
			}
			if flag.NoOptDefVal == "" && flag.Value.Type() != "bool" {
				name += " <value>"
			}
			items = append(items, treeItem{name: name, description: flag.Usage, flag: true})
		})
	}
	add(command.NonInheritedFlags())
	add(command.InheritedFlags())
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].flag != items[j].flag {
			return !items[i].flag
		}
		return items[i].name < items[j].name
	})
	return items
}

// Markdown renders deterministic documentation for the live public subtree.
func Markdown(command *cobra.Command, depth int) (string, error) {
	var output bytes.Buffer
	var visit func(*cobra.Command, int) error
	visit = func(current *cobra.Command, level int) error {
		if depth > 0 && level > depth {
			return nil
		}
		current.DisableAutoGenTag = true
		var page bytes.Buffer
		if err := doc.GenMarkdown(current, &page); err != nil {
			return fmt.Errorf("generate Markdown help for %s: %w", current.CommandPath(), err)
		}
		if output.Len() > 0 {
			output.WriteString("\n---\n\n")
		}
		output.Write(page.Bytes())
		children := append([]*cobra.Command(nil), current.Commands()...)
		sort.Slice(children, func(i, j int) bool { return children[i].Name() < children[j].Name() })
		for _, child := range children {
			if !child.Hidden {
				if err := visit(child, level+1); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if err := visit(command, 0); err != nil {
		return "", err
	}
	return output.String(), nil
}
