package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/CGuiho/buda/internal/config"
	"github.com/CGuiho/buda/internal/repository"
	"github.com/spf13/cobra"
)

func NewInitCommand(deps Dependencies, factories ...QMDFactory) *cobra.Command {
	deps = normalizeDependencies(deps)
	var wikiID string
	command := &cobra.Command{
		Use: "init", Short: "Initialize or reconcile one explicit OKF-compatible Buda wiki.",
		Example: "  buda init --wiki ./wiki --wiki-id team-knowledge", Args: NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			version := releaseVersion(deps)
			input := bufio.NewReader(deps.In)
			root, err := repository.ResolveTarget(deps.Options.Wiki)
			if err != nil {
				return RepositoryError("resolve initialization target", err)
			}
			projectConfigPath := filepath.Join(root, config.FileName)
			if wikiID == "" {
				if existing, loadErr := config.LoadProject(projectConfigPath); loadErr == nil {
					wikiID = existing.WikiID
				} else if deps.Interactive() && !JSONRequested(deps) {
					fmt.Fprint(command.OutOrStdout(), "Immutable wiki id for this selected repository: ")
					answer, readErr := readPromptLine(input)
					if readErr != nil || answer == "" {
						return UsageError("a non-empty --wiki-id answer is required")
					}
					wikiID = answer
				} else {
					return UsageError("--wiki-id is required for a new wiki; an existing valid project configuration was not found")
				}
			}
			globalPath, global, globalPending, migrated, globalErr := resolveGlobalConfig(deps.HomeDir, version)
			if globalErr != nil {
				return MutationError("prepare global Buda configuration", globalErr)
			}
			project := config.ProjectConfig{Schema: config.CurrentSchema, WikiID: wikiID}
			if _, statErr := os.Stat(projectConfigPath); statErr == nil {
				project, err = config.LoadProject(projectConfigPath)
				if err != nil {
					return RepositoryError("load existing project configuration", err)
				}
			} else if !os.IsNotExist(statErr) {
				return RepositoryError("inspect existing project configuration", statErr)
			}
			policyPrompted, err := reconcileAgentPolicies(command.OutOrStdout(), input, deps.Interactive() && !JSONRequested(deps), &global, globalPending && !migrated)
			if err != nil {
				return err
			}
			configuration, err := config.Merge(global, project)
			if err != nil {
				return RepositoryError("merge Buda configuration", err)
			}
			candidate := repository.Repository{Root: root, ConfigPath: projectConfigPath,
				Bundle: filepath.Join(root, configuration.Bundle), QMDProject: filepath.Join(root, configuration.QMD.ProjectDirectory),
				Derived: filepath.Join(root, configuration.Derived), Collection: configuration.QMD.Collection, Config: configuration}
			client, err := qmdFactory(factories)(candidate)
			if err != nil {
				return RepositoryError("configure qmd adapter", err)
			}
			if _, err := client.CheckCompatibility(command.Context()); err != nil {
				return externalError("validate qmd compatibility before initialization", err)
			}
			if globalPending || policyPrompted {
				if err := persistGlobalConfiguration(globalPath, global, version); err != nil {
					return MutationError("persist global Buda configuration", err)
				}
			}
			result, err := repository.Initialize(root, repository.InitOptions{WikiID: wikiID, Configuration: configuration, Version: version, Now: dependencyNow(deps),
				BeforeCommit: func(candidate repository.Repository) error {
					if err := client.EnsureProject(command.Context()); err != nil {
						return externalError("validate qmd and establish project-local collection", err)
					}
					return nil
				}})
			if err != nil {
				if _, ok := err.(exitCoder); ok {
					return err
				}
				return MutationError("initialize selected wiki", err)
			}
			deps.Options.Wiki = result.Repository.Root
			skills, err := deps.Agents.InstallSkill(false, "")
			if err != nil {
				return MutationError("install embedded Buda skill globally", err)
			}
			instructions, err := deps.Agents.ApplyInstructions(result.Repository.Root)
			if err != nil {
				return MutationError("apply embedded Buda repository instructions", err)
			}
			if _, err := repository.Open(result.Repository.Root); err != nil {
				return MutationError("revalidate merged Buda project configuration", err)
			}
			if _, err := config.LoadGlobal(globalPath); err != nil {
				return MutationError("revalidate global Buda configuration", err)
			}
			output := map[string]any{"command": "init", "wiki": result.Repository.Root, "bundle": result.Repository.Bundle,
				"qmd_project_directory": result.Repository.QMDProject, "collection": result.Repository.Collection, "wiki_id": result.Repository.Config.WikiID,
				"created": result.Created, "unchanged": result.Unchanged, "skills": skills, "instructions": instructions,
				"project_config": projectConfigPath, "global_config": globalPath, "global_config_created": globalPending,
				"policy_prompted":           policyPrompted,
				"effective_agent_evolution": result.Repository.Config.EffectiveAgent()}
			if JSONRequested(deps) {
				return WriteJSON(command, output)
			}
			fmt.Fprintf(command.OutOrStdout(), "wiki: %s\nbundle: %s\nqmd project: %s\ncollection: %s\nwiki id: %s\nproject config: %s\nglobal config: %s\n",
				result.Repository.Root, result.Repository.Bundle, result.Repository.QMDProject, result.Repository.Collection, result.Repository.Config.WikiID, projectConfigPath, globalPath)
			if result.Unchanged {
				fmt.Fprintln(command.OutOrStdout(), "initialization: already current")
			} else {
				fmt.Fprintf(command.OutOrStdout(), "initialized: %d paths\n", len(result.Created))
			}
			fmt.Fprintln(command.OutOrStdout(), "qmd: compatible project-local collection established")
			fmt.Fprintln(command.OutOrStdout(), "agent resources: reconciled")
			return nil
		},
	}
	command.Flags().StringVar(&wikiID, "wiki-id", "", "Immutable user-assigned identifier for this wiki")
	return command
}

