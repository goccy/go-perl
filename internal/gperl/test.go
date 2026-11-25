package gperl

// gperl test: run a project's test files (t/*.t) the way prove does. Each
// file runs as a perl process of its own, several at once. With
// --in-process they run in ONE process instead — every test file gets its
// own interpreter instance, and the instances run in parallel on
// goroutines, so a suite pays no process per file — where a test file is
// an isolated program (see perlRunOpts.isolated): it is not the host
// process, and cannot listen on the network, say.

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// tapResult is what one test file's TAP stream amounted to.
type tapResult struct {
	planned   int   // the plan's test count; -1 when no plan was seen
	run       int   // test lines seen
	failed    []int // numbers of the tests that failed (not counting TODO)
	todoOK    []int // TODO tests that unexpectedly passed
	skipAll   string
	skippedN  int
	bailout   string
	hasBailed bool
}

var (
	tapPlan = regexp.MustCompile(`^1\.\.(\d+)(?:\s*#\s*(?i:skip)\S*\s*(.*))?$`)
	// After the description, a # starts a directive (TODO, SKIP) or else a
	// comment.
	tapTest = regexp.MustCompile(`^(not )?ok\b(?:\s+(\d+))?([^#]*)(?:#\s*(?:(?i:(todo|skip))\b)?.*)?$`)
	tapBail = regexp.MustCompile(`^Bail out!\s*(.*)$`)
)

// parseTAP reads a test's standard output. Only top-level lines count:
// subtests are indented, and their outcome is restated by a top-level line.
func parseTAP(r io.Reader, each func(line string)) (tapResult, error) {
	res := tapResult{planned: -1}
	sc := bufio.NewScanner(r)
	sc.Buffer(nil, 4<<20)
	for sc.Scan() {
		line := sc.Text()
		if each != nil {
			each(line)
		}
		switch {
		case res.hasBailed:
		case tapBail.MatchString(line):
			res.hasBailed, res.bailout = true, tapBail.FindStringSubmatch(line)[1]
		case tapPlan.MatchString(line):
			m := tapPlan.FindStringSubmatch(line)
			// The pattern admits digits only; a count too large to be a
			// plan stays "no plan".
			if n, err := strconv.Atoi(m[1]); err == nil {
				res.planned = n
				if n == 0 {
					res.skipAll = m[2]
				}
			}
		case tapTest.MatchString(line):
			m := tapTest.FindStringSubmatch(line)
			res.run++
			number := res.run
			if n, err := strconv.Atoi(m[2]); err == nil {
				number = n
			}
			directive := strings.ToLower(m[4])
			switch {
			case directive == "skip":
				res.skippedN++
			case m[1] != "" && directive != "todo":
				res.failed = append(res.failed, number)
			case m[1] == "" && directive == "todo":
				res.todoOK = append(res.todoOK, number)
			}
		}
	}
	return res, sc.Err()
}

// testOutcome is one test file's result.
type testOutcome struct {
	file     string
	tap      tapResult
	status   int    // the program's exit status
	problem  string // why the file failed, "" when it passed
	output   []byte // its standard output, kept for verbose mode
	duration time.Duration
}

// judge decides whether the file passed, the way TAP::Harness does.
func (o *testOutcome) judge() {
	t := o.tap
	switch {
	case t.hasBailed:
		o.problem = "Bailed out: " + t.bailout
	case len(t.failed) > 0:
		o.problem = fmt.Sprintf("Failed %d/%d subtests", len(t.failed), max(t.run, t.planned))
	case t.planned < 0:
		o.problem = "No plan found in TAP output"
	case t.planned != t.run:
		o.problem = fmt.Sprintf("Parse errors: Bad plan.  You planned %d tests but ran %d.", t.planned, t.run)
	case o.status != 0:
		o.problem = fmt.Sprintf("Dubious, test returned %d", o.status)
	}
}

// testConfig is the command line of gperl test.
type testConfig struct {
	verbose   bool
	quiet     bool
	lib       bool
	recurse   bool
	timer     bool
	inProcess bool
	jobs      int
	inc       multiFlag
	ext       string
}

