package perl_test

import (
	"testing"

	perl "github.com/goccy/go-perl"
)

// The in-memory sandbox is a filesystem of its own, not a view of the
// host's, and it keeps what programs expect of one: the mode a file is
// created with under the umask, chmod, modification times, a stat that says
// the same about a name and a handle, a link count that drops when an open
// file is removed, and locks between separate opens of a file.
func TestSandboxFilesHaveModesTimesLinksAndLocks(t *testing.T) {
	r := runProgram(t, perl.Config{}, perl.Program{Code: code(
		`use strict; use warnings;`,
		`use Fcntl qw(:flock O_CREAT O_WRONLY);`,
		`use File::Temp qw(tempfile unlink0);`,
		`sub mode { sprintf "%o", (stat $_[0])[2] & 07777 }`,
		`my $d = "/scratch";`,

		`umask(027);`,
		`printf "previous mask: %o\n", umask(027);`,
		`mkdir($d, 0777) or die "mkdir: $!";`,
		`sysopen(my $fh, "$d/private", O_CREAT|O_WRONLY, 0600) or die "sysopen: $!"; close $fh;`,
		`open($fh, '+>', "$d/f") or die "open: $!"; print {$fh} "hello"; $fh->flush;`,
		`print "created: dir ", mode($d), ", sysopen 0600 ", mode("$d/private"), ", open ", mode("$d/f"), "\n";`,

		`chmod(0600, "$d/f") or die "chmod: $!";`,
		`print "chmod by name: ", mode("$d/f"), "\n";`,
		`chmod(0644, $fh) or die "chmod: $!";`,
		`print "chmod by handle: ", mode("$d/f"), "\n";`,

		`my $mtime = time - 5000;`,
		`utime($mtime, $mtime, "$d/f") or die "utime: $!";`,
		`print "utime: mtime ", ((stat "$d/f")[9] == $mtime ? "set" : "not set"), "\n";`,

		`my @name = stat "$d/f"; my @handle = stat $fh;`,
		`print "stat: handle ", ("@handle" eq "@name" ? "same" : "differs"), ", nlink $name[3], size $name[7]\n";`,
		`unlink("$d/f") or die "unlink: $!";`,
		`print "unlinked but open: nlink ", (stat $fh)[3], "\n";`,
		`close $fh;`,

		`my ($tfh, $tname) = tempfile(DIR => $d);`,
		`print "unlink0: ", (unlink0($tfh, $tname) ? "ok" : "failed"), "\n";`,
		`close $tfh;`,

		`sub try { flock($_[0], $_[1] | LOCK_NB) ? "granted" : ($!{EWOULDBLOCK} ? "would block" : "error $!") }`,
		`open(my $holder, '>', "$d/lock") or die "open: $!";`,
		`open(my $other, '<', "$d/lock") or die "open: $!";`,
		`open(my $dup, '<&', $holder) or die "dup: $!";`,
		`print "flock: holder ", try($holder, LOCK_EX), ", its dup ", try($dup, LOCK_EX), ", other ", try($other, LOCK_SH), "\n";`,
		`close $holder;`,
		`print "one of the holder's handles closed: other ", try($other, LOCK_SH), "\n";`,
		`close $dup;`,
		`print "holder closed: other ", try($other, LOCK_SH), "\n";`,

		`open(my $null, '+<', "/dev/null") or die "open: $!";`,
		`print "flock on /dev/null: ", try($null, LOCK_EX), "\n";`,
	)}, "")
	r.want(t, 0, ""+
		"previous mask: 27\n"+
		"created: dir 750, sysopen 0600 600, open 640\n"+
		"chmod by name: 600\n"+
		"chmod by handle: 644\n"+
		"utime: mtime set\n"+
		"stat: handle same, nlink 1, size 5\n"+
		"unlinked but open: nlink 0\n"+
		"unlink0: ok\n"+
		"flock: holder granted, its dup granted, other would block\n"+
		"one of the holder's handles closed: other would block\n"+
		"holder closed: other granted\n"+
		"flock on /dev/null: granted\n")
}
