//go:build darwin || linux

package gperl_test

import "testing"

// dbmopen creates the database it is asked to open when it is not there,
// and reads it back once closed, as on perl: the open flags it ties with
// are the numbers the program sees.
func TestDbmopenCreatesTheDatabase(t *testing.T) {
	out, status := hostPerl(t, "-MFile::Temp=tempdir", "-e", `
my $dir = tempdir(CLEANUP => 1);
my %seen;
print "dbmopen: ", (dbmopen(%seen, "$dir/seen", 0644) ? "opened" : "failed: $!"), "\n";
$seen{first} = "one";
dbmclose(%seen);
my %again;
dbmopen(%again, "$dir/seen", 0644) or die "dbmopen again: $!";
print "read back: $again{first}\n";
dbmclose(%again);
`)
	if want := "dbmopen: opened\nread back: one\n"; status != 0 || out != want {
		t.Fatalf("status %d, output %q, want %q", status, out, want)
	}
}
