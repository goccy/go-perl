package gperl_test

import (
	"testing"

	"github.com/goccy/go-perl/internal/gperl"
)

// runInThisProcess runs a program on gperl's interpreter in this process,
// as gperl.Run does. The program is this process (the host process): what
// it does to the standard descriptors — a close of standard output, a file
// put in its place, the close of every descriptor when it execs a command
// — it does to the test process's, and that stays when the program ends.
// The descriptors are put back as they were once it has.
func runInThisProcess(tb testing.TB, script string, args []string) (int, error) {
	tb.Helper()
	defer keepStandardDescriptors(tb)()
	return gperl.Run(script, args)
}
