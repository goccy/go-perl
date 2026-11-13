//go:build darwin || linux

package gperl_test

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"testing"
)

// The program that is the host process sees the numbers of the host
// operating system: error numbers, the bits of poll, the flags of open, the
// numbers of sockets, signals. The tests compare with the numbers package
// syscall has for this system and, where a C compiler is there, with the
// system's headers.

// hostPerl runs the test binary as `perl args...` and returns what it
// printed and its exit status.
func hostPerl(t *testing.T, args ...string) (stdout string, status int) {
	t.Helper()
	cmd := perlProcess(args...)
	var errOut strings.Builder
	cmd.Stderr = &errOut
	out, err := cmd.Output()
	var exit *exec.ExitError
	switch {
	case err == nil:
	case errors.As(err, &exit):
		status = exit.ExitCode()
	default:
		t.Fatalf("perl %q: %v\n%s", args, err, errOut.String())
	}
	if status != 0 && errOut.Len() > 0 {
		t.Logf("status %d, stderr: %s", status, errOut.String())
	}
	return string(out), status
}

// perlNumbers runs a program that prints "name number" lines and returns
// them.
func perlNumbers(t *testing.T, program string) map[string]int64 {
	t.Helper()
	out, status := hostPerl(t, "-e", program)
	if status != 0 {
		t.Fatalf("status %d, output %q", status, out)
	}
	numbers := map[string]int64{}
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		name, value, ok := strings.Cut(line, " ")
		n, err := strconv.ParseInt(value, 10, 64)
		if !ok || err != nil {
			t.Fatalf("line %q of the output", line)
		}
		numbers[name] = n
	}
	return numbers
}

// headerNumbers returns the values the system's C headers give the names,
// for the names they define; the test is skipped without a C compiler.
func headerNumbers(t *testing.T, headers []string, names []string) map[string]int64 {
	t.Helper()
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("no C compiler to read the system's headers with")
	}
	var src strings.Builder
	src.WriteString("#define _GNU_SOURCE 1\n#define _DARWIN_C_SOURCE 1\n#include <stdio.h>\n")
	for _, h := range headers {
		fmt.Fprintf(&src, "#include <%s>\n", h)
	}
	src.WriteString("int main(void) {\n")
	for _, name := range names {
		fmt.Fprintf(&src, "#ifdef %s\n    printf(\"%s %%lld\\n\", (long long)(%s));\n#endif\n", name, name, name)
	}
	src.WriteString("    return 0;\n}\n")
	dir := t.TempDir()
	file := filepath.Join(dir, "numbers.c")
	if err := os.WriteFile(file, []byte(src.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	bin := filepath.Join(dir, "numbers")
	if out, err := exec.Command(cc, "-o", bin, file).CombinedOutput(); err != nil {
		t.Fatalf("cc: %v\n%s", err, out)
	}
	out, err := exec.Command(bin).Output()
	if err != nil {
		t.Fatal(err)
	}
	numbers := map[string]int64{}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		name, value, _ := strings.Cut(line, " ")
		n, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			t.Fatalf("line %q of the C program", line)
		}
		numbers[name] = n
	}
	return numbers
}

