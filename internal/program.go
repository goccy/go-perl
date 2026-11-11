package internal

import (
	"context"
	"errors"
	"fmt"
	"math"

	wasm2go "github.com/goccy/perlwasm2go"
	"github.com/goccy/perlwasm2go/base"
)

// RunProgram runs the program desc describes (perl-wasm's perl_run_program
// layout) to its end and returns its exit status. The end is a process's:
// the END blocks, the flush of PerlIO and global destruction all run inside
// the call, and the status is the one they leave (a destructor may assign
// $?, exit, or be killed during global destruction). Afterwards the
// instance holds no interpreter and only has to be closed.
func (p *Perl) RunProgram(ctx context.Context, desc []byte) (status int, err error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	if p.closed.Load() {
		return 0, ErrClosed
	}
	p.drainReleases()

	var buf []byte
	buf = pbAppendUint64(buf, 1, p.h)
	buf = pbAppendBytes(buf, 2, desc)
	buf = pbAppendUint64(buf, 3, uint64(len(desc)))
	disarm := p.armInterrupt(ctx)
	resp, err := p.m.invoke(0, midRunProgram, buf, wasm2go.Inv_0_22)
	interrupted := disarm()

	var killed *base.WasiSignalError
	if errors.As(err, &killed) {
		// A signal ended the program where it stood: like a process killed
		// by a signal, it gets no orderly end.
		p.discard()
		return 0, &SignalError{Signal: int(killed.Signal)}
	}
	var raw *base.WasiExitError
	if errors.As(err, &raw) {
		// The program left through the raw _exit (POSIX::_exit, the end of
		// an exec, or an exit from a destructor during global destruction):
		// nothing of the interpreter may run again.
		p.discard()
		return int(raw.Code), nil
	}
	if err == nil {
		err = pbExtractError(resp)
	}
	if err != nil {
		return 0, err
	}
	code := readScalarAtField(resp, 1, (*pbReader).readUint32)
	if code == math.MaxUint32 {
		return 0, fmt.Errorf("perl: the interpreter rejected the program description")
	}
	if interrupted {
		return int(code), ctx.Err()
	}
	return int(code), nil
}
