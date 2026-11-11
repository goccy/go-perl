//go:build darwin || linux

package gperl_test

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/goccy/go-perl/gperl"
)

// A lock taken with flock is the operating system's, so it holds against
// every other process on the machine and theirs hold against the program.
// The test's own open of the file plays the other process: to the kernel a
// lock belongs to an open file, whoever opened it.
func TestFlockIsSharedWithOtherProcesses(t *testing.T) {
	dir := t.TempDir()
	lockFile := filepath.Join(dir, "lock")
	outside, err := os.OpenFile(lockFile, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	defer outside.Close()
	if err := syscall.Flock(int(outside.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		t.Fatalf("locking the file outside the program: %v", err)
	}

	// While the lock is held outside, the program is refused; once it is
	// released the program takes it and keeps the file open, so the lock
	// stays in place until the program ends.
	script := writeScript(t, "lock.pl", `
use strict; use warnings; use Fcntl qw(:flock);
my ($file, $report, $released) = @ARGV;
sub report { open(my $fh, '>>', $report) or die "report: $!"; syswrite($fh, "@_\n"); close $fh }
open(my $fh, '+<', $file) or die "open: $!";
report("held outside:", (flock($fh, LOCK_EX | LOCK_NB) ? "granted" : ($!{EWOULDBLOCK} ? "would block" : "error $!")));
open(my $signal, '>', $released) or die "signal: $!"; close $signal;
flock($fh, LOCK_EX) or die "flock: $!";
report("released outside: granted");
exit 0;
`)
	report := filepath.Join(dir, "report.txt")
	released := filepath.Join(dir, "refused")
	unlocked := make(chan error, 1)
	go func() {
		// The program creates this file after its refused attempt and
		// then blocks in flock until the lock here is released.
		for {
			if _, err := os.Stat(released); err == nil {
				break
			}
			time.Sleep(time.Millisecond)
		}
		unlocked <- syscall.Flock(int(outside.Fd()), syscall.LOCK_UN)
	}()
	status, err := gperl.Run(script, []string{lockFile, report, released})
	if err != nil || status != 0 {
		t.Fatalf("Run = (%d, %v), want (0, nil)", status, err)
	}
	if err := <-unlocked; err != nil {
		t.Fatalf("unlocking outside the program: %v", err)
	}
	raw, err := os.ReadFile(report)
	if err != nil {
		t.Fatalf("no report written: %v", err)
	}
	const want = "held outside: would block\nreleased outside: granted\n"
	if string(raw) != want {
		t.Fatalf("report:\n%s\nwant:\n%s", raw, want)
	}

	// The program has ended and with it its lock.
	if err := syscall.Flock(int(outside.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		t.Fatalf("locking the file after the program ended: %v", err)
	}
}
