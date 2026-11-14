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
	// outcomePass: the test phase ended successfully.
	outcomePass = "pass"
	// outcomeFail: the distribution built and its test phase failed.
	outcomeFail = "fail"
	// outcomeUnpackFailed: the archive could not be unpacked.
	outcomeUnpackFailed = "unpack_failed"
	// outcomeNoBuildScript: neither Makefile.PL nor Build.PL.
	outcomeNoBuildScript = "no_build_script"
	// outcomeConfigureFailed: Makefile.PL / Build.PL failed.
	outcomeConfigureFailed = "configure_failed"
	// outcomeDeclined: the build script ended successfully without
	// producing a Makefile / Build: it does not support this platform.
	outcomeDeclined = "declined"
	// outcomeUnmet: a requirement the distribution computed while it was
	// configured cannot be met (perl itself, a module the index does not
	// know, or one whose distribution does not provide what is asked).
	outcomeUnmet = "prerequisites_unmet"
	// outcomeBuildFailed: make / Build failed.
	outcomeBuildFailed = "build_failed"
	// outcomeTimeout: a phase ran out of time.
	outcomeTimeout = "timeout"
	// outcomeOutputLimit: a phase was stopped because its output passed
	// the limit (a script that prompts in a loop on a closed stdin, a test
	// that prints without end).
	outcomeOutputLimit = "output_limit"
	// outcomeBlocked: a distribution this one needs could not be
	// installed: one it needs to be configured (it was not attempted),
	// one its configuration asked for (it was not attempted again), or a
	// static requirement when its configuration failed.
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
	Installed bool               `json:"installed"`
	Phases    []phaseRecord      `json:"phases,omitempty"`
	Tests     []testFile         `json:"tests,omitempty"`
	Unmet     []unmetRequirement `json:"unmet,omitempty"`
	BlockedBy []string           `json:"blocked_by,omitempty"`
	// Attempts counts the container runs: a distribution is run again
	// after the requirements it discovered while configuring are installed.
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

// maxAttempts bounds how often a distribution is configured again because
// configuring it revealed further requirements.
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
	volume   string
	results  *os.File
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
	if out, err := exec.Command("docker", "volume", "create", r.volume).CombinedOutput(); err != nil {
		return fmt.Errorf("docker volume create: %v: %s", err, out)
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

	finished := make(chan attempted)
	active := 0
	sequence := 0
	for {
		for active < opts.jobs && r.ready.Len() > 0 {
			n := heap.Pop(&r.ready).(*node)
			active++
			sequence++
			if n.attempts == 0 {
				n.started = time.Now()
			}
			n.attempts++
			go func(n *node, attempts, seq int) {
				finished <- attempted{n, r.attempt(n.d, attempts, seq)}
			}(n, n.attempts, sequence)
		}
		if active == 0 {
			break
		}
		a := <-finished
		active--
		if err := r.settle(a.n, a.res); err != nil {
			return err
		}
	}
	// What is left waits for something that never finished: distributions
	// that need each other.
	var stuck []string
	for name, n := range r.nodes {
		if n.res == nil {
			stuck = append(stuck, name)
		}
	}
	sort.Strings(stuck)
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
			if err := r.release(n); err != nil {
				return nil, err
			}
		}
	}
	return r.nodes[name], nil
}

// sameCycle reports whether two distributions need each other, directly or
// indirectly: neither can wait for the other.
func sameCycle(a, b *dist) bool { return a.Cycle != 0 && a.Cycle == b.Cycle }

// release makes a node whose needs have all finished runnable, or settles
// it as blocked when a distribution it needs to configure is not installed.
//
// The other static requirements, the ones the graph was built from, do not
// settle it: a release whose configuration is dynamic computes its own while
// it is configured (MYMETA), and those can leave out what it needs only on
// other systems (Apache-Test 1.43 lists Win32::Process). settle blocks it
// on the computed requirements that are not met. Configure requirements
// are static by the CPAN::Meta specification.
func (r *runner) release(n *node) error {
	if blocked := r.notInstalledNeeds(n, true); len(blocked) > 0 {
		res := r.newResult(n.d, n.attempts)
		res.Outcome = outcomeBlocked
		res.BlockedBy = blocked
		return r.record(n, res)
	}
	heap.Push(&r.ready, n)
	return nil
}

