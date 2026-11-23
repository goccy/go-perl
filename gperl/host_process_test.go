//go:build darwin || linux

package gperl_test

import (
	"os"
	"testing"
)

// The children of a process that ignores SIGCHLD leave nothing to wait for:
// waitpid lasts until the child has ended and reports ECHILD, and so does
// the wait a command's status comes from.
func TestIgnoredSIGCHLD(t *testing.T) {
	got := runForkScript(t, `
use POSIX ();
$SIG{CHLD} = 'IGNORE';
my $pid = fork // die "fork: $!";
if (!$pid) { select(undef, undef, undef, 0.2); POSIX::_exit(3) }
my $reaped = waitpid($pid, 0);
report("waitpid=$reaped status=$? ", ($!{ECHILD} ? "ECHILD" : "error $!"));
report("the child is gone=", (kill(0, $pid) ? 0 : 1));
my $said = `+"`echo said`"+`;
chomp $said;
report("command: said=$said status=$?");
$SIG{CHLD} = 'DEFAULT';
$pid = fork // die "fork: $!";
if (!$pid) { POSIX::_exit(3) }
$reaped = waitpid($pid, 0);
report("default action: reaped=", ($reaped == $pid ? 1 : 0), " status=", $? >> 8);
`)
	wantLines(t, got,
		"waitpid=-1 status=-1 ECHILD",
		"the child is gone=1",
		"command: said=said status=-1",
		"default action: reaped=1 status=3",
	)
}

// A command inherits the descriptors that are not close-on-exec, under
// their numbers: none above $^F unless the program clears the flag.
func TestCommandsInheritDescriptors(t *testing.T) {
	got := runForkScript(t, `
use Fcntl;
pipe(my $r, my $w) or die "pipe: $!";
report("a new descriptor is close-on-exec=", (fcntl($w, F_GETFD, 0) & FD_CLOEXEC ? 1 : 0));
system("/bin/sh", "-c", "echo hidden >&" . fileno($w) . " 2>/dev/null");
fcntl($w, F_SETFD, 0) or die "F_SETFD: $!";
report("after clearing the flag=", (fcntl($w, F_GETFD, 0) & FD_CLOEXEC ? 1 : 0));
system("/bin/sh", "-c", "echo passed >&" . fileno($w));
close $w;
report("the command wrote: ", join(" ", map { chomp; $_ } <$r>));
{
    local $^F = 20;
    pipe(my $r2, my $w2) or die "pipe: $!";
    system("/bin/sh", "-c", "echo system descriptor >&" . fileno($w2));
    close $w2;
    report("with \$^F raised: ", join(" ", map { chomp; $_ } <$r2>));
}
# through fork and exec, as IPC::Run and IPC::Open3 hand descriptors over
pipe(my $r3, my $w3) or die "pipe: $!";
my $pid = fork // die "fork: $!";
if (!$pid) {
    close $r3;
    fcntl($w3, F_SETFD, 0) or die "F_SETFD: $!";
    exec("/bin/sh", "-c", "echo from the exec >&" . fileno($w3)) or die "exec: $!";
}
close $w3;
report("fork and exec: ", join(" ", map { chomp; $_ } <$r3>));
waitpid($pid, 0);
`)
	wantLines(t, got,
		"a new descriptor is close-on-exec=1",
		"after clearing the flag=0",
		"the command wrote: passed",
		"with $^F raised: system descriptor",
		"fork and exec: from the exec",
	)
}

// The process is the host process: what it reports and does about itself
// is the host's — CPU time spent in the system, the pid of a command, the
// host's name, FIFOs and file owners of the host filesystem, the scheduling
// priority.
func TestProcessIsTheHostProcess(t *testing.T) {
	host, err := os.Hostname()
	if err != nil {
		t.Skip(err)
	}
	// The interpreter's C library has Linux's sizes: the nodename of uname
	// holds 64 bytes of the name, and Sys::Hostname reads gethostname into a
	// buffer of MAXHOSTNAMELEN (64) bytes, which holds 63 and the NUL. A
	// longer name of the host (darwin's may be up to 255 bytes) is cut there.
	nodename, hostname := host[:min(len(host), 64)], host[:min(len(host), 63)]
	dir := t.TempDir()
	got := runForkScriptArgs(t, `
use POSIX ();
my ($dir) = @ARGV;
my $start = time;
do { opendir(my $d, $dir) or die; my @e = readdir $d; } while time - $start < 1;
my @t = times;
report("times: user=", ($t[0] > 0 ? "spent" : 0), " system=", ($t[1] > 0 ? "spent" : 0));
system("/bin/sh", "-c", 'i=0; while [ $i -lt 30000 ]; do i=$((i+1)); done');
@t = times;
report("times: of the children=", ($t[2] + $t[3] > 0 ? "spent" : 0));

my $pid = open(my $from, "-|", "/bin/sh", "-c", 'echo $$') // die "spawn: $!";
my $said = <$from>;
chomp $said;
close $from;
report("pid of a command=", ($pid == $said ? "the command's own" : "$pid, the command says $said"));

my @uname = POSIX::uname();
report("uname: nodename=$uname[1]");
require Sys::Hostname;
report("hostname=", Sys::Hostname::hostname());

my $fifo = "$dir/fifo";
report("mkfifo=", (POSIX::mkfifo($fifo, 0600) ? 1 : "0 $!"), " is a FIFO=", (-p $fifo ? 1 : 0));
my $writer = fork // die "fork: $!";
if (!$writer) { open(my $out, ">", $fifo) or die "open: $!"; syswrite($out, "through the FIFO\n"); POSIX::_exit(0) }
open(my $in, "<", $fifo) or die "open: $!";
my $line = <$in>;
chomp $line;
report("read: $line");
close $in;
waitpid($writer, 0);

open(my $fh, ">", "$dir/owned") or die "create: $!";
report("chown to the owner=", chown($<, -1, "$dir/owned"), " fchown=", chown($<, -1, $fh));
close $fh;
if ($< != 0) {
    report("chown to root=", chown(0, 0, "$dir/owned"), " ", ($!{EPERM} ? "EPERM" : "error $!"));
    report("setuid to another user=", (POSIX::setuid($< + 1) ? 1 : 0), " ", ($!{EPERM} ? "EPERM" : "error $!"));
} else {
    report("chown to root=0 EPERM");
    report("setuid to another user=0 EPERM");
}
my $nice = POSIX::nice(0);
report("nice(0)=", (defined $nice ? "the value" : "undef $!"));
report("getpriority=", (getpriority(0, 0) == getpriority(0, $$) ? "one value" : "differs"));
`, dir)
	wantLines(t, got,
		"times: user=spent system=spent",
		"times: of the children=spent",
		"pid of a command=the command's own",
		"uname: nodename="+nodename,
		"hostname="+hostname,
		"mkfifo=1 is a FIFO=1",
		"read: through the FIFO",
		"chown to the owner=1 fchown=1",
		"chown to root=0 EPERM",
		"setuid to another user=0 EPERM",
		"nice(0)=the value",
		"getpriority=one value",
	)
}

