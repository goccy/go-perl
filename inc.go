package perl

// Search-path registration for an instance: module trees on @INC and
// native XS module directories. Code itself loads through Eval (a file is
// `do 'path'` — no separate entry point).

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
)

// AddInc prepends dirs to the instance's @INC, so `use`/`require` resolve
// modules from them (a vendored local/lib/perl5 tree, an application's lib
// directory). Paths cross via the function bridge, so no Perl-source
// quoting is involved.
func (p *Perl) AddInc(ctx context.Context, dirs ...string) error {
	if len(dirs) == 0 {
		return nil
	}
	vals := make([]any, len(dirs))
	for i, d := range dirs {
		vals[i] = d
	}
	if err := p.Bind("__goperl_inc_dirs", func([]any) ([]any, error) { return vals, nil }); err != nil {
		return fmt.Errorf("bind inc dirs: %w", err)
	}
	if r, err := p.Eval(ctx, `unshift @INC, __goperl_inc_dirs(); 1;`); err != nil {
		return fmt.Errorf("extend @INC: %w", err)
	} else if r.Error != nil {
		return fmt.Errorf("extend @INC: %w", r.Error)
	}
	return nil
}

// xsDirLoader is installed by the xs package (its init) so AddXSDir works
// without application code importing the loader; the psgi and gperl entry
// points link it in.
var xsDirLoader func(p *Perl, dir string) error

// RegisterXSDirLoader wires the native-module directory loader AddXSDir
// dispatches to. Called by the xs package's init; not for application use.
func RegisterXSDirLoader(fn func(p *Perl, dir string) error) { xsDirLoader = fn }

// AddXSDir registers the native XS modules built under dir — the `gperl xs
// build` output layout, dir/<goos>_<goarch>/<Module-Name>.so — with the
// instance, the way AddInc registers a module tree: pass the xs root
// (typically "local/xs") and the running binary's architecture directory is
// selected underneath it. Registration is cheap (each module boots lazily
// on first `use`), and a missing directory simply registers nothing.
//
// The loader itself lives in the go-perl/xs package; using AddXSDir from a
// binary that links neither psgi nor gperl requires importing it
// (`import _ "github.com/goccy/go-perl/xs"`).
func (p *Perl) AddXSDir(dir string) error {
	if xsDirLoader == nil {
		return fmt.Errorf("perl: native XS support is not linked into this binary; import github.com/goccy/go-perl/xs")
	}
	return xsDirLoader(p, filepath.Join(dir, runtime.GOOS+"_"+runtime.GOARCH))
}