func resolveGlobalConfig(homeDir func() (string, error), version string) (string, config.GlobalConfig, bool, bool, error) {
	if homeDir == nil {
		homeDir = os.UserHomeDir
	}
	home, err := homeDir()
	if err != nil {
		return "", config.GlobalConfig{}, false, false, err
	}
	path := config.GlobalPath(home)
	if _, err := os.Stat(path); err == nil {
		global, err := config.LoadGlobal(path)
		if err != nil {
			return path, config.GlobalConfig{}, false, false, err
		}
		return path, global, false, false, nil
	} else if !os.IsNotExist(err) {
		return path, config.GlobalConfig{}, false, false, err
	}
	// One-time read-only discovery of the 0.1.x global buda.yaml. Strictly
	// validated before mapping; the legacy file is left untouched and the new
	// global file is only persisted after every interactive answer exists.
	legacyPath := filepath.Join(home, ".guiho", "buda", "buda.yaml")
	if _, err := os.Stat(legacyPath); err == nil {
		legacyConfig, loadErr := config.Load(legacyPath)
		if loadErr != nil {
			return path, config.GlobalConfig{}, false, false, fmt.Errorf("validate legacy global configuration: %w", loadErr)
		}
		global := config.GlobalConfig{
			Schema:  config.CurrentSchema,
			Bundle:  &legacyConfig.Bundle,
			Derived: &legacyConfig.Derived,
			QMD: &config.QMDConfig{
				Executable:       legacyConfig.QMD.Executable,
				Collection:       legacyConfig.QMD.Collection,
				ProjectDirectory: legacyConfig.QMD.ProjectDirectory,
			},
			Agent: &config.AgentConfig{
				Evolution: legacyConfig.EffectiveAgent().Evolution,
			},
		}
		if err := global.Validate(); err != nil {
			return path, config.GlobalConfig{}, false, false, err
		}
		return path, global, true, true, nil
	} else if !os.IsNotExist(err) {
		return path, config.GlobalConfig{}, false, false, err
	}
	return path, config.DefaultGlobal(), true, false, nil
}

func readPromptLine(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	line = strings.TrimSpace(line)
	if err != nil && line == "" {
		return "", err
	}
	return line, nil
}

type policyPrompt struct {
	name  string
	value *config.Policy
}

func reconcileAgentPolicies(out io.Writer, input *bufio.Reader, interactive bool, global *config.GlobalConfig, forcePrompt bool) (bool, error) {
	if global.Agent == nil {
		agent := config.AgentConfig{}
		global.Agent = &agent
	}
	policies := []policyPrompt{
		{name: "upgrades", value: &global.Agent.Evolution.Upgrade},
		{name: "bug issue creation", value: &global.Agent.Evolution.Issues.Bugs},
		{name: "improvement issue creation", value: &global.Agent.Evolution.Issues.Improvements},
		{name: "review issue creation", value: &global.Agent.Evolution.Issues.Reviews},
	}
	missing := make([]policyPrompt, 0, len(policies))
	for _, policy := range policies {
		if forcePrompt || strings.TrimSpace(string(*policy.value)) == "" {
			missing = append(missing, policy)
		}
	}
	if len(missing) == 0 {
		return false, nil
	}
	if !interactive {
		return false, UsageError("init requires an interactive terminal to configure agent.evolution policies; provide the answers interactively and retry")
	}
	fmt.Fprintln(out, "Buda agent-evolution policies govern upgrades and GitHub issue creation.")
	fmt.Fprintln(out, "disabled prohibits the action; always-ask requests approval; always-proceed authorizes it without another question.")
	fmt.Fprintln(out, "The recommended choice is always-proceed so the agent can keep Buda current and report useful findings.")
	fmt.Fprint(out, "Apply always-proceed to all four global policies? [y/N] ")
	allAnswer, _ := readPromptLine(input)
	if strings.EqualFold(allAnswer, "y") || strings.EqualFold(allAnswer, "yes") {
		for _, policy := range missing {
			*policy.value = config.PolicyAlwaysProceed
		}
		return true, nil
	}
	for _, policy := range missing {
		fmt.Fprintf(out, "Policy for %s [disabled|always-ask|always-proceed] (default always-ask): ", policy.name)
		answer, _ := readPromptLine(input)
		if answer == "" {
			*policy.value = config.PolicyAlwaysAsk
			continue
		}
		choice := config.Policy(answer)
		if choice != config.PolicyDisabled && choice != config.PolicyAlwaysAsk && choice != config.PolicyAlwaysProceed {
			return false, UsageError("invalid policy for %s: %q; choose disabled, always-ask, or always-proceed", policy.name, answer)
		}
		*policy.value = choice
	}
	return true, nil
}

func persistGlobalConfiguration(path string, value config.GlobalConfig, version string) error {
	data, err := config.MarshalGlobal(value, version)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	temp, err := os.CreateTemp(filepath.Dir(path), ".buda-policy-*")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)
	if _, err := temp.Write(data); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Sync(); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	return os.Rename(tempPath, path)
}