func sortedNames(numbers map[string]int64) []string {
	names := make([]string, 0, len(numbers))
	for name := range numbers {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// compareWithHeaders checks every number the program has against the
// system's headers: a name the headers define has their value.
func compareWithHeaders(t *testing.T, what string, headers []string, got map[string]int64, atLeast int) {
	t.Helper()
	want := headerNumbers(t, headers, sortedNames(got))
	if len(want) < atLeast {
		t.Fatalf("%s: the headers define %d of the %d names, expected at least %d", what, len(want), len(got), atLeast)
	}
	for _, name := range sortedNames(want) {
		if got[name] != want[name] {
			t.Errorf("%s %s = %d, the system's is %d", what, name, got[name], want[name])
		}
	}
}

// Every error constant of Errno and of POSIX is the system's, and Errno has
// the system's names.
func TestErrorNumbersAreTheSystems(t *testing.T) {
	errno := perlNumbers(t, `use Errno; print "$_ ", Errno->can($_)->(), "\n" for sort keys %!`)
	posix := perlNumbers(t, `use POSIX (); for (sort grep { /\AE[A-Z0-9]+\z/ && !/\AE(?:OF|XIT_\w+)\z/ } keys %POSIX::) { my $v = eval { POSIX->can($_)->() }; print "$_ $v\n" if defined $v }`)
	for name, want := range map[string]syscall.Errno{
		"EPERM": syscall.EPERM, "ENOENT": syscall.ENOENT, "EINTR": syscall.EINTR, "EIO": syscall.EIO,
		"EBADF": syscall.EBADF, "EAGAIN": syscall.EAGAIN, "EWOULDBLOCK": syscall.EWOULDBLOCK,
		"ENOMEM": syscall.ENOMEM, "EACCES": syscall.EACCES, "EEXIST": syscall.EEXIST,
		"ENOTDIR": syscall.ENOTDIR, "EISDIR": syscall.EISDIR, "EINVAL": syscall.EINVAL,
		"EMFILE": syscall.EMFILE, "ENOSPC": syscall.ENOSPC, "EPIPE": syscall.EPIPE,
		"ERANGE": syscall.ERANGE, "EDEADLK": syscall.EDEADLK, "ENAMETOOLONG": syscall.ENAMETOOLONG,
		"ENOSYS": syscall.ENOSYS, "ENOTEMPTY": syscall.ENOTEMPTY, "ELOOP": syscall.ELOOP,
		"ENOTSOCK": syscall.ENOTSOCK, "EADDRINUSE": syscall.EADDRINUSE, "ECONNREFUSED": syscall.ECONNREFUSED,
		"ECONNRESET": syscall.ECONNRESET, "ETIMEDOUT": syscall.ETIMEDOUT, "EINPROGRESS": syscall.EINPROGRESS,
		"ENOTSUP": syscall.ENOTSUP, "EOPNOTSUPP": syscall.EOPNOTSUPP, "ENOTCONN": syscall.ENOTCONN,
		"EHOSTUNREACH": syscall.EHOSTUNREACH, "EAFNOSUPPORT": syscall.EAFNOSUPPORT,
		// names the interpreter's own C library does not have
		"ENOTBLK": syscall.ENOTBLK, "ESHUTDOWN": syscall.ESHUTDOWN, "EHOSTDOWN": syscall.EHOSTDOWN,
	} {
		if got, ok := errno[name]; !ok || got != int64(want) {
			t.Errorf("Errno::%s = %d (defined: %v), want %d", name, got, ok, want)
		}
		if got, ok := posix[name]; ok && got != int64(want) {
			t.Errorf("POSIX::%s = %d, want %d", name, got, want)
		}
	}
	if len(posix) < 60 {
		t.Errorf("POSIX has %d error constants", len(posix))
	}
	compareWithHeaders(t, "Errno", []string{"errno.h"}, errno, 70)
	compareWithHeaders(t, "POSIX", []string{"errno.h"}, posix, 60)
}

// A system call that fails leaves the system's number for the error in $!,
// %! knows it by the system's name, and its text is that error's.
func TestFailedCallsLeaveTheSystemsError(t *testing.T) {
	dir := t.TempDir()
	out, status := hostPerl(t, "-e", `
use Errno; use Socket;
my $dir = shift;
sub report { my ($what) = @_; my $n = $! + 0; printf "%s %d %s|%s\n", $what, $n, join(",", sort grep { $!{$_} } keys %!), "$!" }
mkdir "$dir/d" or die; open(my $f, ">", "$dir/d/f") or die; close $f;
rmdir "$dir/d" or report("rmdir-not-empty");
open(my $x, "<", "$dir/none") or report("open-missing");
open($x, "<", "$dir/" . ("n" x 5000)) or report("open-long-name");
mkdir "$dir/d" or report("mkdir-existing");
symlink "loop", "$dir/loop"; open($x, "<", "$dir/loop") or report("open-loop");
open($x, ">", "$dir/d") or report("open-directory");
opendir(my $dh, "$dir/d/f") or report("opendir-file");
socket(my $s, AF_INET, SOCK_STREAM, 0) or die "socket: $!";
socket(my $l, AF_INET, SOCK_STREAM, 0) or die; bind($l, pack_sockaddr_in(0, INADDR_LOOPBACK)) or die "bind: $!";
my $unused = getsockname($l); close $l;
connect($s, $unused) or report("connect-refused");
close $s; { no warnings; print {$s} "x" or report("print-closed"); }
waitpid(-1, 0) < 0 and report("wait-no-child");
kill 0, 999999 or report("kill-no-process");
`, dir)
	if status != 0 {
		t.Fatalf("status %d, output %q", status, out)
	}
	want := map[string]syscall.Errno{
		"rmdir-not-empty": syscall.ENOTEMPTY, "open-missing": syscall.ENOENT,
		"open-long-name": syscall.ENAMETOOLONG, "mkdir-existing": syscall.EEXIST,
		"open-loop": syscall.ELOOP, "open-directory": syscall.EISDIR, "opendir-file": syscall.ENOTDIR,
		"connect-refused": syscall.ECONNREFUSED, "print-closed": syscall.EBADF,
		"wait-no-child": syscall.ECHILD, "kill-no-process": syscall.ESRCH,
	}
	names := map[syscall.Errno]string{
		syscall.ENOTEMPTY: "ENOTEMPTY", syscall.ENOENT: "ENOENT", syscall.ENAMETOOLONG: "ENAMETOOLONG",
		syscall.EEXIST: "EEXIST", syscall.ELOOP: "ELOOP", syscall.EISDIR: "EISDIR", syscall.ENOTDIR: "ENOTDIR",
		syscall.ECONNREFUSED: "ECONNREFUSED", syscall.EBADF: "EBADF", syscall.ECHILD: "ECHILD", syscall.ESRCH: "ESRCH",
	}
	seen := 0
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		fields := strings.SplitN(line, " ", 3)
		if len(fields) != 3 {
			t.Fatalf("line %q", line)
		}
		number, _ := strconv.Atoi(fields[1])
		keys, text, _ := strings.Cut(fields[2], "|")
		w, ok := want[fields[0]]
		if !ok {
			t.Errorf("unexpected line %q", line)
			continue
		}
		seen++
		if number != int(w) || !strings.Contains(","+keys+",", ","+names[w]+",") || text == "" {
			t.Errorf("%s: $! = %d, names %q, text %q; want %d (%s) and a text", fields[0], number, keys, text, int(w), names[w])
		}
	}
	if seen != len(want) {
		t.Errorf("%d of %d failures were reported:\n%s", seen, len(want), out)
	}
}

// $! takes the system's numbers when the program assigns to it, also under
// local, POSIX agrees with it, and a number that is no error of any system
// stays the number it is.
func TestAssigningErrorNumbers(t *testing.T) {
	out, status := hostPerl(t, "-e", `
use Errno qw(ENOENT EACCES EEXIST); use POSIX ();
my ($enoent, $eacces) = @ARGV;
$! = $enoent; print "assigned: ", $!+0, " ", ($!{ENOENT} ? "ENOENT" : "?"), " ", ($! == ENOENT ? "equal" : "differs"), "\n";
{ local $! = $eacces; print "local: ", $!+0, " ", ($!{EACCES} ? "EACCES" : "?"), "\n"; }
print "restored: ", $!+0, "\n";
print "posix: ", POSIX::errno(), " ", (POSIX::strerror($enoent) eq "$!" ? "same text" : "other text"), " ", (length POSIX::strerror($eacces) ? "has text" : "no text"), "\n";
$! = 9999; print "no error of the system: ", $!+0, " [$!]\n";
$! = 0; print "zero: ", $!+0, " [$!]\n";
$! = EEXIST; my $copy = $!; print "copied: ", $copy+0, " ", (length "$copy" ? "with text" : "no text"), "\n";
# every number comes back as it went in; the one exception is a second
# name of the system's for what the interpreter's C library has one error
# for, which comes back as the first
for my $n (1 .. 130) { $! = $n; print "not kept: $n -> ", $!+0, "\n" if $! != $n && !($n == Errno::EOPNOTSUPP() && $! == Errno::ENOTSUP()) }
`, strconv.Itoa(int(syscall.ENOENT)), strconv.Itoa(int(syscall.EACCES)))
	want := fmt.Sprintf("assigned: %d ENOENT equal\nlocal: %d EACCES\nrestored: %d\nposix: %d same text has text\n"+
		"no error of the system: 9999 [Unknown error 9999]\nzero: 0 []\ncopied: %d with text\n",
		int(syscall.ENOENT), int(syscall.EACCES), int(syscall.ENOENT), int(syscall.ENOENT), int(syscall.EEXIST))
	if status != 0 || out != want {
		t.Errorf("status %d\n got: %q\nwant: %q", status, out, want)
	}
}

// The exit status of a perl that dies is the error number, the system's: a
// program file that is not there ends the process with ENOENT's.
func TestExitStatusesFromErrors(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "no-such-program.pl")
	for _, c := range []struct {
		what string
		args []string
		want int
	}{
		{"a program file that does not exist", []string{missing}, int(syscall.ENOENT)},
		{"die after a failed open", []string{"-e", `open(my $f, "<", "/nonexistent/x") or die "open: $!\n"`}, int(syscall.ENOENT)},
		{"die with $! assigned", []string{"-e", `use Errno qw(EACCES); $! = EACCES; die "no\n"`}, int(syscall.EACCES)},
		{"die with $! = 5", []string{"-e", `$! = 5; die "five\n"`}, 5},
		{"die with $? set and no error", []string{"-e", `$! = 0; $? = 3 << 8; die "three\n"`}, 3},
		{"die with neither", []string{"-e", `$! = 0; $? = 0; die "plain\n"`}, 255},
		{"exit with a number", []string{"-e", `exit 44`}, 44},
	} {
		if _, status := hostPerl(t, c.args...); status != c.want {
			t.Errorf("%s: status %d, want %d", c.what, status, c.want)
		}
	}
}

