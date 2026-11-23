package perl

// Search-path registration for an instance: module trees on @INC and
// native XS module directories. Code itself loads through Eval (a file is
// `do 'path'` — no separate entry point).

import (
	"context"
	"fmt"

	"github.com/goccy/go-perl/internal"
)

// AddInc prepends dirs to the instance's @INC, so `use`/`require` resolve
// modules from them (a module tree, an application's lib
// directory). Paths cross via the function bridge, so no Perl-source
// quoting is involved.
func (p *Perl) AddInc(ctx context.Context, dirs ...string) error {
	if len(dirs) == 0 {
		return nil
	}
	vals := make([]Value, len(dirs))
	for i, d := range dirs {
		vals[i] = ValueOf(d)
	}
	if err := p.Bind("__goperl_inc_dirs", func([]Value) ([]Value, error) { return vals, nil }); err != nil {
		return fmt.Errorf("bind inc dirs: %w", err)
	}
	if r, err := p.Eval(ctx, `unshift @INC, __goperl_inc_dirs(); 1;`); err != nil {
		return fmt.Errorf("extend @INC: %w", err)
	} else if r.Error != nil {
		return fmt.Errorf("extend @INC: %w", r.Error)
	}
	return nil
}

// AddXSDir registers the native XS modules in dir with the instance, the
// way AddInc registers a module tree. Two layouts are recognized: a perl
// architecture directory, where auto/<Module/Path>/<Name>.so is the module
// Module::Path, and a flat directory of <Module-Name>.so files (the package
// separator spelled "-"). Registration is cheap (each module boots lazily on first `use`), and a
// missing directory simply registers nothing.
//
// The loader itself lives in the go-perl/xs package; using AddXSDir from a
// binary that links neither psgi nor gperl requires importing it
// (`import _ "github.com/goccy/go-perl/xs"`).
func (p *Perl) AddXSDir(dir string) error {
	linked, err := internal.XSDirLoad(p.raw, dir)
	if !linked {
		return fmt.Errorf("perl: native XS support is not linked into this binary; import github.com/goccy/go-perl/xs")
	}
	return err
}
