//go:build darwin || linux

package gperl_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/goccy/go-perl/gperl"
)

// A program whose #! line names another interpreter is handed to that
// interpreter, with its arguments.
func TestProgramForAnotherInterpreterIsHandedOver(t *testing.T) {
	if _, err := os.Stat("/bin/sh"); err != nil {
		t.Skip("no /bin/sh")
	}
	report := filepath.Join(t.TempDir(), "report.txt")
	script := writeScript(t, "for-sh", "#!/bin/sh\necho \"run by sh as $0 with $# arguments: $1\" > \"$2\"\nexit 6\n")
	status, err := gperl.Run(script, []string{"first", report})
	if err != nil || status != 6 {
		t.Fatalf("Run = (%d, %v), want (6, nil)", status, err)
	}
	raw, err := os.ReadFile(report)
	if err != nil {
		t.Fatalf("the interpreter named on the #! line did not run the program: %v", err)
	}
	if got, want := strings.TrimSpace(string(raw)), "run by sh as "+script+" with 2 arguments: first"; got != want {
		t.Fatalf("report = %q, want %q", got, want)
	}

	// A #! line that names perl stays with perl, switches and all.
	perlScript := writeScript(t, "for-perl", "#!/usr/bin/perl -l\nopen(my $fh, '>', $ARGV[0]) or die; print {$fh} 'run by perl'; close $fh;\n")
	status, err = gperl.Run(perlScript, []string{report})
	if err != nil || status != 0 {
		t.Fatalf("Run = (%d, %v), want (0, nil)", status, err)
	}
	if raw, _ := os.ReadFile(report); string(raw) != "run by perl\n" {
		t.Fatalf("report = %q, want %q", raw, "run by perl\n")
	}
}

// A program handed to the interpreter its #! line names is handed over with
// the command line perl was given, its own switches included, after the
// argument of the #! line.
func TestHandOverPassesTheCommandLine(t *testing.T) {
	shim := perlCommandShim(t)
	dir := t.TempDir()
	report := filepath.Join(dir, "report.txt")
	interp := filepath.Join(dir, "interp")
	if err := os.WriteFile(interp, []byte("#!/bin/sh\nfor a in \"$@\"; do echo \"$a\"; done > \""+report+"\"\nexit 7\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	program := writeScript(t, "handed-over", "#!"+interp+" -from-the-line\nthis is no perl\n")
	script := writeScript(t, "run.pl", `
use strict; use warnings;
my ($perl, $program) = @ARGV;
system($perl, "-w", "-I/some/dir", $program, "first", "second");
exit($? >> 8);
`)
	status, err := gperl.Run(script, []string{shim, program})
	if err != nil || status != 7 {
		t.Fatalf("Run = (%d, %v), want (7, nil)", status, err)
	}
	raw, err := os.ReadFile(report)
	if err != nil {
		t.Fatalf("the interpreter named on the #! line did not run: %v", err)
	}
	want := []string{"-from-the-line", "-w", "-I/some/dir", program, "first", "second"}
	if got := strings.Split(strings.TrimSpace(string(raw)), "\n"); strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("arguments of the interpreter:\n  %s\nwant:\n  %s", strings.Join(got, "\n  "), strings.Join(want, "\n  "))
	}
}
