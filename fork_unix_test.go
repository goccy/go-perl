//go:build unix

package perl_test

import (
	"syscall"
	"testing"

	perl "github.com/goccy/go-perl"
)

// A program may run under a limit on the size of the files the process
// makes (RLIMIT_FSIZE, which setrlimit or a shell's ulimit -f lowers) far
// below the length of its memory's mapping, which is the ceiling of the
// memory and more: it forks all the same. The image of the memory a fork
// makes is a file, and the limit applies to it: fork failed when the image
// would have been longer than the limit allows, and the process was sent
// SIGXFSZ.
func TestForkUnderAFileSizeLimit(t *testing.T) {
	var saved syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_FSIZE, &saved); err != nil {
		t.Skipf("getrlimit: %v", err)
	}
	lim := saved
	lim.Cur = 1 << 30
	if lim.Cur > saved.Cur {
		t.Skipf("the file size limit is already below 1 GiB (%d)", saved.Cur)
	}
	if err := syscall.Setrlimit(syscall.RLIMIT_FSIZE, &lim); err != nil {
		t.Skipf("setrlimit: %v", err)
	}
	defer syscall.Setrlimit(syscall.RLIMIT_FSIZE, &saved)
	r := runProgram(t, perl.Config{Fork: true}, perl.Program{Code: code(
		`my @kids;`,
		`for my $n (1 .. 2) {`,
		`    my $pid = fork // die "fork $n: $!\n";`,
		`    if (!$pid) { my $pid2 = fork // die "fork in the child: $!\n"; if (!$pid2) { exit 7 } waitpid($pid2, 0); exit($n + ($? >> 8)) }`,
		`    push @kids, $pid;`,
		`}`,
		`for my $pid (@kids) { waitpid($pid, 0); print "status ", $? >> 8, "\n" }`,
	)}, "")
	r.want(t, 0, "status 8\nstatus 9\n")
}
