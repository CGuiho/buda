package help

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
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
// ParseDepth validates the public help-tree-depth grammar. The convention
// intentionally exposes the word max rather than leaking the renderer's
// internal zero value to users.
func ParseDepth(value string) (int, error) {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "max" {
		return 0, nil
	}
	depth, err := strconv.Atoi(value)
	if err != nil || depth <= 1 {
		return 0, fmt.Errorf("help-tree-depth must be max or an integer greater than 1")
	}
	return depth, nil
}

// Tree renders the live public subtree rooted at command. Global persistent
// flags are shown once at the root of the rendered scope unless repeatGlobals
// is true, in which case they are included at every applicable descendant.
func Tree(command *cobra.Command, depth string, repeatGlobals bool) (string, error) {
	maximum, err := ParseDepth(depth)
	if err != nil {
		return "", err
	}
	var output strings.Builder
	output.WriteString("COMMAND TREE\n\n")
	fmt.Fprintf(&output, "%s  %s\n", commandUsage(command), command.Short)
	renderChildren(&output, command, "", 1, maximum, true, repeatGlobals)
	return output.String(), nil
}

func renderChildren(output *strings.Builder, command *cobra.Command, prefix string, level, maximum int, initial, repeatGlobals bool) {
	if maximum > 0 && level > maximum {
		return
	}
	items := commandTreeItems(command, initial || repeatGlobals)
	for index, item := range items {
		last := index == len(items)-1
		branch, next := "├── ", prefix+"│   "
		if last {
			branch, next = "└── ", prefix+"    "
		}
		fmt.Fprintf(output, "%s%s%s  %s\n", prefix, branch, item.name, item.description)
		if item.command != nil {
			renderChildren(output, item.command, next, level+1, maximum, false, repeatGlobals)
		}
	}
}

func commandTreeItems(command *cobra.Command, includeInherited bool) []treeItem {
	command.InitDefaultHelpFlag()
	var items []treeItem
	for _, child := range command.Commands() {
		if !child.Hidden {
			items = append(items, treeItem{name: child.Use, description: child.Short, command: child})
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
	if includeInherited {
		add(command.InheritedFlags())
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].flag != items[j].flag {
			return !items[i].flag
		}
		return items[i].name < items[j].name
	})
	return items
}

func commandUsage(command *cobra.Command) string {
	suffix := strings.TrimSpace(strings.TrimPrefix(command.Use, command.Name()))
	if suffix == "" {
		return command.CommandPath()
	}
	return command.CommandPath() + " " + suffix
}

// Markdown renders deterministic documentation for only the invoked command.
// Cobra's generated page already includes that command's flags and examples;
// recursively concatenating descendants made --help-docs ambiguous.
func Markdown(command *cobra.Command) (string, error) {
	command.DisableAutoGenTag = true
	var page bytes.Buffer
	if err := doc.GenMarkdown(command, &page); err != nil {
		return "", fmt.Errorf("generate Markdown help for %s: %w", command.CommandPath(), err)
	}
	return page.String(), nil
}
