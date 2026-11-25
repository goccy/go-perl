package gperl

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestParseTAP(t *testing.T) {
	src := strings.Join([]string{
		"1..7",
		"ok 1 - first",
		"not ok 2 - broken",
		"# a diagnostic",
		"    ok 1 - inside a subtest",
		"    not ok 2 - inside a subtest",
		"ok 3 - subtest",
		"ok 4 # skip no network",
		"not ok 5 - known # TODO not yet",
		"ok 6 - surprise # TODO thought broken",
		"not ok - unnumbered",
		"something that is not TAP",
	}, "\n")
	var seen int
	res, err := parseTAP(strings.NewReader(src), func(string) { seen++ })
	if err != nil {
		t.Fatal(err)
	}
	want := tapResult{planned: 7, run: 7, failed: []int{2, 7}, todoOK: []int{6}, skippedN: 1}
	if !reflect.DeepEqual(res, want) {
		t.Fatalf("parseTAP = %+v\nwant       %+v", res, want)
	}
	if seen != 12 {
		t.Fatalf("the line callback ran %d times", seen)
	}

	// A # that starts no directive starts a comment: the line is a test
	// line all the same.
	res, err = parseTAP(strings.NewReader("1..5\nok 1 # foo\nnot ok 2 # bar\nok 3 # SKIP gone\nnot ok 4 # TODO later\nok 6 # 7\n"), nil)
	if want := (tapResult{planned: 5, run: 5, failed: []int{2}, skippedN: 1}); err != nil || !reflect.DeepEqual(res, want) {
		t.Fatalf("comments = %+v (%v)\nwant       %+v", res, err, want)
	}

	res, err = parseTAP(strings.NewReader("1..0 # SKIP not here\n"), nil)
	if err != nil || res.planned != 0 || res.skipAll != "not here" {
		t.Fatalf("skip-all plan = %+v, %v", res, err)
	}
	res, err = parseTAP(strings.NewReader("ok 1\nBail out! database is down\nnot ok 2\n"), nil)
	if err != nil || !res.hasBailed || res.bailout != "database is down" || res.run != 1 {
		t.Fatalf("bail out = %+v, %v", res, err)
	}
	res, err = parseTAP(strings.NewReader("ok 1\n"), nil)
	if err != nil || res.planned != -1 {
		t.Fatalf("no plan = %+v, %v", res, err)
	}
}

