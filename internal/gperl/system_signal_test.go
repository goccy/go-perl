//go:build darwin || linux

package gperl_test

import "testing"

// A signal handler that dies while system waits for its command (an alarm
// that times the command out) leaves system then, as perl's interrupted
// wait runs the handler: system does not wait for the command to end, and
// the die is what the program sees.
func TestSystemLeftByASignalHandlerThatDies(t *testing.T) {
	out, status := hostPerl(t, "-MTime::HiRes=time", "-e", `
$| = 1;
for my $cmd (['sleep', '5'], ['sleep 5']) {
    my $t0 = time;
    my $rv;
    eval {
        local $SIG{ALRM} = sub { die "timeout\n" };
        alarm 1;
        $rv = system(@$cmd);
        alarm 0;
    };
    alarm 0;
    my $took = time - $t0;
    print scalar(@$cmd), " argument(s): ", ($@ eq "timeout\n" ? "the handler's die" : "no die ($@)"),
        ", ", (defined $rv ? "system returned $rv" : "system did not return"),
        ", ", ($took < 3 ? "at the alarm" : "after ${took}s"), "\n";
}
`)
	want := "2 argument(s): the handler's die, system did not return, at the alarm\n" +
		"1 argument(s): the handler's die, system did not return, at the alarm\n"
	if status != 0 || out != want {
		t.Fatalf("status %d, output:\n%s\nwant:\n%s", status, out, want)
	}
}
