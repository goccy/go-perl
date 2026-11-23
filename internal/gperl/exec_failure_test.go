//go:build darwin || linux

package gperl_test

import (
	"strings"
	"testing"
)

// A command that cannot be started gets perl's warning, "Can't exec", when
// the exec warnings are on, from system (a string and a list) and a piped
// open alike, and system returns -1 with $! set: perl's forked child warns
// so when its exec fails.
func TestCommandThatCannotBeStartedWarns(t *testing.T) {
	const missing = "gperl-no-such-command"
	program := `
my $r = system("` + missing + `");
print "system: $r ", ($!{ENOENT} ? "ENOENT" : "other"), "\n";
$r = system("` + missing + `", "an argument");
print "system of a list: $r\n";
my $pid = open(my $fh, "-|", "` + missing + `");
print "piped open: ", (defined $pid ? "a pid" : "failed"), "\n";
`
	cmd := perlProcess("-we", "use Errno;"+program)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("perl: %v, stderr %s", err, stderr.String())
	}
	if want := "system: -1 ENOENT\nsystem of a list: -1\npiped open: failed\n"; string(out) != want {
		t.Errorf("output %q, want %q", out, want)
	}
	warning := `Can't exec "` + missing + `": No such file or directory at -e line `
	if n := strings.Count(stderr.String(), warning); n != 3 {
		t.Errorf("%d warnings %q in stderr, want 3:\n%s", n, warning, stderr.String())
	}
	// Without the warnings on, nothing is said.
	quiet := perlProcess("-e", "use Errno;"+program)
	var quietErr strings.Builder
	quiet.Stderr = &quietErr
	if _, err := quiet.Output(); err != nil || quietErr.Len() != 0 {
		t.Errorf("without warnings: err %v, stderr %q", err, quietErr.String())
	}
}