// Test implements `gperl test [options] [files or directories]`: it runs
// the test files and reports like prove. With no path it runs t/
// recursively with lib/ on @INC. The returned status is 0 when every file
// passed. `gperl test -c [-o out]` builds the project's test binary instead
// (RunTestBinary), which takes the options and paths when it runs.
func Test(args []string) (int, error) {
	c := &testConfig{}
	fl := c.flagSet("gperl test")
	compile := fl.Bool("c", false, "build the project's test binary instead of running the tests")
	out := fl.String("o", "", "with -c, the path of the test binary (default: <directory>.test)")
	if err := fl.Parse(splitShortOptions(args, "vqlr", "jI")); err != nil {
		return 2, err
	}
	building, others := false, fl.NArg() > 0
	fl.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "c", "o":
			building = true
		default:
			others = true
		}
	})
	if building {
		if others || !*compile {
			return 2, errors.New("usage: gperl test -c [-o out]; the test binary takes the options and paths when it runs")
		}
		return 0, buildBinary("", *out, nil)
	}
	files, err := c.files(fl.Args())
	if err != nil {
		return 2, err
	}
	projectDir, err := os.Getwd()
	if err != nil {
		return 2, err
	}
	deps, err := hostDeps(projectDir)
	if err != nil {
		return 2, err
	}
	return runTests(projectDir, files, c, deps.inc()), nil
}

// RunTestBinary is the entry point of a test binary gperl test -c makes: it
// runs the test files of the project at projectDir (the binary's own copy)
// that args name, with gperl test's options, and returns the exit status.
// The dependencies are the ones the binary carries.
func RunTestBinary(projectDir string, args []string) int {
	name := filepath.Base(os.Args[0])
	c := &testConfig{}
	fl := c.flagSet(name)
	if err := fl.Parse(splitShortOptions(args, "vqlr", "jI")); err != nil {
		// The flag set has said why.
		return 2
	}
	// Paths are the project's, as gperl test takes them in its directory.
	if err := os.Chdir(projectDir); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", name, err)
		return 1
	}
	files, err := c.files(fl.Args())
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", name, err)
		return 2
	}
	return runTests(projectDir, files, c, nil)
}

// flagSet declares the options of gperl test, which a test binary takes
// too, on a flag set called name.
func (c *testConfig) flagSet(name string) *flag.FlagSet {
	fl := flag.NewFlagSet(name, flag.ContinueOnError)
	for _, n := range []string{"v", "verbose"} {
		fl.BoolVar(&c.verbose, n, false, "print every test's output")
	}
	for _, n := range []string{"q", "quiet"} {
		fl.BoolVar(&c.quiet, n, false, "print only failures and the summary")
	}
	for _, n := range []string{"l", "lib"} {
		fl.BoolVar(&c.lib, n, false, "add lib/ to @INC")
	}
	for _, n := range []string{"r", "recurse"} {
		fl.BoolVar(&c.recurse, n, false, "descend into directories")
	}
	for _, n := range []string{"j", "jobs"} {
		fl.IntVar(&c.jobs, n, runtime.GOMAXPROCS(0), "number of test files to run at once")
	}
	fl.BoolVar(&c.timer, "timer", false, "print each file's elapsed time")
	fl.BoolVar(&c.inProcess, "in-process", false, "run the test files in this process, each as an isolated program")
	fl.Var(&c.inc, "I", "directory to add to @INC (repeatable)")
	fl.StringVar(&c.ext, "ext", ".t", "extension of test files")
	return fl
}

// files checks the options and returns the test files the paths name: t/,
// recursively with lib/ on @INC, when there is none.
func (c *testConfig) files(paths []string) ([]string, error) {
	if c.jobs < 1 {
		return nil, errors.New("--jobs must be at least 1")
	}
	if len(paths) == 0 {
		// The conventional invocation: prove -lr t.
		paths, c.lib, c.recurse = []string{"t"}, true, true
	}
	files, err := testFiles(paths, c)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no test files found in %s", strings.Join(paths, " "))
	}
	return files, nil
}

// runTests runs the test files of the project at projectDir, each with
// depInc (the dependencies) on @INC, reports like prove and returns the
// exit status.
func runTests(projectDir string, files []string, c *testConfig, depInc []string) int {
	if c.lib {
		c.inc.values = append(c.inc.values, "lib")
	}

	start := time.Now()
	outcomes := make([]*testOutcome, len(files))
	var (
		wg      sync.WaitGroup
		slots   = make(chan struct{}, c.jobs)
		printMu sync.Mutex
		printed int
	)
	width := 0
	for _, f := range files {
		width = max(width, len(f))
	}
	// Results print in file order, each as soon as every file before it
	// is done.
	report := func() {
		printMu.Lock()
		defer printMu.Unlock()
		for printed < len(outcomes) && outcomes[printed] != nil {
			printOutcome(os.Stdout, outcomes[printed], width, c)
			printed++
		}
	}
	for i, file := range files {
		wg.Add(1)
		slots <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-slots }()
			o := runTestFile(projectDir, file, c, depInc)
			printMu.Lock()
			outcomes[i] = o
			printMu.Unlock()
			report()
		}()
	}
	wg.Wait()
	return summarize(os.Stdout, outcomes, time.Since(start))
}

