//go:build darwin || linux

package gperl_test

import "testing"

// Fcntl has the old names of flags of open that perl on Linux has, for the
// same flags: O_NDELAY, FNDELAY and FNONBLOCK are O_NONBLOCK, FAPPEND is
// O_APPEND (AnyEvent::SerialPort opens its device with O_NDELAY), and a
// file opened with O_NDELAY is in non-blocking mode.
func TestFcntlHasTheOldNamesOfOpenFlags(t *testing.T) {
	out, status := hostPerl(t, "-MFcntl=:DEFAULT,:mode,F_GETFL,O_NDELAY,FNDELAY,FNONBLOCK,FAPPEND", "-e", `
print "O_NDELAY: ", (O_NDELAY == O_NONBLOCK ? "O_NONBLOCK" : O_NDELAY), "\n";
print "FNDELAY: ", (FNDELAY == O_NONBLOCK ? "O_NONBLOCK" : FNDELAY), "\n";
print "FNONBLOCK: ", (FNONBLOCK == O_NONBLOCK ? "O_NONBLOCK" : FNONBLOCK), "\n";
print "FAPPEND: ", (FAPPEND == O_APPEND ? "O_APPEND" : FAPPEND), "\n";
pipe(my $r, my $w) or die "pipe: $!";
sysopen(my $fh, "/dev/null", O_RDONLY | O_NDELAY) or die "sysopen: $!";
my $flags = fcntl($fh, F_GETFL, 0) or die "fcntl: $!";
print "opened with O_NDELAY: ", ($flags & O_NONBLOCK ? "non-blocking" : "blocking"), "\n";
`)
	want := "O_NDELAY: O_NONBLOCK\nFNDELAY: O_NONBLOCK\nFNONBLOCK: O_NONBLOCK\nFAPPEND: O_APPEND\n" +
		"opened with O_NDELAY: non-blocking\n"
	if status != 0 || out != want {
		t.Fatalf("status %d, output:\n%s\nwant:\n%s", status, out, want)
	}
}