// The bits of poll are the system's: IO::Poll's constants, the masks a
// program gives and the events it gets back.
func TestPollBitsAreTheSystems(t *testing.T) {
	bits := perlNumbers(t, `use IO::Poll qw(POLLIN POLLPRI POLLOUT POLLERR POLLHUP POLLNVAL POLLRDNORM POLLWRNORM POLLRDBAND POLLWRBAND POLLNORM); print "$_ ", IO::Poll->can($_)->(), "\n" for qw(POLLIN POLLPRI POLLOUT POLLERR POLLHUP POLLNVAL POLLRDNORM POLLWRNORM POLLRDBAND POLLWRBAND POLLNORM)`)
	for name, want := range map[string]int64{"POLLIN": 1, "POLLPRI": 2, "POLLOUT": 4, "POLLERR": 8, "POLLHUP": 0x10, "POLLNVAL": 0x20, "POLLRDNORM": 0x40} {
		if bits[name] != want {
			t.Errorf("%s = %#x, want %#x", name, bits[name], want)
		}
	}
	if bits["POLLNORM"] != bits["POLLRDNORM"] {
		t.Errorf("POLLNORM = %#x, POLLRDNORM = %#x", bits["POLLNORM"], bits["POLLRDNORM"])
	}
	delete(bits, "POLLNORM") // IO::Poll's own name
	compareWithHeaders(t, "IO::Poll", []string{"poll.h"}, bits, 10)

	// Through a pipe: numbers written out, as a program that knows its
	// system writes them, and IO::Poll's own.
	out, status := hostPerl(t, "-e", `
use IO::Poll qw(POLLIN POLLOUT POLLPRI POLLHUP POLLERR POLLNVAL POLLRDNORM POLLWRNORM);
pipe(my $r, my $w) or die "pipe: $!";
my @p = (fileno($w), 4, fileno($r), 1);
my $n = IO::Poll::_poll(0, @p);
print "by number: $n ready; writer events $p[1], reader events $p[3]\n";
syswrite($w, "x");
@p = (fileno($w), 4, fileno($r), 1);
$n = IO::Poll::_poll(1000, @p);
print "after a write: $n ready; writer events $p[1], reader events $p[3]\n";
my $poll = IO::Poll->new; $poll->mask($r => POLLIN); $poll->mask($w => POLLOUT);
$n = $poll->poll(1);
print "IO::Poll: $n ready; reader ", ($poll->events($r) == POLLIN ? "POLLIN" : $poll->events($r)), ", writer ", ($poll->events($w) == POLLOUT ? "POLLOUT" : $poll->events($w)), "\n";
sysread($r, my $buf, 1); close $w;
@p = (fileno($r), 1);
$n = IO::Poll::_poll(1000, @p);
print "writer closed: $n ready; at its end ", (($p[1] & 0x11) ? "yes" : "no"), "; other bits ", ($p[1] & ~0x11), "\n";
@p = (987, 1);
$n = IO::Poll::_poll(0, @p);
print "not a descriptor: events $p[1]\n";
# The events are the bits that were asked for: a loop that asks for
# POLLRDNORM and POLLWRNORM gets those, not POLLIN and POLLOUT.
pipe($r, $w) or die "pipe: $!"; syswrite($w, "x");
@p = (fileno($r), POLLRDNORM, fileno($w), POLLWRNORM);
$n = IO::Poll::_poll(1000, @p);
print "asked for the normal bands: $n ready; ", ($p[1] == POLLRDNORM ? "POLLRDNORM" : $p[1]), " ", ($p[3] == POLLWRNORM ? "POLLWRNORM" : $p[3]), "\n";
@p = (fileno($r), POLLIN | POLLRDNORM | POLLPRI);
$n = IO::Poll::_poll(1000, @p);
print "asked for both names of readable: ", ($p[1] == (POLLIN | POLLRDNORM) ? "both" : $p[1]), "\n";
`)
	want := "by number: 1 ready; writer events 4, reader events 0\n" +
		"after a write: 2 ready; writer events 4, reader events 1\n" +
		"IO::Poll: 2 ready; reader POLLIN, writer POLLOUT\n" +
		"writer closed: 1 ready; at its end yes; other bits 0\n" +
		"not a descriptor: events 32\n" +
		"asked for the normal bands: 2 ready; POLLRDNORM POLLWRNORM\n" +
		"asked for both names of readable: both\n"
	if status != 0 || out != want {
		t.Errorf("status %d\n got: %q\nwant: %q", status, out, want)
	}
}

