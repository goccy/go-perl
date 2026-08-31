package perl

// Guest exit handling.

import (
	"github.com/goccy/go-perl/internal"
)

// ExitCode reports whether err (from Eval, Call, or RunFile) is the guest
// terminating itself with a Perl-level exit(), and returns the exit status.
// An embedder running a whole script typically forwards the status to
// os.Exit after Closing the instance — a Call-surfaced exit leaves the
// interpreter cleanly unwound, so Close still flushes PerlIO and runs END
// blocks. (An exit during Eval reaches the host as a raw wasi proc_exit
// instead; that instance should only be Closed.)
func ExitCode(err error) (int, bool) {
	return internal.WasiExitCode(err)
}
