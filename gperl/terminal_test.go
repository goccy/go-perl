//go:build darwin || linux

package gperl_test

import (
	"os"
	"strings"
	"testing"
)

// ptyPrelude gives a test program a pseudo-terminal without a module: the
// requests that make one usable are made with ioctl, under the numbers of
// the system the test runs on, as IO::Tty's C code makes them. open_pty
// returns the handle of the master side and the path of the terminal, or
// nothing where the system has none to give. from_master returns the next
// line that arrives at the master side, or what has arrived of it after
// ten seconds.
const ptyPrelude = `
use POSIX ();
my %REQUEST = $^O eq 'linux'
    ? (unlock => 0x40045431, number => 0x80045430, get_size => 0x5413, set_size => 0x5414,
       take => 0x540E, leave => 0x5422)
    : (grant => 0x20007454, unlock => 0x20007452, name => 0x40807453, get_size => 0x40087468,
       set_size => 0x80087467, take => 0x20007461, leave => 0x20007471);
sub open_pty {
    open(my $master, '+<', '/dev/ptmx') or return;
    if ($^O eq 'linux') {
        my $unlocked = pack('i', 0);
        ioctl($master, $REQUEST{unlock}, $unlocked) or die "unlock: $!";
        my $number = pack('I', 0);
        ioctl($master, $REQUEST{number}, $number) or die "number: $!";
        return ($master, '/dev/pts/' . unpack('I', $number));
    }
    ioctl($master, $REQUEST{grant}, 0) or die "grant: $!";
    ioctl($master, $REQUEST{unlock}, 0) or die "unlock: $!";
    my $name = '';
    ioctl($master, $REQUEST{name}, $name) or die "name: $!";
    $name =~ s/\0.*//s;
    return ($master, $name);
}
my $PENDING = '';
sub from_master {
    my ($master) = @_;
    while ($PENDING !~ /\n/) {
        my $ready = '';
        vec($ready, fileno($master), 1) = 1;
        last unless select($ready, undef, undef, 10);
        last unless sysread($master, $PENDING, 256, length $PENDING);
    }
    return $1 if $PENDING =~ s/\A(.*?)\r?\n//s;
    my $rest = $PENDING;
    $PENDING = '';
    return $rest;
}
`

// The names POSIX has for the settings of a terminal have the system's
// numbers: the bits of the flag words, the places of the control
// characters, the speeds, the selectors of tcsetattr, tcflow and tcflush.
func TestTerminalNumbersAreTheSystems(t *testing.T) {
	got := perlNumbers(t, `use POSIX (); for (qw(VINTR VQUIT VERASE VKILL VEOF VTIME VMIN VSTART VSTOP VSUSP VEOL NCCS
IGNBRK BRKINT IGNPAR PARMRK INPCK ISTRIP INLCR IGNCR ICRNL IXON IXOFF OPOST
B0 B50 B75 B110 B134 B150 B200 B300 B600 B1200 B1800 B2400 B4800 B9600 B19200 B38400
CSIZE CS5 CS6 CS7 CS8 CSTOPB CREAD PARENB PARODD HUPCL CLOCAL
ISIG ICANON ECHO ECHOE ECHOK ECHONL NOFLSH TOSTOP IEXTEN
TCOOFF TCOON TCIOFF TCION TCIFLUSH TCOFLUSH TCIOFLUSH TCSANOW TCSADRAIN TCSAFLUSH)) { print "$_ ", POSIX->can($_)->(), "\n" }`)
	if len(got) != 70 {
		t.Fatalf("POSIX has %d of the 70 names", len(got))
	}
	compareWithHeaders(t, "termios", []string{"termios.h"}, got, 70)
}