// "Can't exec" is said where perl says it. system, a piped open (of a
// string, of a list, for reading and for writing) and backticks run the
// command in a child perl forks, and it is the child whose exec fails: it
// warns in its copy of the program, with the command's standard handles,
// and ends. With no handler the warning goes to the child's stderr (into
// the pipe for a trailing 2>&1, where nobody reads it), a handler that
// records records into the child's copy (the parent's records nothing), a
// handler that prints prints to the child's stdout (into the pipe of
// backticks and of a piped open for reading), and under no warnings 'exec'
// nothing is said. A command with shell metacharacters is the shell's to
// report. exec warns in the process itself, which goes on. Backticks of a
// command with a trailing 2>&1 leave the parent's STDERR alone, a file
// here: the child's stderr is the pipe (Shell's $Shell::capture_stderr
// test). gperl said all
// of it in the parent, whose handler recorded seven (System::Info's tests
// failed Test::Warnings for an lsb_release that is not installed). What is
// expected is what perl 5.36 on Linux prints.
func TestCommandThatCannotBeStartedWarnsInTheChild(t *testing.T) {
	program := `open STDERR, '>&', \*STDOUT or die "dup: $!";
select STDERR; $| = 1; select STDOUT; $| = 1;
my $m = "gperl-no-such-command";
sub cases_warn {
    use warnings;
    my ($label) = @_;
    print "-- $label\n";
    my $x = ` + "`" + `$m one` + "`" + `;
    print "backticks: [", (defined $x ? $x : "undef"), "] status $?\n";
    my $r = system("$m two");
    print "system of a string: $r\n";
    $r = system($m, "three");
    print "system of a list: $r\n";
    my $ok = open(my $fh, "$m four |");
    print "piped open: ", ($ok ? "opened" : "failed"), "\n";
    close $fh if $ok;
    $ok = open($fh, "-|", $m, "five");
    print "piped open of a list: ", ($ok ? "opened" : "failed"), "\n";
    close $fh if $ok;
    $ok = open($fh, "$m six 2>&1 |");
    print "piped open with 2>&1: ", ($ok ? "opened" : "failed"), "\n";
    if ($ok) { my @l = <$fh>; print "read: @l"; close $fh }
    $ok = open($fh, "| $m seven");
    print "piped open for writing: ", ($ok ? "opened" : "failed"), "\n";
    close $fh if $ok;
    $r = system("$m eight 2>/dev/null; exit 3");
    print "system with shell metacharacters: ", $r >> 8, "\n";
    $x = ` + "`" + `$m nine 2>/dev/null || echo by the shell` + "`" + `;
    print "backticks with shell metacharacters: $x";
}
sub cases_quiet {
    use warnings;
    no warnings 'exec';
    my ($label) = @_;
    print "-- $label\n";
    my $x = ` + "`" + `$m one` + "`" + `;
    print "backticks: [", (defined $x ? $x : "undef"), "] status $?\n";
    my $r = system("$m two");
    print "system of a string: $r\n";
    $r = system($m, "three");
    print "system of a list: $r\n";
    my $ok = open(my $fh, "$m four |");
    print "piped open: ", ($ok ? "opened" : "failed"), "\n";
    close $fh if $ok;
}
cases_warn("no handler");
{
    my @w;
    local $SIG{__WARN__} = sub { push @w, $_[0] };
    cases_warn("a recording handler");
    print "recorded: ", scalar(@w), "\n";
}
{
    local $SIG{__WARN__} = sub { print STDOUT "handler: $_[0]" };
    cases_warn("a printing handler");
}
cases_quiet("no warnings 'exec'");
{
    use warnings;
    my @w;
    local $SIG{__WARN__} = sub { push @w, $_[0] };
    my $r = exec("$m ten");
    print "exec: ", ($r ? "true" : "false"), ", recorded ", scalar(@w), ": ", join("", @w);
}
{
    use warnings;
    open my $save, '>&', \*STDERR or die "dup: $!";
    my $f = "/tmp/gperl-exec-warn-$$";
    open STDERR, '>', $f or die "open $f: $!";
    my $x = ` + "`" + `$m eleven 2>&1` + "`" + `;
    open STDERR, '>&', $save or die "dup back: $!";
    my $size = -s $f;
    unlink $f;
    print "backticks with 2>&1, stderr in a file: [", (defined $x ? $x : "undef"), "] the file ",
        ($size ? "has $size bytes" : "is empty"), "\n";
}
print "end\n";`
	out, status := hostPerl(t, "-e", program)
	want := `-- no handler
Can't exec "gperl-no-such-command": No such file or directory at -e line 8.
backticks: [undef] status -1
Can't exec "gperl-no-such-command": No such file or directory at -e line 10.
system of a string: -1
Can't exec "gperl-no-such-command": No such file or directory at -e line 12.
system of a list: -1
Can't exec "gperl-no-such-command": No such file or directory at -e line 14.
piped open: failed
Can't exec "gperl-no-such-command": No such file or directory at -e line 17.
piped open of a list: failed
piped open with 2>&1: failed
Can't exec "gperl-no-such-command": No such file or directory at -e line 23.
piped open for writing: failed
system with shell metacharacters: 3
backticks with shell metacharacters: by the shell
-- a recording handler
backticks: [undef] status -1
system of a string: -1
system of a list: -1
piped open: failed
piped open of a list: failed
piped open with 2>&1: failed
piped open for writing: failed
system with shell metacharacters: 3
backticks with shell metacharacters: by the shell
recorded: 0
-- a printing handler
backticks: [undef] status -1
handler: Can't exec "gperl-no-such-command": No such file or directory at -e line 10.
system of a string: -1
handler: Can't exec "gperl-no-such-command": No such file or directory at -e line 12.
system of a list: -1
piped open: failed
piped open of a list: failed
piped open with 2>&1: failed
handler: Can't exec "gperl-no-such-command": No such file or directory at -e line 23.
piped open for writing: failed
system with shell metacharacters: 3
backticks with shell metacharacters: by the shell
-- no warnings 'exec'
backticks: [undef] status -1
system of a string: -1
system of a list: -1
piped open: failed
exec: false, recorded 1: Can't exec "gperl-no-such-command": No such file or directory at -e line 62.
backticks with 2>&1, stderr in a file: [undef] the file is empty
end
`
	if status != 0 || out != want {
		t.Fatalf("status %d, output:\n%s\nwant:\n%s", status, out, want)
	}
}

// A forked process's /dev/stdout is its own standard output: the child of
// open($fh, "-|") that writes to /dev/stdout writes into the pipe its
// parent reads, as on perl.
func TestForkedProcessOpensItsOwnStandardOutput(t *testing.T) {
	out, status := hostPerl(t, "-e", `
my $pid = open(my $in, "-|") // die "fork: $!";
if (!$pid) {
    open(my $o, ">", "/dev/stdout") or die "open /dev/stdout: $!";
    print $o "through /dev/stdout\n";
    close $o;
    exit 0;
}
my @got = <$in>;
close $in;
print "the parent read: @got";
`)
	if want := "the parent read: through /dev/stdout\n"; status != 0 || out != want {
		t.Fatalf("status %d, output %q, want %q", status, out, want)
	}
}
