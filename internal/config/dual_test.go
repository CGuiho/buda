package config

import (
	"strings"
	"testing"
)

func TestDualConfigStrictMergeAndPolicyInheritance(t *testing.T) {
	global, err := DecodeGlobal(strings.NewReader(`schema: 1
bundle: shared
agent:
  evolution:
    upgrade: always-proceed
    issues:
      bugs: disabled
`))
	if err != nil {
		t.Fatal(err)
	}
	project, err := DecodeProject(strings.NewReader(`schema: 1
wiki_id: selected
qmd:
  collection: project-collection
agent:
  evolution:
    issues:
      reviews: always-proceed
`))
	if err != nil {
		t.Fatal(err)
	}
	effective, err := Merge(global, project)
	if err != nil {
		t.Fatal(err)
	}
	if effective.Bundle != "shared" || effective.QMD.Collection != "project-collection" || effective.WikiID != "selected" {
		t.Fatalf("effective config = %#v", effective)
	}
	policy := effective.EffectiveAgent().Evolution
	if policy.Upgrade != PolicyAlwaysProceed || policy.Issues.Bugs != PolicyDisabled || policy.Issues.Reviews != PolicyAlwaysProceed || policy.Issues.Improvements != PolicyAlwaysAsk {
		t.Fatalf("effective policy = %#v", policy)
	}
	if _, err := DecodeGlobal(strings.NewReader("schema: 1\nunknown: true\n")); err == nil {
		t.Fatal("unknown global field was accepted")
	}
}