// The services database and the names of addresses are the host's:
// getservbyname and getservbyport, gethostbyaddr, and getnameinfo with
// names.
func TestNetworkDatabasesAreTheHosts(t *testing.T) {
	if _, err := os.Stat("/etc/services"); err != nil {
		t.Skip("the host has no services database")
	}
	got := runForkScript(t, `
use Socket qw(inet_aton pack_sockaddr_in getnameinfo AF_INET NI_NUMERICHOST NI_NUMERICSERV NI_NAMEREQD);
my @by_name = getservbyname("http", "tcp");
report("getservbyname: name=$by_name[0] port=$by_name[2] proto=$by_name[3]");
report("in scalar context: ", scalar(getservbyname("domain", "udp")));
my @by_port = getservbyport(80, "tcp");
report("getservbyport: name=$by_port[0] port=$by_port[2] proto=$by_port[3]");
report("by an alias: ", (getservbyname("www", "tcp"))[2] // "none");
report("no such service: ", (defined getservbyname("no-such-service-anywhere", "tcp") ? "found" : "none"));
report("no such protocol: ", (defined getservbyname("http", "no-such-protocol") ? "found" : "none"));

my $name = gethostbyaddr(inet_aton("127.0.0.1"), AF_INET);
report("gethostbyaddr of the loopback address: ", (defined $name && $name !~ /^[0-9.]+$/ ? "a name" : "none"));
my ($err, $host, $service) = getnameinfo(pack_sockaddr_in(80, inet_aton("127.0.0.1")));
report("getnameinfo: error=", ($err ? "$err" : "none"), " host=", ($host eq $name ? "the same name" : $host), " service=$service");
($err, $host, $service) = getnameinfo(pack_sockaddr_in(80, inet_aton("127.0.0.1")), NI_NUMERICHOST | NI_NUMERICSERV);
report("getnameinfo in numbers: host=$host service=$service");
($err, $host, $service) = getnameinfo(pack_sockaddr_in(1, inet_aton("192.0.2.77")), NI_NAMEREQD);
report("an address without a name, a name required: ", ($err ? "refused" : "host=$host"));
`)
	wantLines(t, got,
		"getservbyname: name=http port=80 proto=tcp",
		"in scalar context: 53",
		"getservbyport: name=http port=80 proto=tcp",
		"by an alias: 80",
		"no such service: none",
		"no such protocol: none",
		"gethostbyaddr of the loopback address: a name",
		"getnameinfo: error=none host=the same name service=http",
		"getnameinfo in numbers: host=127.0.0.1 service=80",
		"an address without a name, a name required: refused",
	)
}

// stat reports the preferred block size, the blocks allocated to a file,
// and the device a special file stands for.
func TestStatBlocksAndDevice(t *testing.T) {
	got := runForkScript(t, `
my $dir = $REPORT =~ s{/[^/]*$}{}r;
open(my $out, ">", "$dir/data") or die "create: $!";
print {$out} "x" x 10000;
close $out;
my @st = stat("$dir/data");
report("a file of 10000 bytes: block size ", ($st[11] > 0 ? "named" : "'$st[11]'"), ", blocks ", ($st[12] * 512 >= 10000 ? "hold it" : "'$st[12]'"), ", device ", $st[6]);
open(my $in, "<", "$dir/data") or die "open: $!";
my @by_handle = stat($in);
report("by handle: the same=", ("@st[6,11,12]" eq "@by_handle[6,11,12]" ? 1 : 0));
my @zero = stat("/dev/zero");
report("a device: is one=", (-c "/dev/zero" ? 1 : 0), " stands for a device=", ($zero[6] ? 1 : 0));
my @null = stat("/dev/null");
report("the null device stands for a device=", ($null[6] ? 1 : 0));
`)
	wantLines(t, got,
		"a file of 10000 bytes: block size named, blocks hold it, device 0",
		"by handle: the same=1",
		"a device: is one=1 stands for a device=1",
		"the null device stands for a device=1",
	)
}
