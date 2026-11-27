//go:build linux

package gperl_test

import (
	"os"
	"os/user"
	"strings"
	"testing"
)

// getpwent and getgrent walk the host's user and group databases, as perl
// does: the first entry is the first line of /etc/passwd (root), the user
// the process runs as is among them, and the walk starts over after
// setpwent.
func TestUserDatabaseWalk(t *testing.T) {
	data, err := os.ReadFile("/etc/passwd")
	if err != nil {
		t.Skip("no /etc/passwd")
	}
	first := strings.SplitN(strings.SplitN(string(data), "\n", 2)[0], ":", 2)[0]
	me, err := user.Current()
	if err != nil {
		t.Fatal(err)
	}
	out, status := hostPerl(t, "-e", `
my @names;
while (my @pw = getpwent) { push @names, $pw[0] }
endpwent;
print "first user: $names[0]\n";
my $me = getpwuid($<);
print "the process's user: ", ((grep { $_ eq $me } @names) ? "listed" : "missing"), "\n";
setpwent;
my @again = getpwent;
print "after setpwent: $again[0]\n";
endpwent;
my @groups;
while (my @gr = getgrent) { push @groups, $gr[0] }
endgrent;
print "first group: $groups[0]\n";
my $mine = getgrgid($();
print "the process's group: ", ((grep { $_ eq $mine } @groups) ? "listed" : "missing"), "\n";
`)
	groups, _ := os.ReadFile("/etc/group")
	firstGroup := strings.SplitN(strings.SplitN(string(groups), "\n", 2)[0], ":", 2)[0]
	want := "first user: " + first + "\n" +
		"the process's user: listed\n" +
		"after setpwent: " + first + "\n" +
		"first group: " + firstGroup + "\n" +
		"the process's group: listed\n"
	if status != 0 || out != want {
		t.Fatalf("status %d (user %s), output:\n%s\nwant:\n%s", status, me.Username, out, want)
	}
}
