//go:build darwin || linux

package gperl_test

import (
	"os"
	"os/user"
	"testing"
)

// What the platform cannot do fails with an error; nothing reports success
// for something that did not happen. What it can do, it does.
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
report("mkfifo: ", attempt(POSIX::mkfifo("$dir/fifo", 0600)), "; created=", (-e "$dir/fifo" ? 1 : 0));
report("setsid: ", attempt(POSIX::setsid() != -1));
report("setpgid: ", attempt(POSIX::setpgid(0, 0)));
report("setpriority: ", attempt(eval { setpriority(0, 0, 5) }));
report("nice: ", attempt(defined POSIX::nice(1) ? 1 : 0));
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
report("services: ", (defined getservbyname("http", "tcp") ? "found" : "none"));
report("networks: ", (defined getnetbyname("loopback") ? "found" : "none"));
my @host = gethostbyname("127.0.0.1");
report("host by number: ", (@host ? inet_ntoa($host[4]) : "none"));
my @local = gethostbyname("localhost");
report("host by name: ", (@local ? inet_ntoa($local[4]) : "none"));
report("host by address: ", (defined gethostbyaddr(inet_aton("127.0.0.1"), AF_INET) ? "found" : "none"));
report("inet_aton of a number: ", inet_ntoa(inet_aton("10.1.2.3")));
`)
	wantLines(t, got,
		"chown to the same owner: done",
		"chown leaving both alone: done",
		"chown to another user: EPERM; owner unchanged=1",
		"chown of a missing file: ENOENT",
		"chroot: EPERM",
		"mkfifo: ENOSYS; created=0",
		"setsid: ENOSYS",
		"setpgid: ENOSYS",
		"setpriority: ENOSYS",
		"nice: ENOSYS",
		"ttyname of a file: ENOTTY",
		"anonymous temporary file: fd above stdio=1 reads back=kept",
		"cuserid: "+me.Username,
		"protocols: tcp=6 17=udp unknown=none",
		"protocol table: ip icmp tcp udp ipv6 ipv6-icmp",
		"services: none",
		"networks: none",
		"host by number: 127.0.0.1",
		"host by name: 127.0.0.1",
		"host by address: none",
		"inet_aton of a number: 10.1.2.3",
	)
}