// The flags of open are the system's: the constants of Fcntl and POSIX, a
// mode written as a number, and what fcntl reports and takes.
func TestOpenFlagsAreTheSystems(t *testing.T) {
	flags := perlNumbers(t, `use Fcntl; print "$_ ", Fcntl->can($_)->(), "\n" for qw(O_RDONLY O_WRONLY O_RDWR O_ACCMODE O_APPEND O_CREAT O_EXCL O_TRUNC O_NONBLOCK O_NOCTTY O_NOFOLLOW O_DIRECTORY O_SYNC F_GETFD F_SETFD F_GETFL F_SETFL FD_CLOEXEC)`)
	for name, want := range map[string]int{
		"O_RDONLY": syscall.O_RDONLY, "O_WRONLY": syscall.O_WRONLY, "O_RDWR": syscall.O_RDWR,
		"O_ACCMODE": syscall.O_ACCMODE, "O_APPEND": syscall.O_APPEND, "O_CREAT": syscall.O_CREAT,
		"O_EXCL": syscall.O_EXCL, "O_TRUNC": syscall.O_TRUNC, "O_NONBLOCK": syscall.O_NONBLOCK,
		"O_NOCTTY": syscall.O_NOCTTY, "O_NOFOLLOW": syscall.O_NOFOLLOW, "O_DIRECTORY": syscall.O_DIRECTORY,
		"O_SYNC": syscall.O_SYNC, "F_GETFL": syscall.F_GETFL, "F_SETFL": syscall.F_SETFL,
	} {
		if flags[name] != int64(want) {
			t.Errorf("Fcntl::%s = %#x, want %#x", name, flags[name], want)
		}
	}
	posix := perlNumbers(t, `use POSIX (); print "$_ ", POSIX->can($_)->(), "\n" for qw(O_RDONLY O_WRONLY O_RDWR O_ACCMODE O_APPEND O_CREAT O_EXCL O_TRUNC O_NONBLOCK O_NOCTTY)`)
	for name, value := range posix {
		if value != flags[name] {
			t.Errorf("POSIX::%s = %#x, Fcntl's is %#x", name, value, flags[name])
		}
	}
	compareWithHeaders(t, "Fcntl", []string{"fcntl.h"}, flags, 16)

	dir := t.TempDir()
	out, status := hostPerl(t, "-e", `
use Fcntl; use Errno; use POSIX ();
my ($path, $create, $rdwr, $nonblock, $accmode, $unknown) = @ARGV;
sysopen(my $fh, $path, $create) or die "sysopen by number: $!";
print "created\n";
print "again: ", (sysopen(my $again, $path, $create) ? "opened" : ($!{EEXIST} ? "EEXIST" : "$!")), "\n";
syswrite($fh, "data\n");
my $got = fcntl($fh, F_GETFL, 0);
print "access of a write-only handle: ", (($got & O_ACCMODE) == O_WRONLY ? "O_WRONLY" : sprintf("%#x", $got)), "\n";
close $fh;
sysopen($fh, $path, $rdwr) or die "sysopen read-write: $!";
$got = fcntl($fh, F_GETFL, 0) or die "F_GETFL: $!";
print "read-write by number: ", (($got & $accmode) == $rdwr ? "yes" : sprintf("%#x", $got)), "; nonblocking ", (($got & $nonblock) ? 1 : 0), "\n";
fcntl($fh, F_SETFL, $got | $nonblock) or die "F_SETFL: $!";
$got = fcntl($fh, F_GETFL, 0);
print "after F_SETFL: nonblocking ", (($got & O_NONBLOCK) ? 1 : 0), ", still read-write ", (($got & O_ACCMODE) == O_RDWR ? 1 : 0), "\n";
close $fh;
sysopen($fh, $path, O_RDONLY) or die "sysopen O_RDONLY: $!";
$got = fcntl($fh, F_GETFL, 0);
print "read-only: flags ", ($got == 0 ? "0" : sprintf("%#x", $got)), ", true ", ($got ? 1 : 0), "\n";
close $fh;
my $fd = POSIX::open($path, $rdwr | O_APPEND) // die "POSIX::open: $!";
POSIX::write($fd, "more\n", 5); POSIX::close($fd);
open($fh, "<", $path) or die; local $/; print "content: ", join("|", split /\n/, <$fh>), "\n"; close $fh;
print "a flag no system has: ", (sysopen($fh, $path, $rdwr | $unknown) ? "opened" : ($!{EINVAL} ? "EINVAL" : "$!")), "\n";
print "POSIX::open with it: ", (defined POSIX::open($path, $rdwr | $unknown) ? "opened" : ($!{EINVAL} ? "EINVAL" : "$!")), "\n";
`, filepath.Join(dir, "file"), strconv.Itoa(syscall.O_WRONLY|syscall.O_CREAT|syscall.O_EXCL),
		strconv.Itoa(syscall.O_RDWR), strconv.Itoa(syscall.O_NONBLOCK), strconv.Itoa(syscall.O_ACCMODE),
		strconv.Itoa(1<<29))
	want := "created\nagain: EEXIST\naccess of a write-only handle: O_WRONLY\n" +
		"read-write by number: yes; nonblocking 0\n" +
		"after F_SETFL: nonblocking 1, still read-write 1\n" +
		"read-only: flags 0, true 1\n" +
		"content: data|more\n" +
		"a flag no system has: EINVAL\nPOSIX::open with it: EINVAL\n"
	if status != 0 || out != want {
		t.Errorf("status %d\n got: %q\nwant: %q", status, out, want)
	}
}

