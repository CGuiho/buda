package cmd

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"github.com/CGuiho/buda/internal/config"
)

func TestReconcileAgentPoliciesOffersRecommendedAllProceedChoice(t *testing.T) {
	global := config.DefaultGlobal()
	var output bytes.Buffer
	changed, err := reconcileAgentPolicies(&output, bufio.NewReader(strings.NewReader("yes\n")), true, &global, true)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("policy reconciliation did not report a change")
	}
	if global.Agent == nil {
		t.Fatal("global agent policy was not created")
	}
	policies := []config.Policy{
		global.Agent.Evolution.Upgrade,
		global.Agent.Evolution.Issues.Bugs,
		global.Agent.Evolution.Issues.Improvements,
		global.Agent.Evolution.Issues.Reviews,
	}
	for _, policy := range policies {
		if policy != config.PolicyAlwaysProceed {
			t.Fatalf("policy = %q, want always-proceed", policy)
		}
	}
	if !strings.Contains(output.String(), "disabled prohibits") || !strings.Contains(output.String(), "recommended choice is always-proceed") {
		t.Fatalf("policy explanation omitted: %q", output.String())
	}
}

func TestReconcileAgentPoliciesPreservesValidAndDefaultsSkipped(t *testing.T) {
	global := config.DefaultGlobal()
	global.Agent.Evolution.Upgrade = config.PolicyDisabled
	global.Agent.Evolution.Issues.Bugs = ""
	global.Agent.Evolution.Issues.Improvements = ""
	global.Agent.Evolution.Issues.Reviews = ""
	var output bytes.Buffer
	answers := "n\nalways-proceed\n\n\n"
	changed, err := reconcileAgentPolicies(&output, bufio.NewReader(strings.NewReader(answers)), true, &global, false)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("policy reconciliation did not report a change")
	}
	if global.Agent.Evolution.Upgrade != config.PolicyDisabled {
		t.Fatalf("existing upgrade policy changed to %q", global.Agent.Evolution.Upgrade)
	}
	if global.Agent.Evolution.Issues.Bugs != config.PolicyAlwaysProceed {
		t.Fatalf("bug policy = %q", global.Agent.Evolution.Issues.Bugs)
	}
	if global.Agent.Evolution.Issues.Improvements != config.PolicyAlwaysAsk || global.Agent.Evolution.Issues.Reviews != config.PolicyAlwaysAsk {
		t.Fatalf("skipped policies = %#v", global.Agent.Evolution.Issues)
	}
}

func TestReconcileAgentPoliciesFailsClosedWithoutInteractiveTerminal(t *testing.T) {
	global := config.DefaultGlobal()
	_, err := reconcileAgentPolicies(&bytes.Buffer{}, bufio.NewReader(strings.NewReader("yes\n")), false, &global, true)
	if ExitCode(err) != 2 || !strings.Contains(err.Error(), "interactive terminal") {
		t.Fatalf("error = %v, code = %d", err, ExitCode(err))
	}
}
