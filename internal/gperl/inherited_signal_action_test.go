//go:build linux

package gperl_test

import (
	"os"
	"os/exec"
	"testing"
)

// A signal the process ignores by an action it did not get from the
// program (here, inherited: started under nohup, or by a shell's trap with
// an empty action) reads as ignored in the program, as on perl: $SIG{HUP}
// is IGNORE and POSIX::sigaction reports it, and the signal does nothing to
// the program.
// Native code that ignores a signal (a network library ignores SIGPIPE)
// leaves the program the same action.
func TestProgramSeesTheIgnoredActionOfTheProcess(t *testing.T) {
	cmd := exec.Command("/bin/sh", "-c", `trap '' HUP; exec "$0" "$@"`, os.Args[0], "-MPOSIX", "-e", `
print "SIG{HUP}: ", ($SIG{HUP} // "undef"), "\n";
my $action = POSIX::SigAction->new;
POSIX::sigaction(POSIX::SIGHUP(), undef, $action);
print "sigaction: $action->{HANDLER}\n";
kill HUP => $$;
print "after kill: alive\n";
print "SIG{USR1}: ", ($SIG{USR1} // "undef"), "\n";
`)
	cmd.Env = append(os.Environ(), asPerlEnv+"=1")
	out, err := cmd.CombinedOutput()
	want := "SIG{HUP}: IGNORE\nsigaction: IGNORE\nafter kill: alive\nSIG{USR1}: undef\n"
	if err != nil || string(out) != want {
		t.Fatalf("err %v, output:\n%s\nwant:\n%s", err, out, want)
	}
}