// A test file whose #! line asks for taint mode runs in it, as prove runs
// it: the switch goes on its command line too, which perl requires.
func TestTestFileTaintSwitch(t *testing.T) {
	dir := t.TempDir()
	for name, src := range map[string]string{
		"t/taint.t": "#!perl -T\nuse Test::More tests => 1;\nis(${^TAINT}, 1, 'taint mode');\n",
		"t/warn.t":  "#!/usr/bin/perl -wt\nuse Test::More tests => 1;\nis(${^TAINT}, -1, 'taint warnings');\n",
		"t/plain.t": "#!perl -w\nuse Test::More tests => 1;\nis(${^TAINT}, 0, 'no taint');\n",
	} {
		p := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Chdir(dir)
	t.Setenv("GPERLMODCACHE", filepath.Join(t.TempDir(), "mod"))
	var status int
	out := captureStdout(t, func() {
		var err error
		if status, err = Test(nil); err != nil {
			t.Errorf("Test: %v", err)
		}
	})
	if status != 0 || !strings.Contains(out, "All tests successful.\n") || !strings.Contains(out, "Files=3, Tests=3, ") {
		t.Errorf("status %d:\n%s", status, out)
	}
}

func TestJudge(t *testing.T) {
	for want, o := range map[string]testOutcome{
		"":                            {tap: tapResult{planned: 2, run: 2}},
		"Failed 1/3 subtests":         {tap: tapResult{planned: 3, run: 3, failed: []int{2}}},
		"No plan found in TAP output": {tap: tapResult{planned: -1, run: 1}},
		"Parse errors: Bad plan.  You planned 2 tests but ran 1.": {tap: tapResult{planned: 2, run: 1}},
		"Dubious, test returned 3":                                {tap: tapResult{planned: 1, run: 1}, status: 3},
		"Bailed out: stop":                                        {tap: tapResult{planned: 1, run: 1, hasBailed: true, bailout: "stop"}},
	} {
		o.judge()
		if o.problem != want {
			t.Errorf("judge(%+v) = %q, want %q", o.tap, o.problem, want)
		}
	}
}

func TestSplitShortOptions(t *testing.T) {
	for _, c := range []struct{ in, want string }{
		{"-lr t", "-l -r t"},
		{"-lrv -j4 -Ilib t/a.t", "-l -r -v -j 4 -I lib t/a.t"},
		{"-j 4 --verbose", "-j 4 --verbose"},
		{"-j=4 -I=lib", "-j=4 -I=lib"},
		{"-lx t", "-lx t"},
		{"--jobs 4 -lr t", "--jobs 4 -l -r t"},
		{"-l -- -rv", "-l -- -rv"},
	} {
		got := strings.Join(splitShortOptions(strings.Fields(c.in), "vqlr", "jI"), " ")
		if got != c.want {
			t.Errorf("splitShortOptions(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestReportShape(t *testing.T) {
	outcomes := []*testOutcome{
		{file: "t/a.t", tap: tapResult{planned: 2, run: 2}, output: []byte("1..2\nok 1\nok 2")},
		{file: "t/long-name.t", tap: tapResult{planned: 3, run: 3, failed: []int{2, 3}}, status: 2},
		{file: "t/skip.t", tap: tapResult{planned: 0, skipAll: "no db"}},
		{file: "t/todo.t", tap: tapResult{planned: 1, run: 1, todoOK: []int{1}}},
		{file: "t/plan.t", tap: tapResult{planned: -1, run: 1}},
	}
	var b bytes.Buffer
	for _, o := range outcomes {
		o.judge()
		printOutcome(&b, o, len("t/long-name.t"), &testConfig{})
	}
	status := summarize(&b, outcomes, 1400*time.Millisecond)
	want := `t/a.t .......... ok
t/long-name.t .. Failed 2/3 subtests
t/skip.t ....... skipped: no db
t/todo.t ....... ok
t/plan.t ....... No plan found in TAP output

Test Summary Report
-------------------
t/long-name.t (Wstat: 512 Tests: 3 Failed: 2)
  Failed tests:  2, 3
  Non-zero exit status: 2
t/todo.t (Wstat: 0 Tests: 1 Failed: 0)
  TODO passed:   1
t/plan.t (Wstat: 0 Tests: 1 Failed: 0)
  No plan found in TAP output
Files=5, Tests=7,  1 wallclock secs
Result: FAIL
`
	if b.String() != want || status != 1 {
		t.Fatalf("status %d, report:\n%s\nwant:\n%s", status, b.String(), want)
	}

	b.Reset()
	printOutcome(&b, outcomes[0], 5, &testConfig{verbose: true, timer: true})
	if got := b.String(); !strings.HasPrefix(got, "t/a.t .. \n1..2\nok 1\nok 2\nok ") || !strings.HasSuffix(got, " ms\n") {
		t.Fatalf("verbose line = %q", got)
	}
	b.Reset()
	printOutcome(&b, outcomes[0], 13, &testConfig{quiet: true})
	printOutcome(&b, outcomes[1], 13, &testConfig{quiet: true})
	if got := b.String(); strings.Contains(got, "t/a.t") || !strings.Contains(got, "t/long-name.t") {
		t.Fatalf("quiet output = %q", got)
	}
	b.Reset()
	if status := summarize(&b, outcomes[:1], 0); status != 0 || !strings.Contains(b.String(), "All tests successful.\nFiles=1, Tests=2") || !strings.HasSuffix(b.String(), "Result: PASS\n") {
		t.Fatalf("passing summary (status %d):\n%s", status, b.String())
	}
}

// A project's test files run each as a program of its own,
// and the report follows prove.
func TestTestCommand(t *testing.T) {
	dir := t.TempDir()
	for name, src := range map[string]string{
		"lib/My/Calc.pm":   "package My::Calc;\nsub add { $_[0] + $_[1] }\n1;\n",
		"t/01-pass.t":      "use Test::More tests => 2;\nuse My::Calc;\nis(My::Calc::add(40, 2), 42, 'adds');\nok(1);\n",
		"t/02-fail.t":      "use Test::More tests => 2;\nok(1);\nok(0, 'broken');\n",
		"t/03-skip.t":      "use Test::More skip_all => 'elsewhere';\n",
		"t/deep/04-more.t": "use Test::More tests => 1;\nok(1, \"cwd is the project: \" . (-d 'lib' ? 1 : 0));\n",
		"t/notes.txt":      "not a test",
	} {
		p := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Chdir(dir)
	t.Setenv("GPERLMODCACHE", filepath.Join(t.TempDir(), "mod"))

	run := func(args ...string) (string, int) {
		t.Helper()
		var status int
		out := captureStdout(t, func() {
			var err error
			if status, err = Test(args); err != nil {
				t.Errorf("Test(%v): %v", args, err)
			}
		})
		return out, status
	}
	out, status := run()
	for _, want := range []string{
		"t/01-pass.t ....... ok\n",
		"t/02-fail.t ....... Failed 1/2 subtests\n",
		"t/03-skip.t ....... skipped: elsewhere\n",
		"t/deep/04-more.t .. ok\n",
		"t/02-fail.t (Wstat: 256 Tests: 2 Failed: 1)\n  Failed test:  2\n  Non-zero exit status: 1\n",
		"Files=4, Tests=5, ",
		"Result: FAIL\n",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("report lacks %q:\n%s", want, out)
		}
	}
	if status != 1 {
		t.Errorf("status = %d, want 1", status)
	}
	// In one process, each file an isolated program, the report is the same.
	if processOut, processStatus := run("--in-process"); processStatus != 1 ||
		!strings.Contains(processOut, "t/02-fail.t (Wstat: 256 Tests: 2 Failed: 1)\n") ||
		!strings.Contains(processOut, "Files=4, Tests=5, ") {
		t.Errorf("--in-process run (status %d):\n%s", processStatus, processOut)
	}

	out, status = run("-lv", "-j1", "t/01-pass.t")
	if status != 0 || !strings.Contains(out, "ok 1 - adds\n") || !strings.Contains(out, "All tests successful.\n") {
		t.Errorf("one passing file, verbose (status %d):\n%s", status, out)
	}
	// Without -r only the directory's own files run.
	out, status = run("-l", "t")
	if strings.Contains(out, "04-more") || !strings.Contains(out, "Files=3, ") {
		t.Errorf("non-recursive run (status %d):\n%s", status, out)
	}
	for _, bad := range [][]string{{"--no-such-option"}, {"-j", "0"}, {"missing-dir"}, {"lib"}} {
		if _, err := Test(bad); err == nil {
			t.Errorf("Test(%v) succeeded", bad)
		}
	}
}
