package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/CGuiho/buda/internal/agent"
	"github.com/CGuiho/buda/internal/health"
	packdomain "github.com/CGuiho/buda/internal/pack"
	"github.com/CGuiho/buda/internal/qmd"
	"github.com/CGuiho/buda/internal/repository"
	"github.com/spf13/cobra"
)

type doctorCheck struct {
	State   string `json:"state"`
	Repair  string `json:"repair,omitempty"`
	Message string `json:"message,omitempty"`
	Detail  any    `json:"detail,omitempty"`
}

func NewDoctorCommand(deps Dependencies, factories ...QMDFactory) *cobra.Command {
	return &cobra.Command{
		Use:     "doctor",
		Short:   "Diagnose Buda repository and qmd readiness without repair.",
		Example: "  buda doctor --wiki ./wiki --json",
		Args:    NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			repo, client, err := openQMD(deps, factories)
			if err != nil {
				return err
			}
			checks := map[string]doctorCheck{
				"configuration": {State: "valid"},
				"repository":    {State: "resolved"},
			}
			failureCode := 0

			canonical, scanErr := health.Scan(repo.Bundle, repo.Config.WikiID, dependencyNow(deps))
			if scanErr != nil {
				checks["canonical"] = doctorCheck{State: "error", Repair: "buda lint --wiki <path>", Message: scanErr.Error()}
				failureCode = 1
			} else if !canonical.Conformant || !canonical.Healthy {
				checks["canonical"] = doctorCheck{State: "unhealthy", Repair: "buda lint --wiki <path>", Detail: canonical}
				failureCode = 1
			} else {
				checks["canonical"] = doctorCheck{State: "healthy", Detail: canonical}
			}

			globalSkills, globalErr := deps.Agents.SkillInstallations(false, "")
			localSkills, localErr := deps.Agents.SkillInstallations(true, repo.Root)
			instructions, instructionErr := deps.Agents.ListInstructions(repo.Root)
			if globalErr != nil || localErr != nil || instructionErr != nil {
				checks["agent_resources"] = doctorCheck{State: "error", Repair: "buda agent skill upgrade; buda agent instruction upgrade --wiki <path>", Message: firstError(globalErr, localErr, instructionErr)}
				failureCode = maxExit(failureCode, 1)
			} else {
				current := installationsCurrent(globalSkills) && optionalInstallationsCurrent(localSkills) && instructionsCurrent(instructions)
				state := "current"
				if !current {
					state = "stale"
					failureCode = maxExit(failureCode, 1)
				}
				checks["agent_resources"] = doctorCheck{State: state, Repair: "buda agent skill upgrade; buda agent skill upgrade --local --wiki <path>; buda agent instruction upgrade --wiki <path>", Detail: map[string]any{"global_skills": globalSkills, "local_skills": localSkills, "instructions": instructions}}
			}

			reproducible, packMessage := checkPackReproducibility(repo)
			if reproducible {
				checks["pack_reproducibility"] = doctorCheck{State: "reproducible"}
			} else {
				checks["pack_reproducibility"] = doctorCheck{State: "error", Repair: "buda pack --wiki <path> --output <archive.zip>", Message: packMessage}
				failureCode = maxExit(failureCode, 1)
			}

			diagnostic, qmdErr := client.Doctor(command.Context())
			if qmdErr != nil {
				checks["qmd"] = doctorCheck{State: "unavailable", Repair: "install a supported qmd, then run: buda index --wiki <path>", Message: qmdErr.Error(), Detail: qmdErrorDetail(qmdErr)}
				failureCode = 4
			} else {
				state := diagnostic.State
				if state == "" {
					state = "ready"
				}
				checks["qmd"] = doctorCheck{State: state, Detail: diagnostic}
				if state != "ready" {
					checks["qmd"] = doctorCheck{State: state, Repair: "follow qmd doctor recommendations, then rerun: buda doctor --wiki <path>", Detail: diagnostic}
					failureCode = maxExit(failureCode, 1)
				}
			}

			result := map[string]any{
				"command": "doctor", "wiki": repo.Root, "bundle": repo.Bundle,
				"qmd_project_directory": repo.QMDProject, "collection": repo.Collection,
				"checks": checks, "healthy": failureCode == 0, "repair_performed": false,
			}
			if JSONRequested(deps) {
				if err := WriteJSON(command, result); err != nil {
					return err
				}
				if failureCode != 0 {
					return &renderedError{cause: &codedError{code: failureCode, message: "doctor reported readiness failures"}}
				}
				return nil
			}
			fmt.Fprintf(command.OutOrStdout(), "wiki: %s\nbundle: %s\nqmd project: %s\ncollection: %s\n", repo.Root, repo.Bundle, repo.QMDProject, repo.Collection)
			for _, name := range []string{"configuration", "repository", "canonical", "agent_resources", "pack_reproducibility", "qmd"} {
				check := checks[name]
				fmt.Fprintf(command.OutOrStdout(), "%s: %s\n", name, check.State)
				if check.Repair != "" && check.State != "valid" && check.State != "resolved" && check.State != "healthy" && check.State != "current" && check.State != "reproducible" && check.State != "ready" {
					fmt.Fprintf(command.OutOrStdout(), "  repair: %s\n", check.Repair)
				}
			}
			fmt.Fprintln(command.OutOrStdout(), "repair performed: no")
			if failureCode != 0 {
				return &codedError{code: failureCode, message: "doctor reported readiness failures"}
			}
			return nil
		},
	}
}

func installationsCurrent(values []agent.SkillInstallation) bool {
	if len(values) == 0 {
		return false
	}
	for _, value := range values {
		if !value.Installed || !value.Current {
			return false
		}
	}
	return true
}

func optionalInstallationsCurrent(values []agent.SkillInstallation) bool {
	installed := 0
	for _, value := range values {
		if !value.Installed {
			continue
		}
		installed++
		if !value.Current {
			return false
		}
	}
	return installed == 0 || installed == len(values)
}

func instructionsCurrent(values []agent.InstructionTarget) bool {
	if len(values) == 0 {
		return false
	}
	for _, value := range values {
		if value.Malformed || !value.Managed || !value.Current {
			return false
		}
	}
	return true
}

func checkPackReproducibility(repo repository.Repository) (bool, string) {
	temporary, err := os.MkdirTemp("", "buda-doctor-pack-*")
	if err != nil {
		return false, err.Error()
	}
	defer os.RemoveAll(temporary)
	var digests []string
	for _, name := range []string{"first.zip", "second.zip"} {
		result, createErr := packdomain.Create(packdomain.Options{WikiRoot: repo.Root, BundleRoot: repo.Bundle, BundleName: repo.Config.Bundle, WikiID: repo.Config.WikiID, Output: filepath.Join(temporary, name)})
		if createErr != nil {
			return false, createErr.Error()
		}
		digests = append(digests, result.ArchiveSHA256)
	}
	if digests[0] != digests[1] {
		return false, "two packs of unchanged canonical state produced different digests"
	}
	return true, ""
}

func qmdErrorDetail(err error) any {
	var commandError *qmd.CommandError
	if errors.As(err, &commandError) {
		return commandError
	}
	return nil
}

func firstError(values ...error) string {
	for _, value := range values {
		if value != nil {
			return value.Error()
		}
	}
	return ""
}

func maxExit(left, right int) int {
	if right > left {
		return right
	}
	return left
}