// The numbers of sockets are the system's: address families, types, levels,
// options and flags, as constants, written as numbers, and in what the
// socket functions return.
func TestSocketNumbersAreTheSystems(t *testing.T) {
	names := `AF_UNSPEC AF_INET AF_INET6 AF_UNIX PF_INET PF_INET6 PF_UNIX SOCK_STREAM SOCK_DGRAM SOL_SOCKET
		SO_REUSEADDR SO_KEEPALIVE SO_BROADCAST SO_ERROR SO_TYPE SO_RCVBUF SO_SNDBUF SO_LINGER SO_ACCEPTCONN
		MSG_PEEK MSG_WAITALL MSG_DONTWAIT MSG_OOB IPPROTO_TCP IPPROTO_IP IPPROTO_IPV6 TCP_NODELAY IP_TTL IPV6_V6ONLY
		AI_PASSIVE AI_NUMERICHOST NI_NUMERICHOST NI_NUMERICSERV EAI_NONAME`
	socket := perlNumbers(t, `use Socket (); for (qw(`+names+`)) { my $v = eval { Socket->can($_)->() }; print "$_ $v\n" if defined $v }`)
	for name, want := range map[string]int{
		"AF_INET": syscall.AF_INET, "AF_INET6": syscall.AF_INET6, "AF_UNIX": syscall.AF_UNIX,
		"PF_INET": syscall.AF_INET, "PF_INET6": syscall.AF_INET6, "PF_UNIX": syscall.AF_UNIX,
		"SOCK_STREAM": syscall.SOCK_STREAM, "SOCK_DGRAM": syscall.SOCK_DGRAM, "SOL_SOCKET": syscall.SOL_SOCKET,
		"SO_REUSEADDR": syscall.SO_REUSEADDR, "SO_KEEPALIVE": syscall.SO_KEEPALIVE, "SO_ERROR": syscall.SO_ERROR,
		"SO_TYPE": syscall.SO_TYPE, "SO_RCVBUF": syscall.SO_RCVBUF, "MSG_PEEK": syscall.MSG_PEEK,
		"MSG_WAITALL": syscall.MSG_WAITALL, "MSG_DONTWAIT": syscall.MSG_DONTWAIT, "TCP_NODELAY": syscall.TCP_NODELAY,
		"IP_TTL": syscall.IP_TTL, "IPV6_V6ONLY": syscall.IPV6_V6ONLY,
	} {
		if got, ok := socket[name]; !ok || got != int64(want) {
			t.Errorf("Socket::%s = %d (defined: %v), want %d", name, got, ok, want)
		}
	}
	compareWithHeaders(t, "Socket", []string{"sys/types.h", "sys/socket.h", "netinet/in.h", "netinet/tcp.h", "netdb.h"}, socket, 30)

	out, status := hostPerl(t, "-e", `
use Socket qw(:DEFAULT :addrinfo inet_pton inet_ntop); use Errno; use Fcntl;
my ($af_inet, $sock_stream, $sol_socket, $so_reuseaddr, $so_type, $so_error, $msg_peek, $af_inet6) = @ARGV;
socket(my $l, $af_inet, $sock_stream, 0) or die "socket by number: $!";
setsockopt($l, $sol_socket, $so_reuseaddr, pack("i", 1)) or die "setsockopt by number: $!";
print "SO_REUSEADDR: ", (unpack("i", getsockopt($l, SOL_SOCKET, SO_REUSEADDR)) ? "set" : "not set"), "\n";
print "SO_TYPE: ", (unpack("i", getsockopt($l, $sol_socket, $so_type)) == $sock_stream ? "the stream type" : "another number"), "\n";
bind($l, pack_sockaddr_in(0, INADDR_LOOPBACK)) or die "bind: $!"; listen($l, 5) or die "listen: $!";
my $addr = getsockname($l);
print "family of the address: ", (sockaddr_family($addr) == $af_inet ? "AF_INET" : sockaddr_family($addr)), "\n";
socket(my $c, AF_INET, SOCK_STREAM, 0) or die; connect($c, $addr) or die "connect: $!";
accept(my $s, $l) or die "accept: $!";
send($c, "hello", 0) or die "send: $!";
recv($s, my $peeked, 5, $msg_peek) // die "recv peek: $!";
recv($s, my $read, 5, 0) // die "recv: $!";
print "peeked $peeked, then read $read\n";
print "pending error of a good connection: ", unpack("i", getsockopt($c, $sol_socket, $so_error)), "\n";
close $_ for $c, $s, $l;
socket(my $n, AF_INET, SOCK_STREAM, 0) or die;
my $flags = fcntl($n, F_GETFL, 0); fcntl($n, F_SETFL, $flags | O_NONBLOCK) or die "nonblocking: $!";
my $pending = 0;
if (!connect($n, $addr)) {
    if ($!{EINPROGRESS}) { my $w = ""; vec($w, fileno($n), 1) = 1; select(undef, $w, undef, 5); $pending = unpack("i", getsockopt($n, SOL_SOCKET, SO_ERROR)); }
    else { $pending = $! + 0 }
}
$! = $pending;
print "pending error of a refused connection: ", ($!{ECONNREFUSED} ? "ECONNREFUSED" : "$pending ($!)"), "\n";
my ($err, @res) = getaddrinfo("127.0.0.1", "80", { flags => AI_NUMERICHOST | AI_NUMERICSERV, socktype => $sock_stream });
print "getaddrinfo: ", ($err ? "error $err" : "family " . ($res[0]{family} == $af_inet ? "AF_INET" : $res[0]{family}) . ", type " . ($res[0]{socktype} == $sock_stream ? "stream" : $res[0]{socktype})), "\n";
($err) = getaddrinfo("no.such.host.invalid", "", { flags => AI_NUMERICHOST });
print "getaddrinfo of a name that is no address: ", ($err == EAI_NONAME ? "EAI_NONAME" : $err + 0), ", with a text: ", (length "$err" ? "yes" : "no"), "\n";
($err, my $host, my $service) = getnameinfo(pack_sockaddr_in(80, INADDR_LOOPBACK), NI_NUMERICHOST | NI_NUMERICSERV);
print "getnameinfo: $host $service\n";
print "inet_pton by number: ", (defined inet_pton($af_inet6, "::1") ? length(inet_pton($af_inet6, "::1")) : "undef"), " bytes, back: ", inet_ntop($af_inet6, inet_pton($af_inet6, "::1")), "\n";
print "a family no system has: ", (socket(my $x, 12345, $sock_stream, 0) ? "opened" : ($!{EAFNOSUPPORT} ? "EAFNOSUPPORT" : "$!")), "\n";
`, strconv.Itoa(syscall.AF_INET), strconv.Itoa(syscall.SOCK_STREAM), strconv.Itoa(syscall.SOL_SOCKET),
		strconv.Itoa(syscall.SO_REUSEADDR), strconv.Itoa(syscall.SO_TYPE), strconv.Itoa(syscall.SO_ERROR),
		strconv.Itoa(syscall.MSG_PEEK), strconv.Itoa(syscall.AF_INET6))
	want := "SO_REUSEADDR: set\nSO_TYPE: the stream type\nfamily of the address: AF_INET\n" +
		"peeked hello, then read hello\npending error of a good connection: 0\n" +
		"pending error of a refused connection: ECONNREFUSED\n" +
		"getaddrinfo: family AF_INET, type stream\n" +
		"getaddrinfo of a name that is no address: EAI_NONAME, with a text: yes\n" +
		"getnameinfo: 127.0.0.1 80\n" +
		"inet_pton by number: 16 bytes, back: ::1\n" +
		"a family no system has: EAFNOSUPPORT\n"
	if status != 0 || out != want {
		t.Errorf("status %d\n got: %q\nwant: %q", status, out, want)
	}
}

