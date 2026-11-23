//go:build darwin || linux

package gperl_test

import "testing"

// gethostbyname leaves no descriptor behind that a command the program
// starts afterwards would inherit, as perl's leaves none: on darwin the
// runtime's resolver (getaddrinfo) keeps a socket to the system open
// without close-on-exec, which gethostbyname must not ask. The lookup goes
// through the host's gethostbyname where the host has one.
func TestGethostbynameLeavesNothingToCommands(t *testing.T) {
	out, status := hostPerl(t, "-e", `
my $before = qx{ls /dev/fd};
my @host = gethostbyname("localhost") or die "localhost: not found";
my $after = qx{ls /dev/fd};
print $before eq $after ? "nothing new\n" : "before:\n$before" . "after:\n$after";
`)
	if status != 0 || out != "nothing new\n" {
		t.Fatalf("status %d, output %q", status, out)
	}
}
