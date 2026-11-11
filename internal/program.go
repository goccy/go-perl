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
// layout) to its end, END blocks included, and returns its exit status.
// Afterwards the instance only has to be closed.
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
		// The program left through the raw _exit (POSIX::_exit, or the end
		// of an exec): nothing of the interpreter may run again, so there
		// is no orderly end left for Close to perform.
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
