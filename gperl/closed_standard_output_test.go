//go:build darwin || linux

package gperl_test

import (
	"io"
	"testing"
	"time"
)

// A program that closes its standard output and standard error ends them
// for the reader at the other end, as perl does: the reader is at the end
// of input while the program still runs (here, waiting for a line of
// input), and nothing of gperl's keeps the pipes open.
func TestClosedStandardOutputEndsTheReadersInput(t *testing.T) {
	cmd := perlProcess("-e", `
print "out\n"; print STDERR "err\n";
close STDOUT; close STDERR;
my $line = <STDIN>;
`)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer cmd.Wait()
	defer stdin.Close()
	for _, s := range []struct {
		name string
		r    io.Reader
		want string
	}{{"standard output", stdout, "out\n"}, {"standard error", stderr, "err\n"}} {
		got := make(chan string, 1)
		go func() {
			b, _ := io.ReadAll(s.r)
			got <- string(b)
		}()
		select {
		case b := <-got:
			if b != s.want {
				t.Errorf("%s: read %q, want %q", s.name, b, s.want)
			}
		case <-time.After(20 * time.Second):
			t.Fatalf("%s: no end of input while the program waits after closing it", s.name)
		}
	}
}
