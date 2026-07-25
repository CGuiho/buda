package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/CGuiho/buda/internal/agent"
	"github.com/CGuiho/buda/internal/maintenance"
	"github.com/spf13/cobra"
)

func newAgentCommand(deps Dependencies) *cobra.Command {
	command := &cobra.Command{
		Use:   "agent",
		Short: "Manage Buda agent skills, instructions, and prompts.",
		Args:  NoArgs,
		RunE:  showHelp,
	}
	command.AddCommand(newAgentSkillCommand(deps))
	command.AddCommand(newAgentInstructionCommand(deps))
	command.AddCommand(newAgentPromptCommand(deps))
	return command
}

func newAgentSkillCommand(deps Dependencies) *cobra.Command {
	var local bool
	command := &cobra.Command{
		Use:   "skill",
		Short: "Manage the embedded guiho-s-0002-buda skill.",
		Args:  NoArgs,
		RunE:  showHelp,
	}
	command.PersistentFlags().BoolVar(&local, "local", false, "Use both project-local skill destinations under --wiki")
	for _, action := range []string{"install", "uninstall", "update"} {
		action := action
		command.AddCommand(&cobra.Command{
			Use:   action,
			Short: skillActionDescription(action),
			Args:  NoArgs,
			RunE: func(command *cobra.Command, _ []string) error {
				var (
					result []agent.SkillInstallation
					err    error
				)
				if action == "uninstall" {
					result, err = deps.Agents.UninstallSkill(local, deps.Options.Wiki)
				} else {
					result, err = deps.Agents.InstallSkill(local, deps.Options.Wiki)
				}
				if err != nil {
					return err
				}
				return writeAgentValue(command, deps, map[string]any{"action": action, "installations": result})
			},
		})
	}
	command.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List the embedded skill and both resolved installation states.",
		Args:  NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			skill, err := deps.Agents.Skill()
			if err != nil {
				return err
			}
			installed, err := deps.Agents.SkillInstallations(local, deps.Options.Wiki)
			if err != nil {
				return err
			}
			return writeAgentValue(command, deps, map[string]any{"embedded": skill, "installations": installed})
		},
	})
	command.AddCommand(&cobra.Command{
		Use:   "show [id]",
		Short: "Show embedded skill metadata and both installation states.",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) > 1 {
				return UsageError("accepts at most 1 arg(s), received %d", len(args))
			}
			if len(args) == 1 && args[0] != agent.SkillID {
				return UsageError("unknown skill id %q", args[0])
			}
			return nil
		},
		RunE: func(command *cobra.Command, _ []string) error {
			skill, err := deps.Agents.Skill()
			if err != nil {
				return err
			}
			installed, err := deps.Agents.SkillInstallations(local, deps.Options.Wiki)
			if err != nil {
				return err
			}
			return writeAgentValue(command, deps, map[string]any{"embedded": skill, "installations": installed})
		},
	})
	return command
}

func newAgentInstructionCommand(deps Dependencies) *cobra.Command {
	command := &cobra.Command{
		Use:   "instruction",
		Short: "Manage bounded Buda blocks in repository instructions.",
		Args:  NoArgs,
		RunE:  showHelp,
	}
	for _, action := range []string{"apply", "remove", "update"} {
		action := action
		command.AddCommand(&cobra.Command{
			Use:   action,
			Short: instructionActionDescription(action),
			Args:  NoArgs,
			RunE: func(command *cobra.Command, _ []string) error {
				var (
					result []agent.InstructionTarget
					err    error
				)
				if action == "remove" {
					result, err = deps.Agents.RemoveInstructions(deps.Options.Wiki)
				} else {
					result, err = deps.Agents.ApplyInstructions(deps.Options.Wiki)
				}
				if err != nil {
					return err
				}
				return writeAgentValue(command, deps, map[string]any{"action": action, "targets": result})
			},
		})
	}
	command.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List the embedded instruction and managed target states.",
		Args:  NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			context, err := deps.Agents.ReadInstructionContext(deps.Options.Wiki)
			if err != nil {
				return err
			}
			template, err := deps.Agents.InstructionTemplate(context)
			if err != nil {
				return err
			}
			targets, err := deps.Agents.ListInstructions(deps.Options.Wiki)
			if err != nil {
				return err
			}
			return writeAgentValue(command, deps, map[string]any{
				"embedded": map[string]string{"id": "buda", "body": template}, "targets": targets,
			})
		},
	})
	command.AddCommand(&cobra.Command{
		Use:   "show [AGENTS.md|CLAUDE.md]",
		Short: "Show the embedded block or one selected installed block.",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) > 1 {
				return UsageError("accepts at most 1 arg(s), received %d", len(args))
			}
			return nil
		},
		RunE: func(command *cobra.Command, args []string) error {
			var body string
			var err error
			if len(args) == 1 {
				body, err = deps.Agents.InstalledInstruction(deps.Options.Wiki, args[0])
			} else {
				context, contextErr := deps.Agents.ReadInstructionContext(deps.Options.Wiki)
				if contextErr != nil {
					return contextErr
				}
				body, err = deps.Agents.InstructionTemplate(context)
				if err == nil {
					body = agent.InstructionBegin + "\n" + body + "\n" + agent.InstructionEnd
				}
			}
			if err != nil {
				return err
			}
			if JSONRequested(deps) {
				return WriteJSON(command, map[string]string{"instruction": body})
			}
			fmt.Fprintln(command.OutOrStdout(), body)
			return nil
		},
	})
	return command
}

