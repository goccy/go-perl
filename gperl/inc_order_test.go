//go:build darwin || linux

package gperl_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	perl "github.com/goccy/go-perl"
	"github.com/goccy/go-perl/gperl"
)

// shadowLib creates a library directory holding its own copy of the core
// module Text::Tabs, which reports where it was loaded from, and returns
// the directory.
func shadowLib(t *testing.T, name string) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), name)
	if err := os.MkdirAll(filepath.Join(dir, "Text"), 0o755); err != nil {
		t.Fatal(err)
	}
	src := "package Text::Tabs;\nour $FROM = '" + name + "';\nsub import { $main::IMPORTS{'" + name + "'}++ }\n1;\n"
	if err := os.WriteFile(filepath.Join(dir, "Text", "Tabs.pm"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// incReport is a program that reports @INC and where modules come from.
const incReport = `
my ($stdlib) = grep { -e "$_/strict.pm" } @INC;
report("inc: ", join(" ", map { $_ eq $stdlib ? "STDLIB" : $_ } @INC));
require Text::Tabs;
report("Text::Tabs from ", $Text::Tabs::FROM // "the standard library");
require List::Util;
report("List::Util from ", ($INC{"List/Util.pm"} =~ /^\Q$stdlib\E/ ? "the standard library" : $INC{"List/Util.pm"}));
`

// perl searches the -I directories, then PERL5LIB, then its own library:
// a module in a PERL5LIB directory shadows the standard library's copy (this
// is how `make test` hands blib/lib to the tests), and the directories of
// the build-time configuration are not searched at all.
func TestIncOrderPERL5LIBBeforeTheStandardLibrary(t *testing.T) {
	first, second := shadowLib(t, "first"), shadowLib(t, "second")
	t.Setenv("PERL5LIB", first+string(os.PathListSeparator)+second)
	got := runForkScript(t, incReport+`
report("configured directories searched: ", scalar(grep { m{^/usr/lib/perl5} } @INC));
my $pid = fork // die "fork: $!";
if ($pid == 0) { report("fork child inc: ", join(" ", map { $_ eq $stdlib ? "STDLIB" : $_ } @INC)); exit 0 }
waitpid($pid, 0);
`)
	// Run is the project runner: the script's directory leads, as a -I
	// directory would.
	scriptDir := strings.Fields(strings.TrimPrefix(got[0], "inc: "))[0]
	inc := strings.Join([]string{scriptDir, first, second, "STDLIB", "."}, " ")
	wantLines(t, got,
		"inc: "+inc,
		"Text::Tabs from first",
		"List::Util from the standard library",
		"configured directories searched: 0",
		"fork child inc: "+inc,
	)
}

// Without PERL5LIB the standard library's own copy is used.
func TestIncOrderStandardLibraryAlone(t *testing.T) {
	t.Setenv("PERL5LIB", "")
	got := runForkScript(t, incReport)
	if len(got) != 3 || !strings.HasSuffix(got[0], " STDLIB .") {
		t.Fatalf("report = %q", got)
	}
	wantLines(t, got[1:], "Text::Tabs from the standard library", "List::Util from the standard library")
}

// runPerl runs a perl command line in a gperl process of its own and
// returns what it printed.
func runPerl(t *testing.T, args ...string) string {
	t.Helper()
	out := filepath.Join(t.TempDir(), "out.txt")
	shim := perlCommandShim(t)
	script := writeScript(t, "spawn.pl", `
my ($report, $perl, @args) = @ARGV;
open(STDOUT, '>', $report) or die "redirect: $!";
exit(system($perl, @args) >> 8);
`)
	if status, err := gperl.Run(script, append([]string{out, shim}, args...)); err != nil || status != 0 {
		t.Fatalf("perl %q = (%d, %v)", args, status, err)
	}
	raw, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

// The perl command line: -I comes before PERL5LIB, use lib before both, a
// -M module named in PERL5OPT is loaded once and from where @INC says, and
// under -T neither PERL5LIB nor "." is searched.
func TestIncOrderOnThePerlCommandLine(t *testing.T) {
	first, second := shadowLib(t, "first"), shadowLib(t, "second")
	t.Setenv("PERL5LIB", first)
	const show = `my ($s) = grep { -e "$_/strict.pm" } @INC; print join(" ", map { $_ eq $s ? "STDLIB" : $_ } @INC), "\n";`
	const from = `require Text::Tabs; print "from ", ($Text::Tabs::FROM // "stdlib"), "\n";`

	if got, want := runPerl(t, "-e", show+from), first+" STDLIB .\nfrom first\n"; got != want {
		t.Errorf("PERL5LIB only:\n%swant:\n%s", got, want)
	}
	if got, want := runPerl(t, "-I"+second, "-e", show+from), second+" "+first+" STDLIB .\nfrom second\n"; got != want {
		t.Errorf("-I and PERL5LIB:\n%swant:\n%s", got, want)
	}
	if got, want := runPerl(t, "-e", "use lib '"+second+"';"+show+from), second+" "+first+" STDLIB .\nfrom second\n"; got != want {
		t.Errorf("use lib:\n%swant:\n%s", got, want)
	}
	if got, want := runPerl(t, "-T", "-e", show+from), "STDLIB\nfrom stdlib\n"; got != want {
		t.Errorf("-T:\n%swant:\n%s", got, want)
	}
	if got, want := runPerl(t, "-T", "-I"+second, "-e", show+from), second+" STDLIB\nfrom second\n"; got != want {
		t.Errorf("-T -I:\n%swant:\n%s", got, want)
	}

	t.Setenv("PERL5OPT", "-MText::Tabs")
	const loaded = `print "from $Text::Tabs::FROM, imported ", join(",", map { "$_=$main::IMPORTS{$_}" } sort keys %main::IMPORTS), "\n";` +
		`print "PERL5OPT=$ENV{PERL5OPT}\n";`
	if got, want := runPerl(t, "-e", loaded), "from first, imported first=1\nPERL5OPT=-MText::Tabs\n"; got != want {
		t.Errorf("PERL5OPT:\n%swant:\n%s", got, want)
	}
	if got, want := runPerl(t, "-I"+second, "-e", loaded), "from second, imported second=1\nPERL5OPT=-MText::Tabs\n"; got != want {
		t.Errorf("PERL5OPT with -I:\n%swant:\n%s", got, want)
	}
}

// An embedded interpreter has the same order: the directories it is given
// (AddInc, RunFile's, Program.Inc) lead, PERL5LIB from its environment
// follows, the standard library comes last — also in an instance cloned
// from another.
func TestIncOrderOfAnEmbeddedInterpreter(t *testing.T) {
	first, added := shadowLib(t, "first"), shadowLib(t, "added")
	cfg, err := gperl.HostConfig()
	if err != nil {
		t.Fatal(err)
	}
	cfg.Env = []string{"PERL5LIB=" + first}
	p, err := perl.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	const show = `my ($s) = grep { -e "$_/strict.pm" } @INC; join(" ", map { $_ eq $s ? "STDLIB" : $_ } @INC)`
	eval := func(p *perl.Perl) string {
		t.Helper()
		r, err := p.Eval(t.Context(), show)
		if err != nil || r.Error != nil {
			t.Fatalf("Eval: %v %v", err, r.Error)
		}
		s, err := perl.As[perl.ScalarValue](r.Value)
		if err != nil {
			t.Fatal(err)
		}
		return s.String()
	}
	if got, want := eval(p), first+" STDLIB ."; got != want {
		t.Fatalf("@INC = %q, want %q", got, want)
	}
	if err := p.AddInc(t.Context(), added); err != nil {
		t.Fatal(err)
	}
	if got, want := eval(p), added+" "+first+" STDLIB ."; got != want {
		t.Fatalf("@INC after AddInc = %q, want %q", got, want)
	}
	clone, err := p.Clone()
	if err != nil {
		t.Fatal(err)
	}
	defer clone.Close()
	if got, want := eval(clone), added+" "+first+" STDLIB ."; got != want {
		t.Fatalf("@INC of the clone = %q, want %q", got, want)
	}
}
