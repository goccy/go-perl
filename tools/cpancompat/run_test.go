package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// blockingRunner returns a runner where the distribution "Dependent" needs
// "Configure" to be configured and "Runtime" to run (by its static
// requirements), with the given ones of those finished and not installed.
func blockingRunner(t *testing.T, notInstalled ...string) (*runner, *node) {
	t.Helper()
	results, err := os.Create(filepath.Join(t.TempDir(), "results.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { results.Close() })
	r := &runner{
		nodes:   map[string]*node{},
		waiters: map[string][]*node{},
		results: results,
	}
	for _, name := range []string{"Configure", "Runtime"} {
		r.nodes[name] = &node{d: &dist{Dist: name}}
	}
	for _, name := range notInstalled {
		r.nodes[name].res = &result{Dist: name, Outcome: outcomeConfigureFailed}
	}
	d := &dist{
		Dist: "Dependent",
		Requires: []requirement{
			{Module: "Configure::Module", Phase: "configure", Relationship: "requires", Dist: "Configure"},
			{Module: "Runtime::Module", Phase: "runtime", Relationship: "requires", Dist: "Runtime"},
		},
		Needs: []string{"Configure", "Runtime"},
	}
	n := &node{d: d, waiting: map[string]bool{}}
	r.nodes[d.Dist] = n
	return r, n
}

func TestReleaseBlocksOnlyOnConfigureRequirements(t *testing.T) {
	// What it needs to run may not be what its configuration asks for:
	// it is attempted.
	r, n := blockingRunner(t, "Runtime")
	if err := r.release(n); err != nil {
		t.Fatal(err)
	}
	if n.res != nil || r.ready.Len() != 1 {
		t.Errorf("with a runtime requirement not installed: settled %+v, ready %d; want it attempted", n.res, r.ready.Len())
	}

	r, n = blockingRunner(t, "Configure", "Runtime")
	if err := r.release(n); err != nil {
		t.Fatal(err)
	}
	if n.res == nil || n.res.Outcome != outcomeBlocked || !reflect.DeepEqual(n.res.BlockedBy, []string{"Configure"}) {
		t.Errorf("with a configure requirement not installed: settled %+v; want blocked by Configure", n.res)
	}
	if r.ready.Len() != 0 {
		t.Errorf("with a configure requirement not installed: %d ready; want none", r.ready.Len())
	}
}

func TestSettlePutsConfigureFailureDownToMissingNeeds(t *testing.T) {
	r, n := blockingRunner(t, "Runtime")
	res := &result{Dist: "Dependent", Outcome: outcomeConfigureFailed}
	if err := r.settle(n, res); err != nil {
		t.Fatal(err)
	}
	if res.Outcome != outcomeBlocked || !reflect.DeepEqual(res.BlockedBy, []string{"Runtime"}) {
		t.Errorf("configure failed with Runtime not installed: %s by %v; want blocked by Runtime", res.Outcome, res.BlockedBy)
	}

	r, n = blockingRunner(t)
	res = &result{Dist: "Dependent", Outcome: outcomeConfigureFailed}
	if err := r.settle(n, res); err != nil {
		t.Fatal(err)
	}
	if res.Outcome != outcomeConfigureFailed || res.BlockedBy != nil {
		t.Errorf("configure failed with everything installed: %s by %v; want configure_failed", res.Outcome, res.BlockedBy)
	}

	// A distribution that passes despite a static requirement not being
	// installed did not need it.
	r, n = blockingRunner(t, "Runtime")
	res = &result{Dist: "Dependent", Outcome: outcomePass, Installed: true}
	if err := r.settle(n, res); err != nil {
		t.Fatal(err)
	}
	if res.Outcome != outcomePass {
		t.Errorf("passed with Runtime not installed: %s; want pass", res.Outcome)
	}
}