func newAgentPromptCommand(deps Dependencies) *cobra.Command {
	command := &cobra.Command{
		Use:   "prompt",
		Short: "Inspect embedded Buda prompt resources without running a model.",
		Args:  NoArgs,
		RunE:  showHelp,
	}
	command.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List embedded Buda prompts.",
		Args:  NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			prompt, err := deps.Agents.Prompt()
			if err != nil {
				return err
			}
			prompt.Body = ""
			return writeAgentValue(command, deps, []agent.Prompt{prompt})
		},
	})
	command.AddCommand(&cobra.Command{
		Use:   "show <id>",
		Short: "Print one raw embedded Buda prompt.",
		Args:  ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			if args[0] != agent.PromptID {
				return UsageError("unknown prompt id %q", args[0])
			}
			prompt, err := deps.Agents.Prompt()
			if err != nil {
				return err
			}
			if JSONRequested(deps) {
				return WriteJSON(command, prompt)
			}
			fmt.Fprintln(command.OutOrStdout(), prompt.Body)
			return nil
		},
	})
	return command
}

func newMaintenanceCommand(deps Dependencies) *cobra.Command {
	command := &cobra.Command{
		Use:    maintenance.WorkerCommand,
		Hidden: true,
		Args:   NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			wiki := strings.TrimSpace(deps.Options.Wiki)
			if wiki == "" {
				return RepositoryError("hidden agent maintenance requires --wiki", nil)
			}
			return maintenance.Run(deps.Agents, wiki)
		},
	}
	return command
}

func showHelp(command *cobra.Command, _ []string) error { return command.Help() }

func writeAgentValue(command *cobra.Command, deps Dependencies, value any) error {
	if JSONRequested(deps) {
		return WriteJSON(command, value)
	}
	text, err := formatHumanValue(value)
	if err != nil {
		return fmt.Errorf("format agent output: %w", err)
	}
	fmt.Fprint(command.OutOrStdout(), text)
	return nil
}

func formatHumanValue(value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var normalized any
	if err := decoder.Decode(&normalized); err != nil {
		return "", err
	}
	var output strings.Builder
	writeHumanValue(&output, normalized, "")
	return output.String(), nil
}

func writeHumanValue(output *strings.Builder, value any, indent string) {
	switch typed := value.(type) {
	case map[string]any:
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			child := typed[key]
			if isHumanScalar(child) {
				fmt.Fprintf(output, "%s%s: %v\n", indent, key, child)
			} else {
				fmt.Fprintf(output, "%s%s:\n", indent, key)
				writeHumanValue(output, child, indent+"  ")
			}
		}
	case []any:
		if len(typed) == 0 {
			fmt.Fprintf(output, "%s(none)\n", indent)
			return
		}
		for _, child := range typed {
			if isHumanScalar(child) {
				fmt.Fprintf(output, "%s- %v\n", indent, child)
			} else {
				fmt.Fprintf(output, "%s-\n", indent)
				writeHumanValue(output, child, indent+"  ")
			}
		}
	default:
		fmt.Fprintf(output, "%s%v\n", indent, typed)
	}
}

func isHumanScalar(value any) bool {
	if value == nil {
		return true
	}
	kind := reflect.TypeOf(value).Kind()
	return kind != reflect.Map && kind != reflect.Slice && kind != reflect.Array
}

func skillActionDescription(action string) string {
	switch action {
	case "install":
		return "Install the embedded skill transactionally in both destinations."
	case "uninstall":
		return "Remove only the Buda-owned skill from both destinations."
	default:
		return "Reconcile both destinations with the embedded skill."
	}
}

func instructionActionDescription(action string) string {
	switch action {
	case "apply":
		return "Apply one canonical Buda block to every resolved target."
	case "remove":
		return "Remove only complete bounded Buda blocks."
	default:
		return "Refresh and deduplicate complete bounded Buda blocks."
	}
}
