//go:build linux || darwin

package gperl_test

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"syscall"
	"testing"
)

// A write into a pipe whose reader has gone raises SIGPIPE as the write
// returns, as the kernel raises it for perl: a program that leaves its
// output for the flush at its end dies of the signal there (Mojo::Run3's
// t/close.t), where gperl said "Unable to flush stdout: Broken pipe" and
// exited 1. A handler of the program's runs right after the print that
// failed, and with the signal ignored the print fails with EPIPE. What is
// expected is what perl 5.36 on Linux does.
func TestWriteIntoAPipeWithoutReaderRaisesSIGPIPE(t *testing.T) {
	for _, tc := range []struct {
		name    string
		program string
		stderr  string
		status  int // the exit status, or -1 for death by SIGPIPE
	}{
		{"default action", `print STDOUT "x"; exit 0`, "", -1},
		{"handler", `$SIG{PIPE} = sub { print STDERR "handler\n" }; $| = 1; print STDOUT "x"; print STDERR "after: ", ($!{EPIPE} ? "EPIPE" : "other"), "\n"; exit 3`,
			"handler\nafter: EPIPE\n", 3},
		{"ignored", `$SIG{PIPE} = "IGNORE"; $| = 1; print STDOUT "x" or print STDERR "print failed: ", ($!{EPIPE} ? "EPIPE" : "other"), "\n"; exit 4`,
			"print failed: EPIPE\n", 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatal(err)
			}
			r.Close() // the reader is gone before the program writes
			defer w.Close()
			var stderr bytes.Buffer
			cmd := exec.Command(os.Args[0], "-e", tc.program)
			cmd.Env = append(os.Environ(), asPerlEnv+"=1")
			cmd.Stdout = w
			cmd.Stderr = &stderr
			err = cmd.Run()
			var exitErr *exec.ExitError
			if err != nil && !errors.As(err, &exitErr) {
				t.Fatal(err)
			}
			ws := cmd.ProcessState.Sys().(syscall.WaitStatus)
			switch {
			case tc.status < 0:
				if !ws.Signaled() || ws.Signal() != syscall.SIGPIPE {
					t.Errorf("status %v, want death by SIGPIPE; stderr %q", cmd.ProcessState, stderr.String())
				}
			case !ws.Exited() || ws.ExitStatus() != tc.status:
				t.Errorf("status %v, want exit %d; stderr %q", cmd.ProcessState, tc.status, stderr.String())
			}
			if stderr.String() != tc.stderr {
				t.Errorf("stderr %q, want %q", stderr.String(), tc.stderr)
			}
		})
	}
}