// A program makes the requests of a terminal with ioctl, the host's
// numbers and a buffer in the host's layout: the size of the window set
// through one side of a pseudo-terminal is read through the other. Its
// settings are a POSIX::Termios, in the host's numbers. A descriptor that
// is no terminal is refused as the host refuses it.
func TestTerminalRequestsAndSettings(t *testing.T) {
	got := runForkScript(t, ptyPrelude+`
my ($master, $path) = open_pty();
if (!$master) { report("no pseudo-terminal"); exit 0 }
open(my $tty, '+<', $path) or die "open $path: $!";

my $size = pack('S4', 30, 100, 0, 0);
ioctl($master, $REQUEST{set_size}, $size) or die "set the window size: $!";
my $read = '';
my $returned = ioctl($tty, $REQUEST{get_size}, $read);
report("window: ", join("x", (unpack('S4', $read))[0, 1]), " returned=$returned buffer=", length $read);

report("a terminal: -t=", (-t $tty ? 1 : 0), " isatty=", (POSIX::isatty(fileno $tty) ? 1 : 0),
       " ttyname=", (POSIX::ttyname(fileno $tty) eq $path ? "its path" : POSIX::ttyname(fileno $tty) // "none: $!"),
       " ctermid=", POSIX::ctermid());

my $settings = POSIX::Termios->new;
defined $settings->getattr(fileno $tty) or die "getattr: $!";
report("echo at first: ", ($settings->getlflag & POSIX::ECHO() ? "on" : "off"));
$settings->setlflag($settings->getlflag & ~(POSIX::ECHO() | POSIX::ICANON()));
$settings->setcc(POSIX::VMIN(), 3);
$settings->setcc(POSIX::VTIME(), 7);
$settings->setospeed(POSIX::B1200());
$settings->setispeed(POSIX::B1200());
defined $settings->setattr(fileno $tty, POSIX::TCSANOW()) or die "setattr: $!";
my $now = POSIX::Termios->new;
defined $now->getattr(fileno $tty) or die "getattr: $!";
report("after setattr: echo ", ($now->getlflag & POSIX::ECHO() ? "on" : "off"),
       " canonical ", ($now->getlflag & POSIX::ICANON() ? "on" : "off"),
       " VMIN=", $now->getcc(POSIX::VMIN()), " VTIME=", $now->getcc(POSIX::VTIME()),
       " speed ", ($now->getospeed == POSIX::B1200() && $now->getispeed == POSIX::B1200() ? "B1200" : $now->getospeed));
# What is typed at the master is not sent back, and is read as it is.
syswrite($master, "abc");
sysread($tty, my $typed, 16);
my $ready = ''; vec($ready, fileno($master), 1) = 1;
report("typed: read '$typed', echoed: ", (select($ready, undef, undef, 0.3) ? "yes" : "no"));
report("tcflush: ", (defined POSIX::tcflush(fileno $tty, POSIX::TCIOFLUSH()) ? "done" : "failed: $!"),
       ", tcdrain: ", (defined POSIX::tcdrain(fileno $tty) ? "done" : "failed: $!"));

open(my $file, '<', $0) or die;
my $buf = '';
my $on_file = ioctl($file, $REQUEST{get_size}, $buf);
report("the window of a file: ", (defined $on_file ? "answered" : $!{ENOTTY} ? "ENOTTY" : "$!"));
report("the settings of a file: ", (defined POSIX::Termios->new->getattr(fileno $file) ? "answered" : $!{ENOTTY} ? "ENOTTY" : "$!"));
$on_file = ioctl($tty, 0x7fffff01, $buf);
report("a request that is no terminal's: ", (defined $on_file ? "answered" : $!{EINVAL} ? "EINVAL" : "$!"));
report("unopened handle: ", (defined eval { no warnings; ioctl(NOSUCH, $REQUEST{get_size}, $buf) } ? "answered" : $!{EBADF} ? "EBADF" : "$!"));
`)
	if len(got) == 1 && got[0] == "no pseudo-terminal" {
		t.Skip("the system gives no pseudo-terminal")
	}
	wantLines(t, got,
		"window: 30x100 returned=0 but true buffer=256",
		"a terminal: -t=1 isatty=1 ttyname=its path ctermid=/dev/tty",
		"echo at first: on",
		"after setattr: echo off canonical off VMIN=3 VTIME=7 speed B1200",
		"typed: read 'abc', echoed: no",
		"tcflush: done, tcdrain: done",
		"the window of a file: ENOTTY",
		"the settings of a file: ENOTTY",
		"a request that is no terminal's: EINVAL",
		"unopened handle: EBADF",
	)
}

