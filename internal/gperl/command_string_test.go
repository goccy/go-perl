//go:build darwin || linux

package gperl_test

import (
	"os"
	"path/filepath"
	"testing"
)

// A command given as one string runs directly when it has no shell
// metacharacters and through sh -c otherwise — perl's rule. The difference
// shows when the command does not exist: run directly, that is the
// caller's error; run by the shell, it is the shell's exit status 127.
func TestCommandStringRunsDirectlyWithoutMetacharacters(t *testing.T) {
	dir := t.TempDir()
	// executable, but not a program: no interpreter line
	plain := filepath.Join(dir, "plainscript")
	if err := os.WriteFile(plain, []byte("echo ran-by-the-shell\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))

	got := runForkScript(t, `
sub outcome { my ($r) = @_; $r == -1 ? "-1 errno=" . ($!{ENOENT} ? "ENOENT" : "$!") : "exit=" . ($r >> 8) }
sub capture { my ($cmd) = @_; my $out = `+"`$cmd`"+`; defined $out ? $out =~ s/\n/|/gr : "undef" }

report("missing command: ", outcome(system("no-such-command-anywhere")));
report("missing command with arguments: ", outcome(system("no-such-command-anywhere a b")));
report("missing command, shell syntax: ", outcome(system("no-such-command-anywhere 2>/dev/null; exit 9")));
report("missing command in the shell: ", outcome(system("no-such-command-anywhere 2>/dev/null")));

report("words: ", capture("echo  one   two"));
report("leading space and trailing newline: ", capture("   echo trimmed\n"));
report("quotes need the shell: ", capture(q{echo 'one  two'}));
report("variable needs the shell: ", capture(q{V=set; echo $V}));
report("assignment prefix needs the shell: ", capture(q{V=prefix sh -c 'echo $V'}));
report("exec prefix goes to the shell: ", capture("exec echo replaced"));
report("stderr joined without a shell: ", capture(q{sh -c "echo err >&2; echo out" 2>&1}) =~ s/out\|err/err|out/r);
report("2>&1 of a missing command: ", outcome(system("no-such-command-anywhere 2>&1")));
report("no interpreter line: ", capture("plainscript"));
report("empty command: ", outcome(system("")), " and blank: ", outcome(system("   ")));

report("piped open of a missing command: ", (open(my $fh, "no-such-command-anywhere |") ? "opened" : ($!{ENOENT} ? "ENOENT" : "$!")));
report("piped open through the shell: ", (open($fh, "no-such-command-anywhere 2>/dev/null |") ? "opened, close=" . (close($fh) ? "ok" : "exit " . ($? >> 8)) : "failed"));
open($fh, "echo from a pipe |") or die "open: $!";
report("piped open: ", scalar(<$fh>) =~ s/\n//r);
close $fh;
report("backticks of a missing command: ", capture("no-such-command-anywhere"), " status=$?");

my $pid = fork // die "fork: $!";
if ($pid == 0) { exec("no-such-command-anywhere") or report("exec of a missing command: ", ($!{ENOENT} ? "ENOENT" : "$!")); exit 7 }
waitpid($pid, 0);
report("the child went on: exit=", $? >> 8);
$pid = fork // die "fork: $!";
if ($pid == 0) { exec("sh -c 'exit 5'"); exit 7 }
waitpid($pid, 0);
report("exec of a command string: exit=", $? >> 8);
`)
	wantLines(t, got,
		"missing command: -1 errno=ENOENT",
		"missing command with arguments: -1 errno=ENOENT",
		"missing command, shell syntax: exit=9",
		"missing command in the shell: exit=127",
		"words: one two|",
		"leading space and trailing newline: trimmed|",
		"quotes need the shell: one  two|",
		"variable needs the shell: set|",
		"assignment prefix needs the shell: prefix|",
		"exec prefix goes to the shell: replaced|",
		"stderr joined without a shell: err|out|",
		"2>&1 of a missing command: -1 errno=ENOENT",
		"no interpreter line: ran-by-the-shell|",
		"empty command: exit=255 and blank: exit=255",
		"piped open of a missing command: ENOENT",
		"piped open through the shell: opened, close=exit 127",
		"piped open: from a pipe",
		"backticks of a missing command: undef status=-1",
		"exec of a missing command: ENOENT",
		"the child went on: exit=7",
		"exec of a command string: exit=5",
	)
}
