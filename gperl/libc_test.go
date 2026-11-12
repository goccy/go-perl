//go:build darwin || linux

package gperl_test

import (
	"os"
	"os/user"
	"testing"
)

// What the platform cannot do fails with an error; nothing reports success
// for something that did not happen. What it can do, it does. (What the
// process does as the host process — FIFOs, owners, sessions, priorities —
// is in TestProcessIsTheHostProcess and TestProcessGroupsOfForkedProcesses.)
func TestUnsupportedOperationsFailInsteadOfPretending(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("the superuser may give files away")
	}
	me, err := user.Current()
	if err != nil {
		t.Skipf("the host has no entry for the current user: %v", err)
	}
	got := runForkScript(t, `
use POSIX ();
use Socket;
sub failure { $!{ENOSYS} ? "ENOSYS" : $!{EPERM} ? "EPERM" : $!{ENOTTY} ? "ENOTTY" : $!{ENODEV} ? "ENODEV" : "error $!" }
sub attempt { my ($ok) = @_; $ok ? "done" : failure() }
my $dir = $REPORT =~ s{/[^/]*$}{}r;
my $file = "$dir/owned";
open(my $made, '>', $file) or die "create: $!"; close $made;
my ($uid, $gid) = (stat $file)[4, 5];

report("chown to the same owner: ", attempt(chown($uid, $gid, $file)));
report("chown leaving both alone: ", attempt(chown(-1, -1, $file)));
report("chown to another user: ", attempt(chown($uid + 1, -1, $file)), "; owner unchanged=", ((stat $file)[4] == $uid ? 1 : 0));
report("chown of a missing file: ", (chown($uid, $gid, "$dir/missing") ? "done" : ($!{ENOENT} ? "ENOENT" : failure())));
report("chroot: ", attempt(chroot($dir)));
report("ttyname of a file: ", (defined POSIX::ttyname(fileno(STDIN)) ? "named" : failure()));

# a temporary file is a real file of its own
open(my $tmp, '+>', undef) or die "temporary file: $!";
print {$tmp} "kept\n";
seek($tmp, 0, 0);
report("anonymous temporary file: fd above stdio=", (fileno($tmp) > 2 ? 1 : 0), " reads back=", scalar(<$tmp>) =~ s/\n//r);

report("cuserid: ", POSIX::cuserid());
report("protocols: tcp=", scalar(getprotobyname("tcp")), " 17=", scalar(getprotobynumber(17)),
       " unknown=", (defined getprotobyname("no-such-protocol") ? "found" : "none"));
my @all; setprotoent(0); while (my @p = getprotoent()) { push @all, $p[0] } endprotoent();
report("protocol table: @all");
report("networks: ", (defined getnetbyname("loopback") ? "found" : "none"));
my @host = gethostbyname("127.0.0.1");
report("host by number: ", (@host ? inet_ntoa($host[4]) : "none"));
my @local = gethostbyname("localhost");
report("host by name: ", (@local ? inet_ntoa($local[4]) : "none"));
report("inet_aton of a number: ", inet_ntoa(inet_aton("10.1.2.3")));
`)
	wantLines(t, got,
		"chown to the same owner: done",
		"chown leaving both alone: done",
		"chown to another user: EPERM; owner unchanged=1",
		"chown of a missing file: ENOENT",
		"chroot: EPERM",
		"ttyname of a file: ENOTTY",
		"anonymous temporary file: fd above stdio=1 reads back=kept",
		"cuserid: "+me.Username,
		"protocols: tcp=6 17=udp unknown=none",
		"protocol table: ip icmp tcp udp ipv6 ipv6-icmp",
		"networks: none",
		"host by number: 127.0.0.1",
		"host by name: 127.0.0.1",
		"inet_aton of a number: 10.1.2.3",
	)
}

// Read-only is access mode 0 on the systems perl programs are written for,
// and they say so with a literal: sysopen($fh, $path, 0), a DBM tied with
// flags 0.
func TestOpenWithLiteralZeroFlagsIsReadOnly(t *testing.T) {
	got := runForkScript(t, `
use Fcntl;
my $dir = $REPORT =~ s{/[^/]*$}{}r;
my $file = "$dir/data";
open(my $out, '>', $file) or die "create: $!"; print {$out} "content\n"; close $out;
my $in;
report("sysopen with 0: ", (sysopen($in, $file, 0) ? "read " . scalar(<$in>) =~ s/\n//r : "failed: $!"));
report("writing through it: ", (syswrite($in, "x") ? "wrote" : "refused"));
report("sysopen with O_RDONLY: ", (sysopen(my $ro, $file, O_RDONLY) ? "opened" : "failed: $!"));
report("sysopen of a missing file with 0: ", (sysopen(my $none, "$dir/missing", 0) ? "opened" : ($!{ENOENT} ? "ENOENT" : "error $!")));

use SDBM_File;
tie(my %db, 'SDBM_File', "$dir/db", O_CREAT | O_RDWR, 0600) or die "tie: $!";
$db{key} = "value";
untie %db;
my %ro;
report("DBM tied with flags 0: ", (tie(%ro, 'SDBM_File', "$dir/db", 0, 0644) ? "key=$ro{key}" : "failed: $!"));
`)
	wantLines(t, got,
		"sysopen with 0: read content",
		"writing through it: refused",
		"sysopen with O_RDONLY: opened",
		"sysopen of a missing file with 0: ENOENT",
		"DBM tied with flags 0: key=value",
	)
}