// notInstalledNeeds returns, sorted, the distributions n needs by its static
// requirements that finished without being installed; with configureOnly,
// only the ones it needs to be configured. A distribution of n's own
// dependency cycle is left out: n could not wait for it.
func (r *runner) notInstalledNeeds(n *node, configureOnly bool) []string {
	var configure map[string]bool
	if configureOnly {
		configure = map[string]bool{}
		for _, q := range n.d.Requires {
			if q.Phase == "configure" && q.Relationship == "requires" && q.Dist != "" {
				configure[q.Dist] = true
			}
		}
	}
	var out []string
	for _, need := range n.d.Needs {
		if configureOnly && !configure[need] {
			continue
		}
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
	out := filepath.Join(r.runDir, "work", d.Dist)
	if err := os.RemoveAll(out); err != nil {
		return fail(err)
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		return fail(err)
	}
	mode := "build"
	if res.Shipped {
		mode = "shipped"
	}
	name := fmt.Sprintf("cpancompat-%s-%d", r.opts.run, seq)
	ctx, cancel := context.WithTimeout(context.Background(), r.opts.wall)
	defer cancel()
	cmd := exec.CommandContext(ctx, "docker", "run", "--rm", "--name", name,
		"--cpus", r.opts.cpus, "--memory", r.opts.memory, "--pids-limit", "2048",
		"-v", r.volume+":/opt/gperl/local",
		"-v", r.archives+":/archives:ro",
		"-v", out+":/out",
		r.opts.image, "bash", "/opt/cpancompat/dist.sh", "/archives/"+d.Path, mode)
	cmd.Cancel = func() error { return exec.Command("docker", "kill", name).Run() }
	log, err := os.Create(filepath.Join(out, "container.log"))
	if err != nil {
		return fail(err)
	}
	cmd.Stdout, cmd.Stderr = log, log
	runErr := cmd.Run()
	if err := log.Close(); err != nil {
		return fail(err)
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
	for _, p := range res.Phases {
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
			if failed {
				return timeoutOr(p, outcomeConfigureFailed), false
			}
		case "declined":
			return outcomeDeclined, false
		case "prerequisites":
			if failed {
				if len(res.Unmet) > 0 {
					return outcomeUnmet, false
				}
				return timeoutOr(p, outcomeDriverFailed), false
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

// settle takes a distribution back from its worker: it either waits for
// the requirements its configuration revealed, or its result is final.
func (r *runner) settle(n *node, res *result) error {
	if res.Outcome == outcomeConfigureFailed {
		// A configure script commonly loads what the distribution needs
		// without declaring it for configuration, so a failure there is
		// put down to a static requirement that is not installed. Its
		// phases stay in the record.
		if blocked := r.notInstalledNeeds(n, false); len(blocked) > 0 {
			res.Outcome = outcomeBlocked
			res.BlockedBy = blocked
		}
	}
	if res.Outcome == outcomeUnmet && n.attempts < maxAttempts {
		wait := map[string]bool{}
		final := false
		var blocked []string
		for _, u := range res.Unmet {
			provider, known := r.byModule[u.Module]
			if !known || provider == n.d.Dist {
				// perl itself, a module no distribution provides, or
				// one this distribution should have brought.
				final = true
				continue
			}
			m, err := r.ensureNode(provider)
			if err != nil {
				return err
			}
			if m == nil {
				final = true
				continue
			}
			if m.res != nil {
				// Finished already: it is not installed, or what it
				// installs does not meet the requirement.
				if !m.res.Installed {
					blocked = append(blocked, provider)
				}
				final = true
				continue
			}
			wait[provider] = true
		}
		if !final && len(wait) > 0 {
			n.waiting = wait
			for provider := range wait {
				r.waiters[provider] = append(r.waiters[provider], n)
			}
			return nil
		}
		if len(blocked) > 0 {
			sort.Strings(blocked)
			res.Outcome = outcomeBlocked
			res.BlockedBy = blocked
		}
	}
	return r.record(n, res)
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
		if err := r.release(w); err != nil {
			return err
		}
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
