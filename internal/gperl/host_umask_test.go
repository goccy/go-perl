//go:build darwin || linux

package gperl_test

import (
	"syscall"
	"testing"
)

// The host process follows the program's file mode creation mask, as it
// follows its working directory: native code that creates a file for the
// program (DB_File's database, say) and the commands it starts get the mask
// the program set. A forked process's mask stays its own.
func TestHostProcessFollowsTheProgramsUmask(t *testing.T) {
	start := syscall.Umask(0o022)
	defer syscall.Umask(start)
	got := runForkScriptArgs(t, `
use POSIX ();
my $pid = fork // die "fork: $!";
if (!$pid) {
    umask 007;
    POSIX::_exit(0);
}
waitpid($pid, 0);
umask 070;
my $seen = `+"`sh -c umask`"+`;
chomp $seen;
report("a command started after umask 070 has umask ", oct($seen) == 070 ? "070" : $seen);
`)
	wantLines(t, got, "a command started after umask 070 has umask 070")
	if now := syscall.Umask(0o022); now != 0o070 {
		t.Errorf("umask of the host process = %o, want 070", now)
	}
}