// testFiles expands the command line's paths into test files.
func testFiles(paths []string, c *testConfig) ([]string, error) {
	var files []string
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			files = append(files, p)
			continue
		}
		err = filepath.WalkDir(p, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if path != p && !c.recurse {
					return filepath.SkipDir
				}
				return nil
			}
			if strings.HasSuffix(path, c.ext) {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	sort.Strings(files)
	return files, nil
}

// runTestFile runs one test file on its own interpreter instance.
func runTestFile(projectDir, file string, c *testConfig, depInc []string) *testOutcome {
	o := &testOutcome{file: file}
	argv := make([]string, 0, len(c.inc.values)+len(depInc)+2)
	taint, err := taintSwitch(file)
	if err != nil {
		o.problem = err.Error()
		return o
	}
	if taint != "" {
		argv = append(argv, taint)
	}
	for _, dir := range append(c.inc.values[:len(c.inc.values):len(c.inc.values)], depInc...) {
		argv = append(argv, "-I"+dir)
	}
	argv = append(argv, file)
	var stdout bytes.Buffer
	env := append(os.Environ(), "HARNESS_ACTIVE=1", "HARNESS_VERSION=3.50")
	start := time.Now()
	var status int
	if !c.inProcess {
		status, err = runTestProcess(projectDir, env, &stdout, argv)
	} else {
		status, err = runPerlCLI(argv, perlRunOpts{
			dir: projectDir, env: env, stdout: &stdout, stderr: os.Stderr, isolated: true,
		})
	}
	o.duration = time.Since(start)
	o.status = status
	o.output = stdout.Bytes()
	tap, perr := parseTAP(bytes.NewReader(o.output), nil)
	o.tap = tap
	o.judge()
	switch {
	case perr != nil:
		o.problem = "unreadable TAP output: " + perr.Error()
	case err != nil && status == 255:
		// An uncaught die: perl prints the message and exits 255.
		PrintRunError(err)
	case err != nil:
		o.problem = err.Error()
	}
	return o
}

// tapTaint is TAP::Parser::SourceHandler::Perl's get_taint: the taint
// switches of a #! line that names perl.
var tapTaint = regexp.MustCompile(`^#!.*\bperl.*\s-\w*([Tt]+)`)

// taintSwitch returns the switch a test file's command line gets for the
// taint mode its #! line asks for ("-T", "-t"), "" for none, as
// TAP::Parser::SourceHandler::Perl gives it: the #! line is the first line
// of the file's first 512 bytes (TAP::Parser::Source's shebang), and perl
// refuses -T on that line when it is not on the command line too.
func taintSwitch(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	buf := make([]byte, 512)
	n, err := io.ReadFull(f, buf)
	if errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, io.EOF) {
		err = nil // a file shorter than the block
	}
	if err := errors.Join(err, f.Close()); err != nil {
		return "", err
	}
	line, _, _ := bytes.Cut(buf[:n], []byte("\n"))
	m := tapTaint.FindSubmatch(line)
	if m == nil {
		return "", nil
	}
	return "-" + string(m[1]), nil
}

// runTestProcess runs a test file as a perl process of its own: this
// executable under its name as the perl command.
func runTestProcess(dir string, env []string, stdout io.Writer, argv []string) (int, error) {
	exe := selfPerl()
	if exe == "" {
		return 1, errors.New("this executable has no path that runs it as the perl command")
	}
	cmd := exec.Command(exe, argv...)
	cmd.Dir, cmd.Env, cmd.Stdout, cmd.Stderr = dir, env, stdout, os.Stderr
	err := cmd.Run()
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		// The test's own status, which the report judges.
		return exit.ExitCode(), nil
	}
	return 0, err
}