// The requests of ioctl that the interpreter answers and the items of
// langinfo have the system's numbers too.
func TestIoctlAndLanginfoNumbersAreTheSystems(t *testing.T) {
	items := perlNumbers(t, `use I18N::Langinfo (); print "$_ ", I18N::Langinfo->can($_)->(), "\n" for qw(CODESET RADIXCHAR THOUSEP D_FMT T_FMT DAY_1 MON_1 ABDAY_1 YESEXPR NOEXPR AM_STR PM_STR CRNCYSTR)`)
	compareWithHeaders(t, "I18N::Langinfo", []string{"langinfo.h"}, items, 12)
	requests := headerNumbers(t, []string{"sys/ioctl.h"}, []string{"FIONREAD"})
	out, status := hostPerl(t, "-e", `
use I18N::Langinfo qw(langinfo CODESET RADIXCHAR);
my ($fionread, $codeset, $radixchar) = @ARGV;
pipe(my $r, my $w) or die "pipe: $!"; syswrite($w, "12345");
my $n = pack("i", 0);
print "ioctl by number: ", (ioctl($r, $fionread, $n) ? unpack("i", $n) . " bytes to read" : "failed: $!"), "\n";
print "langinfo by number: the radix character is '", langinfo($radixchar), "', by name '", langinfo(RADIXCHAR), "'\n";
print "the codeset by number and by name: ", (langinfo($codeset) eq langinfo(CODESET) && length langinfo(CODESET) ? "the same" : "differ"), "\n";
`, strconv.FormatInt(requests["FIONREAD"], 10), strconv.FormatInt(items["CODESET"], 10), strconv.FormatInt(items["RADIXCHAR"], 10))
	want := "ioctl by number: 5 bytes to read\n" +
		"langinfo by number: the radix character is '.', by name '.'\n" +
		"the codeset by number and by name: the same\n"
	if status != 0 || out != want {
		t.Errorf("status %d\n got: %q\nwant: %q", status, out, want)
	}
}

