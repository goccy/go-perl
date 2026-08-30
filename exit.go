package perl

// Guest exit handling.

import (
	"errors"
	"fmt"

	"github.com/goccy/perlwasm2go/base"
)

// exitStatusError is a guest exit() the perl_call bridge caught cleanly: the
// guest unwound back to the call frame (mirroring perl_run's own my_exit
// handling), so the interpreter is still flushable and destructible.
type exitStatusError struct{ code int }

func (e *exitStatusError) Error() string { return fmt.Sprintf("perl: exit(%d)", e.code) }

// ExitCode reports whether err (from Eval, Call, or RunFile) is the guest
// terminating itself with a Perl-level exit(), and returns the exit status.
// An embedder running a whole script typically forwards the status to
// os.Exit after Closing the instance — a Call-surfaced exit leaves the
// interpreter cleanly unwound, so Close still flushes PerlIO and runs END
// blocks. (An exit during Eval reaches the host as a raw wasi proc_exit
// instead; that instance should only be Closed.)
func ExitCode(err error) (int, bool) {
	var ese *exitStatusError
	if errors.As(err, &ese) {
		return ese.code, true
	}
	var we *base.WasiExitError
	if errors.As(err, &we) {
		return int(we.Code), true
	}
	return 0, false
}