// splitShortOptions rewrites the single-dash spellings prove and plackup
// accept but the flag package does not: a cluster of boolean options
// ("-lr" -> "-l", "-r") and an option with its value attached ("-j4" ->
// "-j", "4"; "-Ilib" -> "-I", "lib"). bools and valued list the
// single-letter options of each kind. Arguments after "--" are left alone.
func splitShortOptions(args []string, bools, valued string) []string {
	out := make([]string, 0, len(args))
	for i, a := range args {
		if a == "--" {
			return append(out, args[i:]...)
		}
		switch {
		case !strings.HasPrefix(a, "-"):
			// An option's value or a path: which one only the flag set
			// knows, and neither is rewritten.
			out = append(out, a)
		case len(a) > 2 && a[1] != '-' && strings.IndexByte(valued, a[1]) >= 0 && a[2] != '=':
			out = append(out, a[:2], a[2:])
		case len(a) > 2 && a[1] != '-' && strings.Trim(a[1:], bools) == "":
			for _, letter := range a[1:] {
				out = append(out, "-"+string(letter))
			}
		default:
			out = append(out, a)
		}
	}
	return out
}

// printOutcome writes a file's result line (and, in verbose mode, its
// output), in prove's shape.
func printOutcome(w io.Writer, o *testOutcome, width int, c *testConfig) {
	lead := o.file + " " + strings.Repeat(".", width-len(o.file)+2) + " "
	var b strings.Builder
	if c.verbose {
		b.WriteString(lead + "\n")
		b.Write(o.output)
		if len(o.output) > 0 && o.output[len(o.output)-1] != '\n' {
			b.WriteString("\n")
		}
		lead = ""
	}
	verdict := "ok"
	switch {
	case o.problem != "":
		verdict = o.problem
	case o.tap.planned == 0:
		verdict = "skipped: " + o.tap.skipAll
	}
	if c.timer {
		verdict += fmt.Sprintf(" %8d ms", o.duration.Milliseconds())
	}
	if !c.quiet || o.problem != "" {
		b.WriteString(lead + verdict + "\n")
	}
	// The line is progress output; a closed stdout only loses it.
	if _, err := io.WriteString(w, b.String()); err != nil {
		fmt.Fprintln(os.Stderr, "gperl test:", err)
	}
}

// summarize writes prove's closing report and returns the exit status.
func summarize(w io.Writer, outcomes []*testOutcome, elapsed time.Duration) int {
	var b strings.Builder
	tests, failedFiles := 0, 0
	var report strings.Builder
	for _, o := range outcomes {
		tests += o.tap.run
		if o.problem == "" && len(o.tap.todoOK) == 0 {
			continue
		}
		fmt.Fprintf(&report, "%s (Wstat: %d Tests: %d Failed: %d)\n", o.file, o.status<<8, o.tap.run, len(o.tap.failed))
		if len(o.tap.failed) > 0 {
			fmt.Fprintf(&report, "  Failed test%s:  %s\n", plural(len(o.tap.failed)), joinInts(o.tap.failed))
		}
		if len(o.tap.todoOK) > 0 {
			fmt.Fprintf(&report, "  TODO passed:   %s\n", joinInts(o.tap.todoOK))
		}
		if o.status != 0 {
			fmt.Fprintf(&report, "  Non-zero exit status: %d\n", o.status)
		}
		if o.problem != "" {
			failedFiles++
			if len(o.tap.failed) == 0 && o.status == 0 {
				fmt.Fprintf(&report, "  %s\n", o.problem)
			}
		}
	}
	if failedFiles == 0 {
		b.WriteString("All tests successful.\n")
	}
	if report.Len() > 0 {
		b.WriteString("\nTest Summary Report\n-------------------\n" + report.String())
	}
	fmt.Fprintf(&b, "Files=%d, Tests=%d, %2d wallclock secs\n", len(outcomes), tests, int(elapsed.Round(time.Second).Seconds()))
	status := 0
	if failedFiles > 0 {
		b.WriteString("Result: FAIL\n")
		status = 1
	} else {
		b.WriteString("Result: PASS\n")
	}
	if _, err := io.WriteString(w, b.String()); err != nil {
		fmt.Fprintln(os.Stderr, "gperl test:", err)
	}
	return status
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

func joinInts(ns []int) string {
	parts := make([]string, len(ns))
	for i, n := range ns {
		parts[i] = strconv.Itoa(n)
	}
	return strings.Join(parts, ", ")
}
