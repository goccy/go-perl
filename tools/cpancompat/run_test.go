package main

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"sync"
	"testing"
)

// testRunner returns a runner of the given distributions, recording into a
// temporary file, with modules mapped to the distributions that provide
// them.
func testRunner(t *testing.T, dists ...*dist) *runner {
	t.Helper()
	results, err := os.Create(filepath.Join(t.TempDir(), "results.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { results.Close() })
	r := &runner{
		opts:     runOptions{jobs: 4},
		dists:    map[string]*dist{},
		byModule: map[string]string{},
		previous: map[string]*result{},
		nodes:    map[string]*node{},
		waiters:  map[string][]*node{},
		results:  results,
	}
	for _, d := range dists {
		r.dists[d.Dist] = d
		for _, m := range d.Modules {
			r.byModule[m] = d.Dist
		}
	}
	return r
}

// blockingRunner returns a runner where the distribution "Dependent" needs
// "Configure" to be configured and "Runtime" to run (by its static
// requirements), with the given ones of those finished and not installed.
func blockingRunner(t *testing.T, notInstalled ...string) (*runner, *node) {
	t.Helper()
	r := testRunner(t)
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
	n := &node{d: d, waiting: map[string]bool{}, attempts: 1}
	r.nodes[d.Dist] = n
	r.byModule["Runtime::Module"] = "Runtime"
	r.byModule["Runtime::Other"] = "Runtime"
	return r, n
}

func TestReleaseRunsWhateverItsNeedsDid(t *testing.T) {
	// Requirements left unmet, even the ones it needs to be configured: it
	// is configured, built and tested anyway, as CPAN.pm does.
	r, n := blockingRunner(t, "Configure", "Runtime")
	r.release(n)
	if n.res != nil || r.ready.Len() != 1 {
		t.Errorf("with its requirements not installed: settled %+v, ready %d; want it attempted", n.res, r.ready.Len())
	}
}

func TestEnsureNodeReleasesEachNodeOnce(t *testing.T) {
	// Scheduling A adds A, B and W. B needs, to be configured, C, which
	// an earlier run did not install: B is run anyway. W waits for B, and
	// A for W and B.
	r := testRunner(t,
		&dist{Dist: "A", Needs: []string{"W", "B"}},
		&dist{Dist: "B", Needs: []string{"C"}, Requires: []requirement{
			{Module: "C", Phase: "configure", Relationship: "requires", Dist: "C"},
		}},
		&dist{Dist: "C"},
		&dist{Dist: "W", Needs: []string{"B"}},
	)
	r.previous["C"] = &result{Dist: "C", Outcome: outcomeBuildFailed}
	if _, err := r.ensureNode("A"); err != nil {
		t.Fatal(err)
	}
	ready := map[string]int{}
	for _, n := range r.ready {
		ready[n.d.Dist]++
	}
	if !reflect.DeepEqual(ready, map[string]int{"B": 1}) {
		t.Errorf("ready: %v; want B once", ready)
	}
	if res := r.nodes["B"].res; res != nil {
		t.Errorf("B settled as %+v; want it attempted", res)
	}
}

func TestRecordStuckBlocksEachByWhatItWaitedFor(t *testing.T) {
	// A and B wait for each other (as configuring them revealed); C waits
	// only for A. Recording A must not release C.
	a := &node{d: &dist{Dist: "A"}, waiting: map[string]bool{"B": true}}
	b := &node{d: &dist{Dist: "B"}, waiting: map[string]bool{"A": true}}
	c := &node{d: &dist{Dist: "C", Needs: []string{"A"}}, waiting: map[string]bool{"A": true}}
	r := testRunner(t)
	r.nodes = map[string]*node{"A": a, "B": b, "C": c}
	r.waiters = map[string][]*node{"A": {b, c}, "B": {a}}
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

func TestResumeRunsAgainWhatHasRun(t *testing.T) {
	// A and B have run and wait for each other; C has not run and waits
	// for A by its static requirements. Nothing runs: A and B are run
	// again, C goes on waiting for A.
	a := &node{d: &dist{Dist: "A"}, waiting: map[string]bool{"B": true}, attempts: 1}
	b := &node{d: &dist{Dist: "B"}, waiting: map[string]bool{"A": true}, attempts: 1}
	c := &node{d: &dist{Dist: "C", Needs: []string{"A"}}, waiting: map[string]bool{"A": true}}
	r := testRunner(t)
	r.nodes = map[string]*node{"A": a, "B": b, "C": c}
	r.waiters = map[string][]*node{"A": {b, c}, "B": {a}}
	if !r.resume() {
		t.Fatal("resume released nothing")
	}
	var ready []string
	for _, n := range r.ready {
		ready = append(ready, n.d.Dist)
	}
	sort.Strings(ready)
	if !reflect.DeepEqual(ready, []string{"A", "B"}) {
		t.Errorf("ready: %v; want A and B", ready)
	}
	if len(a.waiting) != 0 || len(b.waiting) != 0 || !c.waiting["A"] {
		t.Errorf("waiting: A %v, B %v, C %v; want A and B nothing, C A", a.waiting, b.waiting, c.waiting)
	}
	if !reflect.DeepEqual(r.waiters["A"], []*node{c}) || len(r.waiters["B"]) != 0 {
		t.Errorf("waiters of A %d, of B %d; want C only, none", len(r.waiters["A"]), len(r.waiters["B"]))
	}

	// A has run and waits for P, which its configuration asked for; P has
	// not run and needs A. Running A again would not help: P is run.
	a = &node{d: &dist{Dist: "A"}, waiting: map[string]bool{"P": true}, attempts: 1}
	p := &node{d: &dist{Dist: "P", Needs: []string{"A"}}, waiting: map[string]bool{"A": true}}
	r = testRunner(t)
	r.nodes = map[string]*node{"A": a, "P": p}
	r.waiters = map[string][]*node{"A": {p}, "P": {a}}
	if !r.resume() {
		t.Fatal("resume released nothing")
	}
	if r.ready.Len() != 1 || r.ready[0] != p || len(p.waiting) != 0 || !a.waiting["P"] {
		t.Errorf("ready %d (P first: %v), P waiting %v, A waiting %v; want P alone released", r.ready.Len(), r.ready.Len() > 0 && r.ready[0] == p, p.waiting, a.waiting)
	}
}

func TestSettleNamesEachBlockingDistributionOnce(t *testing.T) {
	// Two unmet modules come from one distribution that is not installed.
	r, n := blockingRunner(t, "Runtime")
	res := &result{Dist: "Dependent", Outcome: outcomeFail, Unmet: []unmetRequirement{
		{Module: "Runtime::Module", Wanted: "0"},
		{Module: "Runtime::Other", Wanted: "0"},
	}}
	if err := r.settle(n, res); err != nil {
		t.Fatal(err)
	}
	if res.Outcome != outcomeBlocked || !reflect.DeepEqual(res.BlockedBy, []string{"Runtime"}) || res.OwnOutcome != outcomeFail {
		t.Errorf("settled as %s by %v (ran: %s); want blocked by Runtime once (ran: fail)", res.Outcome, res.BlockedBy, res.OwnOutcome)
	}
}

func TestSettlePutsConfigureFailureDownToMissingNeeds(t *testing.T) {
	r, n := blockingRunner(t, "Runtime")
	res := &result{Dist: "Dependent", Outcome: outcomeConfigureFailed}
	if err := r.settle(n, res); err != nil {
		t.Fatal(err)
	}
	if res.Outcome != outcomeBlocked || !reflect.DeepEqual(res.BlockedBy, []string{"Runtime"}) || res.OwnOutcome != outcomeConfigureFailed {
		t.Errorf("configure failed with Runtime not installed: %s by %v (ran: %s); want blocked by Runtime (ran: configure_failed)",
			res.Outcome, res.BlockedBy, res.OwnOutcome)
	}

	r, n = blockingRunner(t)
	res = &result{Dist: "Dependent", Outcome: outcomeConfigureFailed}
	if err := r.settle(n, res); err != nil {
		t.Fatal(err)
	}
	if res.Outcome != outcomeConfigureFailed || res.BlockedBy != nil || res.OwnOutcome != "" {
		t.Errorf("configure failed with everything installed: %s by %v (ran: %s); want configure_failed", res.Outcome, res.BlockedBy, res.OwnOutcome)
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

func TestSettleWaitsOnlyForProvidersThatHaveNotFinished(t *testing.T) {
	// Runtime has not finished: the distribution, which failed without
	// its module, waits for it and is run again.
	r, n := blockingRunner(t)
	res := &result{Dist: "Dependent", Outcome: outcomeFail, Unmet: []unmetRequirement{{Module: "Runtime::Module", Wanted: "0"}}}
	if err := r.settle(n, res); err != nil {
		t.Fatal(err)
	}
	if n.res != nil || !reflect.DeepEqual(n.waiting, map[string]bool{"Runtime": true}) || !reflect.DeepEqual(r.waiters["Runtime"], []*node{n}) {
		t.Errorf("with Runtime not finished: settled %+v, waiting %v; want it waiting for Runtime", n.res, n.waiting)
	}

	// Runtime finished without being installed: nothing to wait for.
	r, n = blockingRunner(t, "Runtime")
	res = &result{Dist: "Dependent", Outcome: outcomeBuildFailed, Unmet: []unmetRequirement{{Module: "Runtime::Module", Wanted: "0"}}}
	if err := r.settle(n, res); err != nil {
		t.Fatal(err)
	}
	if n.res != res || res.Outcome != outcomeBlocked || res.OwnOutcome != outcomeBuildFailed || len(res.Unmet) != 1 {
		t.Errorf("with Runtime not installed: settled %+v; want blocked by Runtime, ran build_failed, unmet kept", n.res)
	}

	// After its last attempt it waits no longer, and is blocked by what
	// has not finished.
	r, n = blockingRunner(t)
	n.attempts = maxAttempts
	res = &result{Dist: "Dependent", Outcome: outcomeFail, Unmet: []unmetRequirement{{Module: "Runtime::Module", Wanted: "0"}}}
	if err := r.settle(n, res); err != nil {
		t.Fatal(err)
	}
	if n.res != res || res.Outcome != outcomeBlocked || !reflect.DeepEqual(res.BlockedBy, []string{"Runtime"}) {
		t.Errorf("after the last attempt: settled %+v; want blocked by Runtime", n.res)
	}
}

func TestSettlePassIsPassWithRequirementsUnmet(t *testing.T) {
	r, n := blockingRunner(t)
	unmet := []unmetRequirement{{Module: "Runtime::Module", Wanted: "0"}, {Module: "perl", Wanted: "5.050", Have: "5.044000"}}
	res := &result{Dist: "Dependent", Outcome: outcomePass, Installed: true, Unmet: unmet}
	if err := r.settle(n, res); err != nil {
		t.Fatal(err)
	}
	if n.res != res || res.Outcome != outcomePass || res.OwnOutcome != "" || res.BlockedBy != nil || len(res.Unmet) != 2 {
		t.Errorf("passed with requirements unmet: settled %+v; want pass, unmet kept", n.res)
	}
}

func TestSettleUnmetWithoutProviderToBlame(t *testing.T) {
	// perl itself, a module no distribution provides, and a module whose
	// distribution is installed and does not provide what is asked.
	r, n := blockingRunner(t)
	r.nodes["Runtime"].res = &result{Dist: "Runtime", Outcome: outcomePass, Installed: true}
	res := &result{Dist: "Dependent", Outcome: outcomeFail, Unmet: []unmetRequirement{
		{Module: "perl", Wanted: "5.050", Have: "5.044000"},
		{Module: "type::Tiny", Wanted: "0"},
		{Module: "Runtime::Module", Wanted: "< 1.0", Have: "1.2"},
	}}
	if err := r.settle(n, res); err != nil {
		t.Fatal(err)
	}
	if n.res != res || res.Outcome != outcomeUnmet || res.OwnOutcome != outcomeFail || res.BlockedBy != nil {
		t.Errorf("settled as %s by %v (ran: %s); want prerequisites_unmet (ran: fail)", res.Outcome, res.BlockedBy, res.OwnOutcome)
	}
}

func TestOutcomeOf(t *testing.T) {
	ok := func(name string) phaseRecord { return phaseRecord{Phase: name} }
	failed := func(name string, status int) phaseRecord { return phaseRecord{Phase: name, Status: status} }
	rest := []phaseRecord{ok("prerequisites"), ok("build"), ok("test"), ok("install"), ok("publish")}
	for _, tc := range []struct {
		name      string
		phases    []phaseRecord
		unmet     []unmetRequirement
		outcome   string
		installed bool
	}{
		{
			name:    "configure wrote its build script and exited non-zero",
			phases:  append([]phaseRecord{ok("unpack"), failed("configure", 2)}, rest...),
			outcome: outcomePass, installed: true,
		},
		{
			name:    "configure exited non-zero without a build script",
			phases:  []phaseRecord{ok("unpack"), failed("configure", 2)},
			outcome: outcomeConfigureFailed,
		},
		{
			name:    "configure ran out of time without a build script",
			phases:  []phaseRecord{ok("unpack"), {Phase: "configure", Status: 124, TimedOut: true}},
			outcome: outcomeTimeout,
		},
		{
			name:    "requirements unmet and tests failing",
			phases:  []phaseRecord{ok("unpack"), ok("configure"), ok("prerequisites"), ok("build"), failed("test", 2), ok("install"), ok("publish")},
			unmet:   []unmetRequirement{{Module: "Foo", Wanted: "0"}},
			outcome: outcomeFail, installed: true,
		},
		{
			name:    "the requirements helper failed",
			phases:  []phaseRecord{ok("unpack"), ok("configure"), failed("prerequisites", 3)},
			unmet:   []unmetRequirement{{Module: "Foo", Wanted: "0"}},
			outcome: outcomeDriverFailed,
		},
	} {
		res := &result{Phases: tc.phases, Unmet: tc.unmet}
		outcome, installed := outcomeOf(res, false)
		if outcome != tc.outcome || installed != tc.installed {
			t.Errorf("%s: %s, installed %v; want %s, installed %v", tc.name, outcome, installed, tc.outcome, tc.installed)
		}
	}
}

// installedSet is what a fake attempt has installed so far.
type installedSet struct {
	mu   sync.Mutex
	dist map[string]bool
}

func (s *installedSet) has(name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.dist[name]
}

func (s *installedSet) add(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dist[name] = true
}

func TestScheduleRunsDependencyCycleMembersAgain(t *testing.T) {
	// Chemistry-Ring test-requires Chemistry-File-SMILES, which needs
	// Chemistry-Ring: a dependency cycle. Both are run at the same time
	// and fail without the other, and both build, so both are installed.
	// They wait for each other; once nothing else can run, both are run
	// again, find each other, and pass.
	ring := &dist{Dist: "Chemistry-Ring", Modules: []string{"Chemistry::Ring"}, Needs: []string{"Chemistry-File-SMILES"}, Cycle: 1}
	smiles := &dist{Dist: "Chemistry-File-SMILES", Modules: []string{"Chemistry::File::SMILES"}, Needs: []string{"Chemistry-Ring"}, Cycle: 1}
	r := testRunner(t, ring, smiles)
	for _, name := range []string{ring.Dist, smiles.Dist} {
		if _, err := r.ensureNode(name); err != nil {
			t.Fatal(err)
		}
	}
	installed := &installedSet{dist: map[string]bool{}}
	other := map[string]struct{ dist, module string }{
		ring.Dist:   {smiles.Dist, "Chemistry::File::SMILES"},
		smiles.Dist: {ring.Dist, "Chemistry::Ring"},
	}
	var first sync.WaitGroup
	first.Add(2)
	err := r.schedule(func(d *dist, attempts, seq int) *result {
		res := r.newResult(d, attempts)
		needs := other[d.Dist]
		has := installed.has(needs.dist)
		if attempts == 1 {
			// Both first runs look before either has installed.
			first.Done()
			first.Wait()
		}
		res.Outcome, res.Installed = outcomePass, true
		if !has {
			res.Outcome = outcomeFail
			res.Unmet = []unmetRequirement{{Module: needs.module, Wanted: "0"}}
		}
		installed.add(d.Dist)
		return res
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{ring.Dist, smiles.Dist} {
		res := r.nodes[name].res
		if res == nil || res.Outcome != outcomePass || res.Attempts != 2 {
			t.Errorf("%s settled as %+v; want pass on the second attempt", name, res)
		}
	}
}

func TestScheduleRunsPastProviderThatFails(t *testing.T) {
	// strictures' configuration asks for a module of a distribution that
	// does not build. strictures waits for it, is run again once it has
	// failed, and is blocked by it, keeping how it went itself.
	strictures := &dist{Dist: "strictures", Modules: []string{"strictures"}}
	provider := &dist{Dist: "indirect", Modules: []string{"indirect"}}
	for _, passes := range []bool{false, true} {
		r := testRunner(t, strictures, provider)
		if _, err := r.ensureNode(strictures.Dist); err != nil {
			t.Fatal(err)
		}
		var mu sync.Mutex
		runs := map[string]int{}
		err := r.schedule(func(d *dist, attempts, seq int) *result {
			mu.Lock()
			runs[d.Dist]++
			mu.Unlock()
			res := r.newResult(d, attempts)
			if d.Dist == provider.Dist {
				res.Outcome = outcomeBuildFailed
				return res
			}
			res.Outcome, res.Installed = outcomeFail, true
			if passes {
				res.Outcome = outcomePass
			}
			res.Unmet = []unmetRequirement{{Module: "indirect", Wanted: "0"}}
			return res
		})
		if err != nil {
			t.Fatal(err)
		}
		res := r.nodes[strictures.Dist].res
		switch {
		case passes:
			if res == nil || res.Outcome != outcomePass || runs[strictures.Dist] != 1 {
				t.Errorf("passing without indirect: %+v after %d runs; want pass after 1", res, runs[strictures.Dist])
			}
		default:
			if res == nil || res.Outcome != outcomeBlocked || res.OwnOutcome != outcomeFail ||
				!reflect.DeepEqual(res.BlockedBy, []string{"indirect"}) || runs[strictures.Dist] != 2 {
				t.Errorf("failing without indirect: %+v after %d runs; want blocked by indirect (ran: fail) after 2", res, runs[strictures.Dist])
			}
		}
		if res := r.nodes[provider.Dist].res; res == nil || res.Outcome != outcomeBuildFailed {
			t.Errorf("indirect settled as %+v; want build_failed", res)
		}
	}
}
