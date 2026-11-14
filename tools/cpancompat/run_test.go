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

func TestEnsureNodeReleasesEachNodeOnce(t *testing.T) {
	// Scheduling A adds A, B and W. B needs, to be configured, C, which
	// an earlier run did not install, so releasing B settles it as
	// blocked; W waited only for B, so recording B releases W. W must
	// not be released again as one of the nodes that wait for nothing.
	results, err := os.Create(filepath.Join(t.TempDir(), "results.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	defer results.Close()
	r := &runner{
		dists: map[string]*dist{
			"A": {Dist: "A", Needs: []string{"W", "B"}},
			"B": {Dist: "B", Needs: []string{"C"}, Requires: []requirement{
				{Module: "C", Phase: "configure", Relationship: "requires", Dist: "C"},
			}},
			"C": {Dist: "C"},
			"W": {Dist: "W", Needs: []string{"B"}},
		},
		previous: map[string]*result{"C": {Dist: "C", Outcome: outcomeBuildFailed}},
		nodes:    map[string]*node{},
		waiters:  map[string][]*node{},
		results:  results,
	}
	if _, err := r.ensureNode("A"); err != nil {
		t.Fatal(err)
	}
	ready := map[string]int{}
	for _, n := range r.ready {
		ready[n.d.Dist]++
	}
	if !reflect.DeepEqual(ready, map[string]int{"W": 1}) {
		t.Errorf("ready: %v; want W once", ready)
	}
	if res := r.nodes["B"].res; res == nil || res.Outcome != outcomeBlocked {
		t.Errorf("B settled as %+v; want blocked", res)
	}
}

func TestRecordStuckBlocksEachByWhatItWaitedFor(t *testing.T) {
	// A and B wait for each other (as configuring them revealed); C waits
	// only for A. Recording A must not release C.
	results, err := os.Create(filepath.Join(t.TempDir(), "results.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	defer results.Close()
	a := &node{d: &dist{Dist: "A"}, waiting: map[string]bool{"B": true}}
	b := &node{d: &dist{Dist: "B"}, waiting: map[string]bool{"A": true}}
	c := &node{d: &dist{Dist: "C", Needs: []string{"A"}}, waiting: map[string]bool{"A": true}}
	r := &runner{
		nodes:   map[string]*node{"A": a, "B": b, "C": c},
		waiters: map[string][]*node{"A": {b, c}, "B": {a}},
		results: results,
	}
	if err := r.recordStuck(); err != nil {
		t.Fatal(err)
	}
	for name, want := range map[string][]string{"A": {"B"}, "B": {"A"}, "C": {"A"}} {
		res := r.nodes[name].res
		if res == nil || res.Outcome != outcomeBlocked || !reflect.DeepEqual(res.BlockedBy, want) {
			t.Errorf("%s settled as %+v; want blocked by %v", name, res, want)
		}
	}
	if r.ready.Len() != 0 || r.done != 3 {
		t.Errorf("%d ready, %d recorded; want none ready, 3 recorded", r.ready.Len(), r.done)
	}
}

func TestSettleNamesEachBlockingDistributionOnce(t *testing.T) {
	// Two unmet modules come from one distribution that is not installed.
	r, n := blockingRunner(t, "Runtime")
	r.byModule = map[string]string{"Runtime::Module": "Runtime", "Runtime::Other": "Runtime"}
	res := &result{Dist: "Dependent", Outcome: outcomeUnmet, Unmet: []unmetRequirement{
		{Module: "Runtime::Module", Wanted: "0"},
		{Module: "Runtime::Other", Wanted: "0"},
	}}
	n.attempts = 1
	if err := r.settle(n, res); err != nil {
		t.Fatal(err)
	}
	if res.Outcome != outcomeBlocked || !reflect.DeepEqual(res.BlockedBy, []string{"Runtime"}) {
		t.Errorf("settled as %s by %v; want blocked by Runtime once", res.Outcome, res.BlockedBy)
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