// A forked process that starts a session has no controlling terminal until
// it takes one; then /dev/tty is that terminal for it, and for the command
// it becomes. The process it was forked from keeps its own throughout.
func TestForkedProcessTakesAControllingTerminal(t *testing.T) {
	got := runForkScript(t, ptyPrelude+`
my ($master, $path) = open_pty();
if (!$master) { report("no pseudo-terminal"); exit 0 }
open(my $tty, '+<', $path) or die "open $path: $!";

# In the forked process itself.
my $pid = fork // die "fork: $!";
if (!$pid) {
    my $before = ioctl($tty, $REQUEST{take}, 0);
    report("child: taking a terminal without a session: ", (defined $before ? "done" : $!{EPERM} ? "EPERM" : "$!"));
    POSIX::setsid() > 0 or die "setsid: $!";
    report("child: /dev/tty in a new session: ", (open(my $none, '+<', '/dev/tty') ? "opened" : $!{ENXIO} ? "ENXIO" : "$!"));
    defined ioctl($tty, $REQUEST{take}, 0) or die "take the terminal: $!";
    open(my $own, '+<', '/dev/tty') or die "open /dev/tty: $!";
    report("child: /dev/tty is the terminal: ", (POSIX::ttyname(fileno $own) eq $path ? "yes" : "no"));
    report("child: tcgetpgrp: ", (defined POSIX::tcgetpgrp(fileno $own) && POSIX::tcgetpgrp(fileno $own) >= 0 ? "answered" : $!{ENOSYS} ? "ENOSYS" : "$!"));
    syswrite($own, "from the child\n");
    exit 0;
}
report("parent: read '", from_master($master), "'");
waitpid($pid, 0);
report("parent: child status ", $? >> 8);

# In the command the forked process becomes.
$pid = fork // die "fork: $!";
if (!$pid) {
    POSIX::setsid() > 0 or die "setsid: $!";
    defined ioctl($tty, $REQUEST{take}, 0) or die "take the terminal: $!";
    open(STDIN, '<&', $tty) or die; open(STDOUT, '>&', $tty) or die; open(STDERR, '>&', $tty) or die;
    close $tty; close $master;
    exec '/bin/sh', '-c', 'echo from the command > /dev/tty';
    exit 127;
}
report("parent: read '", from_master($master), "'");
waitpid($pid, 0);
report("parent: command status ", $? >> 8);
`)
	if len(got) == 1 && got[0] == "no pseudo-terminal" {
		t.Skip("the system gives no pseudo-terminal")
	}
	wantLines(t, got,
		"child: taking a terminal without a session: EPERM",
		"child: /dev/tty in a new session: ENXIO",
		"child: /dev/tty is the terminal: yes",
		"child: tcgetpgrp: ENOSYS",
		"parent: read 'from the child'",
		"parent: child status 0",
		"parent: read 'from the command'",
		"parent: command status 0",
	)
}

// A program whose standard output is a terminal writes it line by line, as
// perl does: a prompt is there before the program reads the answer. (The
// interpreter a program runs in may be a copy of one made before the
// program's streams were known; what its standard handles are is taken
// from the descriptors the program has.)
func TestProgramOnATerminalWritesLineByLine(t *testing.T) {
	t.Setenv("TEST_PERL", os.Args[0])
	out, status := hostPerl(t, "-e", ptyPrelude+`
my ($master, $path) = open_pty();
if (!$master) { print "no pseudo-terminal\n"; exit 0 }
my $pid = fork // die "fork: $!";
if (!$pid) {
    POSIX::setsid() > 0 or die "setsid: $!";
    open(my $tty, '+<', $path) or die "open $path: $!";
    defined ioctl($tty, $REQUEST{take}, 0) or die "take the terminal: $!";
    open(STDIN, '<&', $tty) or die; open(STDOUT, '>&', $tty) or die; open(STDERR, '>&', $tty) or die;
    close $tty; close $master;
    $ENV{GPERL_TEST_RUN_AS_PERL} = 1;
    exec $ENV{TEST_PERL}, '-e', 'print "prompt\n"; my $answer = <STDIN>; print "got $answer"; sleep 600';
    exit 127;
}
print "before the answer: ", from_master($master), "\n";
syswrite($master, "forty-two\n");
print "the terminal's echo: ", from_master($master), "\n";
print "after the answer: ", from_master($master), "\n";
kill 'KILL', $pid;
waitpid($pid, 0);
`)
	if strings.TrimSpace(out) == "no pseudo-terminal" {
		t.Skip("the system gives no pseudo-terminal")
	}
	want := "before the answer: prompt\nthe terminal's echo: forty-two\nafter the answer: got forty-two\n"
	if status != 0 || out != want {
		t.Fatalf("status %d, output %q, want %q", status, out, want)
	}
}
