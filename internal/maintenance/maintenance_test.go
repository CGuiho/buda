package maintenance

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestShouldScheduleOnlyNormalRepositoryCommands(t *testing.T) {
	root := &cobra.Command{Use: "buda"}
	repository := &cobra.Command{Use: "lint"}
	agent := &cobra.Command{Use: "agent"}
	agentChild := &cobra.Command{Use: "list"}
	hidden := &cobra.Command{Use: WorkerCommand, Hidden: true}
	uninstall := &cobra.Command{Use: "uninstall"}
	root.AddCommand(repository, agent, hidden, uninstall)
	agent.AddCommand(agentChild)
	for name, test := range map[string]struct {
		command *cobra.Command
		want    bool
	}{
		"root": {root, false}, "repository": {repository, true}, "agent": {agent, false},
		"agent child": {agentChild, false}, "hidden": {hidden, false}, "uninstall": {uninstall, false},
	} {
		t.Run(name, func(t *testing.T) {
			if got := ShouldSchedule(test.command); got != test.want {
				t.Fatalf("ShouldSchedule() = %v, want %v", got, test.want)
			}
		})
	}
}