// The numbers of signals are the system's: POSIX's constants, %Config's
// list, a signal sent by number, the number in the status of a child a
// signal ended, and the numbers POSIX's signal functions take.
func TestSignalNumbersAreTheSystems(t *testing.T) {
	signals := perlNumbers(t, `use POSIX (); use Config; my @n = split " ", $Config{sig_name}; my @v = split " ", $Config{sig_num}; my %c; @c{@n} = @v;
for (qw(HUP INT QUIT KILL USR1 USR2 PIPE ALRM TERM CHLD CONT STOP TSTP)) { print "SIG$_ ", POSIX->can("SIG$_")->(), "\n"; print "config-$_ $c{$_}\n" }
for (qw(SIG_BLOCK SIG_UNBLOCK SIG_SETMASK SA_RESTART SA_NOCLDSTOP)) { print "$_ ", POSIX->can($_)->(), "\n" }`)
	for name, want := range map[string]syscall.Signal{
		"HUP": syscall.SIGHUP, "INT": syscall.SIGINT, "KILL": syscall.SIGKILL, "USR1": syscall.SIGUSR1,
		"USR2": syscall.SIGUSR2, "PIPE": syscall.SIGPIPE, "TERM": syscall.SIGTERM, "CHLD": syscall.SIGCHLD,
		"CONT": syscall.SIGCONT, "STOP": syscall.SIGSTOP, "TSTP": syscall.SIGTSTP,
	} {
		if signals["SIG"+name] != int64(want) {
			t.Errorf("POSIX::SIG%s = %d, want %d", name, signals["SIG"+name], int(want))
		}
		if signals["config-"+name] != int64(want) {
			t.Errorf("%%Config's number of %s = %d, want %d", name, signals["config-"+name], int(want))
		}
	}
	posix := map[string]int64{}
	for name, value := range signals {
		if !strings.HasPrefix(name, "config-") {
			posix[name] = value
		}
	}
	compareWithHeaders(t, "POSIX", []string{"signal.h"}, posix, 15)

	out, status := hostPerl(t, "-e", `
use POSIX ();
my ($usr1, $usr2, $term) = @ARGV;
my @got;
$SIG{USR1} = sub { push @got, "USR1" }; $SIG{USR2} = sub { push @got, "USR2" };
kill $usr1, $$; kill $usr2, $$; kill "USR1", $$;
print "handlers ran for: @got\n";
POSIX::sigaction($usr2, POSIX::SigAction->new(sub { push @got, "by sigaction" })) or die "sigaction: $!";
kill $usr2, $$;
print "the last: $got[-1]\n";
my $set = POSIX::SigSet->new($usr1);
print "in the set: ", ($set->ismember($usr1) ? "USR1" : ""), ($set->ismember($usr2) ? " USR2" : ""), "\n";
POSIX::sigprocmask(POSIX::SIG_BLOCK(), $set) or die "sigprocmask: $!";
@got = (); kill $usr1, $$; print "while blocked: ", scalar(@got), " delivered\n";
POSIX::sigprocmask(POSIX::SIG_UNBLOCK(), $set) or die "sigprocmask: $!";
print "after unblocking: @got\n";
system("sh", "-c", 'kill -USR1 $$');
print "status of a child that USR1 ended: signal ", ($? & 127) == $usr1 ? "USR1" : ($? & 127), ", WTERMSIG ", (POSIX::WTERMSIG($?) == $usr1 ? "USR1" : POSIX::WTERMSIG($?)), "\n";
my $pid = fork() // die "fork: $!";
if (!$pid) { sleep 30; exit 0 }
kill $term, $pid; waitpid($pid, 0);
print "status of a child that TERM ended: ", ($? & 127) == $term ? "TERM" : $?, "\n";
`, strconv.Itoa(int(syscall.SIGUSR1)), strconv.Itoa(int(syscall.SIGUSR2)), strconv.Itoa(int(syscall.SIGTERM)))
	want := "handlers ran for: USR1 USR2 USR1\nthe last: by sigaction\nin the set: USR1\n" +
		"while blocked: 0 delivered\nafter unblocking: USR1\n" +
		"status of a child that USR1 ended: signal USR1, WTERMSIG USR1\n" +
		"status of a child that TERM ended: TERM\n"
	if status != 0 || out != want {
		t.Errorf("status %d\n got: %q\nwant: %q", status, out, want)
	}
}
