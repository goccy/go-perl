//go:build unix

package perl_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"
	"time"

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

// The program ends while a fork child is inside a call the host cannot end:
// a Go function the program called, which returns only later. The end of
// the program does not wait for it: the child is ended all the same, runs
// no Perl code once the call does return, and its goroutine goes then.
func TestProgramEndsWhileAChildIsInAGoFunction(t *testing.T) {
	release := make(chan struct{})
	var released bool
	defer func() {
		if !released {
			close(release)
		}
	}()
	report := filepath.Join(t.TempDir(), "report")
	cfg := hostConfig(t)
	cfg.Fork = true
	var errOut bytes.Buffer
	cfg.Stderr = &errOut
	p, err := perl.New(cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := p.Bind("main::hold", func([]perl.Value) ([]perl.Value, error) {
		<-release
		return nil, nil
	}); err != nil {
		t.Fatalf("Bind: %v", err)
	}

	type outcome struct {
		status int
		err    error
	}
	ended := make(chan outcome, 1)
	go func() {
		status, err := p.RunProgram(context.Background(), perl.Program{
			Code: code(
				`pipe(my $r, my $w) or die "pipe: $!";`,
				`my $pid = fork // die "fork: $!";`,
				`if (!$pid) {`,
				`    close $r; syswrite($w, "in\n"); close $w;`,
				`    hold();`,
				`    open my $fh, ">", $ARGV[0] or die $!; print $fh "the child went on\n"; close $fh;`,
				`    exit 0;`,
				`}`,
				`close $w; my $in = <$r>; select(undef, undef, undef, 0.3);`,
				`exit 5;`,
			),
			Args: []string{report},
		})
		if err == nil {
			err = p.Close()
		}
		ended <- outcome{status, err}
	}()
	select {
	case got := <-ended:
		if got.err != nil || got.status != 5 {
			t.Fatalf("the program ended with (%d, %v), want (5, nil); stderr %q", got.status, got.err, errOut.String())
		}
	case <-time.After(30 * time.Second):
		t.Fatal("the program's end waits for the child's Go function")
	}
	// The call returns now; the child must not get any further.
	before := runtime.NumGoroutine()
	released = true
	close(release)
	deadline := time.Now().Add(10 * time.Second)
	for runtime.NumGoroutine() >= before && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if runtime.NumGoroutine() >= before {
		t.Errorf("the child's goroutine is still there after its Go function returned")
	}
	if got, err := os.ReadFile(report); err == nil {
		t.Fatalf("the child ran on after the program had ended: %q", got)
	}
}
