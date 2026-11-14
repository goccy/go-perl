package main

import (
	"bufio"
	"container/heap"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// The outcome of running one distribution's own build and test suite on
// the interpreter.
const (
	// outcomePass: the test phase ended successfully. A distribution that
	// passes is a pass even when requirements it named were not met.
	outcomePass = "pass"
	// outcomeFail: the distribution built and its test phase failed.
	outcomeFail = "fail"
	// outcomeUnpackFailed: the archive could not be unpacked.
	outcomeUnpackFailed = "unpack_failed"
	// outcomeNoBuildScript: neither Makefile.PL nor Build.PL.
	outcomeNoBuildScript = "no_build_script"
	// outcomeConfigureFailed: Makefile.PL / Build.PL did not write the
	// Makefile / Build, and exited with an error. (One that wrote it is
	// gone on from whatever its exit status, as CPAN clients do.)
	outcomeConfigureFailed = "configure_failed"
	// outcomeDeclined: the build script ended successfully without
	// producing a Makefile / Build: it does not support this platform.
	outcomeDeclined = "declined"
	// outcomeUnmet: the distribution did not pass, and a requirement it
	// computed while it was configured cannot be met (perl itself, a
	// module the index does not know, one of its own, or one whose
	// distribution is installed and does not provide what is asked). It
	// was built and tested anyway; OwnOutcome is how that went.
	outcomeUnmet = "prerequisites_unmet"
	// outcomeBuildFailed: make / Build failed.
	outcomeBuildFailed = "build_failed"
	// outcomeTimeout: a phase ran out of time (the exit status of its
	// time limit says so).
	outcomeTimeout = "timeout"
	// outcomeOutputLimit: a phase was stopped because its output passed
	// the limit (a script that prompts in a loop on a closed stdin, a test
	// that prints without end).
	outcomeOutputLimit = "output_limit"
	// outcomeBlocked: the distribution did not pass, and a distribution
	// providing a module it lacked was not installed when it ran: one
	// providing a requirement it computed while it was configured, or,
	// when its configuration failed (and so computed none), one it needs
	// by its static requirements. It was run anyway, and OwnOutcome is
	// how that went; BlockedBy names those distributions. A distribution
	// left waiting for distributions that never finished is blocked by
	// them without having been run.
	outcomeBlocked = "blocked"
	// outcomeDriverFailed: the run left no usable record (the container
	// could not start, or the helper that reads the configured
	// requirements failed).
	outcomeDriverFailed = "driver_failed"
)

// phaseRecord is one line of the phases.jsonl a container writes.
type phaseRecord struct {
	Phase         string `json:"phase"`
	Status        int    `json:"status"`
	Seconds       int    `json:"seconds"`
	TimedOut      bool   `json:"timed_out"`
	OutputLimited bool   `json:"output_limited,omitempty"`
	NoBuildScript bool   `json:"no_build_script,omitempty"`
}

// testFile is one line of tests.jsonl: the harness's verdict on one test
// file.
type testFile struct {
	File        string `json:"file"`
	Planned     int    `json:"planned"`
	Run         int    `json:"run"`
	Passed      int    `json:"passed"`
	Failed      int    `json:"failed"`
	TodoPassed  int    `json:"todo_passed"`
	Skipped     int    `json:"skipped"`
	SkipAll     bool   `json:"skip_all"`
	ParseErrors int    `json:"parse_errors"`
	Exit        int    `json:"exit"`
	Wait        int    `json:"wait"`
	Problems    bool   `json:"problems"`
}

// unmetRequirement is a requirement of the configured distribution that
// the installed modules do not meet. Have is "" when the module is absent.
type unmetRequirement struct {
	Module string `json:"module"`
	Wanted string `json:"wanted"`
	Have   string `json:"have,omitempty"`
}

// result is the record of one distribution in results.jsonl.
type result struct {
	Dist    string `json:"dist"`
	Release string `json:"release"`
	Outcome string `json:"outcome"`
	// Shipped: the interpreter ships the distribution, and its test files
	// ran against that copy.
	Shipped bool `json:"shipped,omitempty"`
	// Installed: distributions that need this one can be attempted. A
	// distribution is installed whenever it builds, whatever its tests say.
	Installed bool          `json:"installed"`
	Phases    []phaseRecord `json:"phases,omitempty"`
	Tests     []testFile    `json:"tests,omitempty"`
	// Unmet are the requirements the installed modules did not meet when
	// the distribution was configured.
	Unmet     []unmetRequirement `json:"unmet,omitempty"`
	BlockedBy []string           `json:"blocked_by,omitempty"`
	// OwnOutcome is the outcome of the distribution's own configure, build
	// and tests when Outcome puts it down to what it lacked (blocked,
	// prerequisites_unmet): it was run all the same.
	OwnOutcome string `json:"own_outcome,omitempty"`
	// Attempts counts the container runs: a distribution that did not pass
	// while requirements it computed when it was configured were unmet is
	// run again once the distributions providing them have finished.
	Attempts    int    `json:"attempts"`
	Seconds     int    `json:"seconds"`
	Interpreter string `json:"interpreter,omitempty"`
	Finished    string `json:"finished"`
}

type runOptions struct {
	dataDir     string
	run         string
	image       string
	interpreter string
	top         int
	only        []string
	redo        map[string]bool
	jobs        int
	cpus        string
	memory      string
	wall        time.Duration
}

// maxAttempts bounds how often a distribution is run again because it
// lacked requirements whose distributions had not finished.
const maxAttempts = 4

// node is one distribution in the schedule. Only the coordinating
// goroutine reads or writes it.
type node struct {
	d *dist
	// waiting holds the distributions that must finish before this one
	// is attempted (again).
	waiting  map[string]bool
	attempts int
	started  time.Time
	// res is the final result; nil until the distribution is settled.
	res *result
}

type runner struct {
	opts     runOptions
	dists    map[string]*dist
	byModule map[string]string
	provided map[string]string
	previous map[string]*result

	nodes   map[string]*node
	waiters map[string][]*node
	ready   readyHeap
	total   int
	done    int

	runDir   string
	archives string
	// volume is the run's install tree, shared by its containers;
	// buildVolume holds their build directories, which are kept as CPAN
	// clients keep theirs.
	volume      string
	buildVolume string
	results     *os.File
}

// attempted is what a worker hands back.
type attempted struct {
	n   *node
	res *result
}

// readyHeap orders the distributions that can run now: the one most others
// depend on first, so the schedule widens as early as possible.
type readyHeap []*node

func (h readyHeap) Len() int { return len(h) }
func (h readyHeap) Less(i, j int) bool {
	if h[i].d.DependentsTotal != h[j].d.DependentsTotal {
		return h[i].d.DependentsTotal > h[j].d.DependentsTotal
	}
	return h[i].d.Dist < h[j].d.Dist
}
func (h readyHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h *readyHeap) Push(x any)   { *h = append(*h, x.(*node)) }
func (h *readyHeap) Pop() any {
	old := *h
	n := old[len(old)-1]
	*h = old[:len(old)-1]
	return n
}

// runSurvey tests the selected distributions and everything they need, in
// dependency order, each in a container of its own.
func runSurvey(opts runOptions) error {
	r := &runner{
		opts:     opts,
		dists:    map[string]*dist{},
		byModule: map[string]string{},
		previous: map[string]*result{},
		nodes:    map[string]*node{},
		waiters:  map[string][]*node{},
		runDir:   filepath.Join(opts.dataDir, "runs", opts.run),
		archives: filepath.Join(opts.dataDir, "archives"),
		volume:   "cpancompat-" + opts.run,
	}
	r.buildVolume = r.volume + "-build"
	all, err := readDists(filepath.Join(opts.dataDir, "dists.jsonl"))
	if err != nil {
		return err
	}
	for i := range all {
		d := &all[i]
		r.dists[d.Dist] = d
		for _, m := range d.Modules {
			r.byModule[m] = d.Dist
		}
	}
	if r.provided, err = readProvided(filepath.Join(opts.dataDir, "provided.json")); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(r.runDir, "work"), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(r.archives, 0o755); err != nil {
		return err
	}
	resultsPath := filepath.Join(r.runDir, "results.jsonl")
	earlier, err := readJSONLines[result](resultsPath)
	if err != nil {
		return err
	}
	for i := range earlier {
		res := &earlier[i]
		if opts.redo[res.Outcome] {
			delete(r.previous, res.Dist)
		} else {
			r.previous[res.Dist] = res
		}
	}
	if r.results, err = os.OpenFile(resultsPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644); err != nil {
		return err
	}
	defer r.results.Close()
	for _, v := range []string{r.volume, r.buildVolume} {
		if out, err := exec.Command("docker", "volume", "create", v).CombinedOutput(); err != nil {
			return fmt.Errorf("docker volume create: %v: %s", err, out)
		}
	}
	if err := guardNetwork(opts.dataDir); err != nil {
		return err
	}

	targets, err := r.targets(all)
	if err != nil {
		return err
	}
	for _, name := range targets {
		if _, err := r.ensureNode(name); err != nil {
			return err
		}
	}
	fmt.Fprintf(os.Stderr, "cpancompat: %d distributions selected, %d scheduled with what they need, %d already recorded\n",
		len(targets), r.total, r.done)
	return r.schedule(r.attempt)
}

// schedule runs the scheduled distributions, at most -jobs at once, each by
// attempt (on a goroutine of its own), until every one is settled.
func (r *runner) schedule(attempt func(d *dist, attempts, seq int) *result) error {
	finished := make(chan attempted)
	active := 0
	sequence := 0
	for {
		for active < r.opts.jobs && r.ready.Len() > 0 {
			n := heap.Pop(&r.ready).(*node)
			active++
			sequence++
			if n.attempts == 0 {
				n.started = time.Now()
			}
			n.attempts++
			go func(n *node, attempts, seq int) {
				finished <- attempted{n, attempt(n.d, attempts, seq)}
			}(n, n.attempts, sequence)
		}
		if active == 0 {
			if r.resume() {
				continue
			}
			break
		}
		a := <-finished
		active--
		if err := r.settle(a.n, a.res); err != nil {
			return err
		}
	}
	return r.recordStuck()
}

// resume ends a standstill: nothing runs or can run, and distributions wait
// for one another, the members of a dependency cycle, or by requirements
// only their configuration revealed. Waiting for a distribution that has
// run before no longer helps: whatever its earlier runs built is installed.
// So the distributions that have run and wait only for such ones are run
// again (each now finds what the others built); when there are none, the
// ones that have not run and wait only for such ones are run. It reports
// whether it released any, which it always does at a standstill: the ones
// that have not run wait only by static requirements, which do not form
// cycles. Every run counts towards maxAttempts, after which a distribution
// no longer waits, so standstills end.
func (r *runner) resume() bool {
	var again, first []*node
	for _, n := range r.nodes {
		if n.res != nil || len(n.waiting) == 0 {
			continue
		}
		onlyRun := true
		for provider := range n.waiting {
			if r.nodes[provider].attempts == 0 {
				onlyRun = false
				break
			}
		}
		switch {
		case !onlyRun:
		case n.attempts > 0:
			again = append(again, n)
		default:
			first = append(first, n)
		}
	}
	if len(again) == 0 {
		again = first
	}
	for _, n := range again {
		for provider := range n.waiting {
			kept := r.waiters[provider][:0]
			for _, w := range r.waiters[provider] {
				if w != n {
					kept = append(kept, w)
				}
			}
			r.waiters[provider] = kept
		}
		n.waiting = map[string]bool{}
		r.release(n)
	}
	return len(again) > 0
}

// recordStuck settles, once nothing runs or can run and resume released
// nothing, what is left: distributions that wait for something that never
// finished. Each is blocked by the ones it waited for. (resume always
// releases some at a standstill, so nothing should be left.)
func (r *runner) recordStuck() error {
	var stuck []string
	for name, n := range r.nodes {
		if n.res == nil {
			stuck = append(stuck, name)
		}
	}
	sort.Strings(stuck)
	// They wait only for one another: recording one must not release
	// another, which would be run by no one and recorded as blocked by
	// nothing.
	for _, name := range stuck {
		delete(r.waiters, name)
	}
	for _, name := range stuck {
		n := r.nodes[name]
		var on []string
		for w := range n.waiting {
			on = append(on, w)
		}
		sort.Strings(on)
		res := r.newResult(n.d, n.attempts)
		res.Outcome = outcomeBlocked
		res.BlockedBy = on
		if err := r.record(n, res); err != nil {
			return err
		}
	}
	return nil
}

// targets returns the names of the distributions to measure: the ones
// named with -dists, or else the reference set (when the graph has the
// CPAN Testers reports of the interpreter's perl version), the most
// depended-upon first, the first -top of them.
func (r *runner) targets(all []dist) ([]string, error) {
	if len(r.opts.only) > 0 {
		for _, name := range r.opts.only {
			if r.dists[name] == nil {
				return nil, fmt.Errorf("run: unknown distribution %q", name)
			}
		}
		return r.opts.only, nil
	}
	perl, err := graphPerl(r.opts.dataDir)
	if err != nil {
		return nil, err
	}
	order := make([]*dist, 0, len(all))
	for i := range all {
		if perl == "" || all[i].passesOnPerl() {
			order = append(order, &all[i])
		}
	}
	if perl != "" {
		fmt.Fprintf(os.Stderr, "cpancompat: %d of %d distributions pass on perl %s on Linux by their CPAN Testers reports\n",
			len(order), len(all), perl)
	}
	sort.SliceStable(order, func(i, j int) bool {
		if order[i].DependentsTotal != order[j].DependentsTotal {
			return order[i].DependentsTotal > order[j].DependentsTotal
		}
		return order[i].Dist < order[j].Dist
	})
	if r.opts.top > 0 && r.opts.top < len(order) {
		order = order[:r.opts.top]
	}
	names := make([]string, len(order))
	for i, d := range order {
		names[i] = d.Dist
	}
	return names, nil
}

// ensureNode schedules the distribution and, first, everything it needs.
// It returns nil for a name the index does not know.
func (r *runner) ensureNode(name string) (*node, error) {
	if n := r.nodes[name]; n != nil {
		return n, nil
	}
	// Iterative: dependency chains are long.
	var added []*node
	stack := []string{name}
	for len(stack) > 0 {
		cur := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if r.nodes[cur] != nil {
			continue
		}
		d := r.dists[cur]
		if d == nil {
			continue
		}
		n := &node{d: d, waiting: map[string]bool{}}
		r.nodes[cur] = n
		r.total++
		if prev := r.previous[cur]; prev != nil {
			n.res = prev
			r.done++
			continue
		}
		added = append(added, n)
		stack = append(stack, d.Needs...)
	}
	// Every need of an added node has its node now.
	for _, n := range added {
		for _, need := range n.d.Needs {
			m := r.nodes[need]
			if m == nil || m == n || m.res != nil || n.waiting[need] || sameCycle(n.d, m.d) {
				continue
			}
			n.waiting[need] = true
			r.waiters[need] = append(r.waiters[need], n)
		}
	}
	for _, n := range added {
		if len(n.waiting) == 0 {
			r.release(n)
		}
	}
	return r.nodes[name], nil
}

// sameCycle reports whether two distributions need each other, directly or
// indirectly: neither can wait for the other.
func sameCycle(a, b *dist) bool { return a.Cycle != 0 && a.Cycle == b.Cycle }

// release makes a node whose needs have all finished runnable, whatever they
// did: a distribution whose requirements are not all met is still
// configured, built and tested, as CPAN.pm does (it warns and goes on), and
// settle puts a failure down to what it lacked.
func (r *runner) release(n *node) { heap.Push(&r.ready, n) }

// notInstalledNeeds returns, sorted, the distributions n needs by its static
// requirements that finished without being installed. A distribution of n's
// own dependency cycle is left out: n could not wait for it.
func (r *runner) notInstalledNeeds(n *node) []string {
	var out []string
	for _, need := range n.d.Needs {
		m := r.nodes[need]
		if m != nil && m.res != nil && !m.res.Installed && !sameCycle(n.d, m.d) {
			out = append(out, need)
		}
	}
	sort.Strings(out)
	return out
}

func (r *runner) newResult(d *dist, attempts int) *result {
	return &result{
		Dist:        d.Dist,
		Release:     d.Release,
		Attempts:    attempts,
		Interpreter: r.opts.interpreter,
	}
}

// attempt runs the distribution's container once and returns what it
// found. It runs on a worker goroutine and changes nothing shared.
//
// The container runs code from CPAN, which may be malicious: it gets no
// writable directory of this machine, only the archives read-only, and runs
// as an unprivileged user without capabilities, unable to gain any. It
// writes its records into its own /out, which is copied out once it has
// ended (copyOut). Its network reaches the internet but not the LAN, the
// VM's host or the VM (guardNetwork), and is taken away before its test
// phase, when dist.sh asks for that (converse).
func (r *runner) attempt(d *dist, attempts, seq int) *result {
	res := r.newResult(d, attempts)
	res.Shipped = bundled(d, r.provided)
	fail := func(err error) *result {
		res.Outcome = outcomeDriverFailed
		fmt.Fprintf(os.Stderr, "cpancompat: %s: %v\n", d.Dist, err)
		return res
	}
	if _, err := fetchArchive(r.opts.dataDir, d); err != nil {
		return fail(err)
	}
	work := filepath.Join(r.runDir, "work")
	out := filepath.Join(work, d.Dist)
	if err := os.RemoveAll(out); err != nil {
		return fail(err)
	}
	mode := "build"
	if res.Shipped {
		mode = "shipped"
	}
	name := fmt.Sprintf("cpancompat-%s-%d", r.opts.run, seq)
	// A container of that name is one an interrupted driver left.
	_ = exec.Command("docker", "rm", "--force", name).Run()
	ctx, cancel := context.WithTimeout(context.Background(), r.opts.wall)
	defer cancel()
	// With an init process as PID 1: xvfb-run, which the test phase runs
	// under, waits for a signal from the X server that a shell running as
	// PID 1 never gets.
	cmd := exec.CommandContext(ctx, "docker", "run", "--init", "--name", name, "--network", surveyNetwork,
		"--cap-drop", "ALL", "--security-opt", "no-new-privileges",
		"--cpus", r.opts.cpus, "--memory", r.opts.memory, "--pids-limit", "2048",
		"-v", r.volume+":/opt/gperl/local",
		"-v", r.buildVolume+":/work",
		"-v", r.archives+":/archives:ro",
		r.opts.image, "bash", "/opt/cpancompat/dist.sh", "/archives/"+d.Path, mode)
	cmd.Cancel = func() error { return exec.Command("docker", "kill", name).Run() }
	// What docker run prints goes to container.log once the records are
	// copied out: the directory is the container's until then.
	log, err := os.CreateTemp(work, "."+d.Dist+".container.log.")
	if err != nil {
		return fail(err)
	}
	defer os.Remove(log.Name())
	cmd.Stderr = log
	requests, err := cmd.StdoutPipe()
	if err != nil {
		return fail(err)
	}
	runErr := cmd.Start()
	if runErr == nil {
		converse(requests, log, func() { takeOffline(name, log) })
		runErr = cmd.Wait()
	}
	if err := log.Close(); err != nil {
		return fail(err)
	}
	copyErr := copyOut(name, out)
	_ = exec.Command("docker", "rm", "--force", name).Run()
	if err := os.Rename(log.Name(), filepath.Join(out, "container.log")); err != nil {
		return fail(err)
	}
	if copyErr != nil {
		return fail(copyErr)
	}

	if res.Phases, err = readJSONLines[phaseRecord](filepath.Join(out, "phases.jsonl")); err != nil {
		return fail(err)
	}
	if res.Tests, err = readJSONLines[testFile](filepath.Join(out, "tests.jsonl")); err != nil {
		return fail(err)
	}
	if raw, err := os.ReadFile(filepath.Join(out, "unmet.json")); err == nil && len(raw) > 0 {
		var u struct {
			Unmet []unmetRequirement `json:"unmet"`
		}
		// A file that does not parse is the helper having failed, which
		// the status of its phase reports.
		if json.Unmarshal(raw, &u) == nil {
			res.Unmet = u.Unmet
		}
	}
	res.Outcome, res.Installed = outcomeOf(res, errors.Is(ctx.Err(), context.DeadlineExceeded))
	if res.Outcome == outcomeDriverFailed && runErr != nil {
		fmt.Fprintf(os.Stderr, "cpancompat: %s: container: %v\n", d.Dist, runErr)
	}
	return res
}

// outcomeOf derives the outcome from the phases a container recorded.
// wallTimeout tells that the whole container was ended for running too
// long.
func outcomeOf(res *result, wallTimeout bool) (outcome string, installed bool) {
	for i, p := range res.Phases {
		failed := p.Status != 0
		switch p.Phase {
		case "unpack":
			if failed {
				return timeoutOr(p, outcomeUnpackFailed), false
			}
		case "configure":
			if p.NoBuildScript {
				return outcomeNoBuildScript, false
			}
			// A configure script that wrote the build script is gone on
			// from whatever its exit status, so a failed one is the last
			// phase only when it did not write it.
			if failed && i == len(res.Phases)-1 {
				return timeoutOr(p, outcomeConfigureFailed), false
			}
		case "declined":
			return outcomeDeclined, false
		case "prerequisites":
			// The helper that reads the configured requirements failed;
			// what it finds unmet does not stop the distribution.
			if failed {
				return timeoutOr(p, outcomeDriverFailed), false
			}
		case "offline":
			// The container could not be taken off its network before
			// its test phase, which then did not run.
			if failed {
				return outcomeDriverFailed, false
			}
		case "build":
			if failed {
				return timeoutOr(p, outcomeBuildFailed), false
			}
		case "test":
			outcome = outcomePass
			if failed {
				outcome = timeoutOr(p, outcomeFail)
			}
			// What the interpreter ships is there whatever its tests say.
			installed = res.Shipped
		case "publish":
			installed = !failed
		}
	}
	if outcome == "" {
		// The container ended before its test phase did.
		if wallTimeout {
			return outcomeTimeout, false
		}
		return outcomeDriverFailed, false
	}
	return outcome, installed
}

// timeoutOr is the outcome of a failed phase: outcome, unless the phase was
// stopped for running out of time or for writing too much.
func timeoutOr(p phaseRecord, outcome string) string {
	switch {
	case p.TimedOut:
		return outcomeTimeout
	case p.OutputLimited:
		return outcomeOutputLimit
	}
	return outcome
}

// settle takes a distribution back from its worker. One that did not pass
// while requirements it computed when it was configured were unmet waits for
// the distributions providing them that have not finished, and is run again
// once they have, so that they are installed first when they can be; it
// does not wait for one that finished without being installed. Otherwise
// its result is final, and a failure is put down to what it lacked.
func (r *runner) settle(n *node, res *result) error {
	providers, err := r.providers(n, res)
	if err != nil {
		return err
	}
	if res.Outcome != outcomePass && res.Outcome != outcomeDriverFailed && n.attempts < maxAttempts {
		wait := map[string]bool{}
		for _, provider := range providers {
			if r.nodes[provider].res == nil {
				wait[provider] = true
			}
		}
		if len(wait) > 0 {
			n.waiting = wait
			for provider := range wait {
				r.waiters[provider] = append(r.waiters[provider], n)
			}
			return nil
		}
	}
	r.attribute(n, res, providers)
	return r.record(n, res)
}

// providers returns, sorted and once each, the distributions providing the
// modules res found unmet, scheduled (with what they need) when they were
// not yet. Perl itself, a module the index does not know and one of the
// distribution's own have no provider to wait for.
func (r *runner) providers(n *node, res *result) ([]string, error) {
	var out []string
	seen := map[string]bool{}
	for _, u := range res.Unmet {
		provider, known := r.byModule[u.Module]
		if !known || provider == n.d.Dist || seen[provider] {
			continue
		}
		seen[provider] = true
		m, err := r.ensureNode(provider)
		if err != nil {
			return nil, err
		}
		if m != nil {
			out = append(out, provider)
		}
	}
	sort.Strings(out)
	return out, nil
}

// attribute puts a final result that is not a pass down to what the
// distribution lacked, keeping its own outcome in OwnOutcome. It is blocked
// by the providers of its unmet modules that were not installed, and, when
// its configuration failed (so that it computed no requirements), by the
// distributions its static requirements name that were not installed: a
// configure script commonly loads what the distribution needs without
// declaring it for configuration. With none of those but unmet modules all
// the same, its requirements could not be met.
func (r *runner) attribute(n *node, res *result, providers []string) {
	if res.Outcome == outcomePass || res.Outcome == outcomeDriverFailed {
		return
	}
	blockers := map[string]bool{}
	for _, provider := range providers {
		if m := r.nodes[provider]; m.res == nil || !m.res.Installed {
			blockers[provider] = true
		}
	}
	if res.Outcome == outcomeConfigureFailed {
		for _, need := range r.notInstalledNeeds(n) {
			blockers[need] = true
		}
	}
	switch {
	case len(blockers) > 0:
		res.OwnOutcome, res.Outcome, res.BlockedBy = res.Outcome, outcomeBlocked, nil
		for provider := range blockers {
			res.BlockedBy = append(res.BlockedBy, provider)
		}
		sort.Strings(res.BlockedBy)
	case len(res.Unmet) > 0:
		res.OwnOutcome, res.Outcome = res.Outcome, outcomeUnmet
	}
}

// record makes a result final: it is written out, and the distributions
// waiting for this one are released when it was the last they waited for.
func (r *runner) record(n *node, res *result) error {
	if !n.started.IsZero() {
		res.Seconds = int(time.Since(n.started).Seconds())
	}
	res.Finished = time.Now().UTC().Format(time.RFC3339)
	n.res = res
	r.done++
	line, err := json.Marshal(res)
	if err != nil {
		return err
	}
	if _, err := r.results.Write(append(line, '\n')); err != nil {
		return err
	}
	detail := ""
	if len(res.BlockedBy) > 0 {
		detail = " by " + strings.Join(res.BlockedBy, ", ")
	}
	if res.OwnOutcome != "" {
		detail += " (ran: " + res.OwnOutcome + ")"
	}
	fmt.Fprintf(os.Stderr, "[%d/%d] %-20s %s (%ds)%s\n", r.done, r.total, res.Outcome, n.d.Dist, res.Seconds, detail)

	waiters := r.waiters[n.d.Dist]
	delete(r.waiters, n.d.Dist)
	for _, w := range waiters {
		if w.res != nil || !w.waiting[n.d.Dist] {
			continue
		}
		delete(w.waiting, n.d.Dist)
		if len(w.waiting) > 0 {
			continue
		}
		r.release(w)
	}
	return nil
}

// readJSONLines reads a file of one JSON value per line. A file that does
// not exist has no lines.
func readJSONLines[T any](path string) ([]T, error) {
	f, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var rows []T
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<26)
	for sc.Scan() {
		if len(sc.Bytes()) == 0 {
			continue
		}
		var row T
		if err := json.Unmarshal(sc.Bytes(), &row); err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		rows = append(rows, row)
	}
	return rows, sc.Err()
}
