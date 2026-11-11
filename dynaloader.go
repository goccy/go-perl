package perl

// Dynamic loading of native XS modules (Config.NativeModules).
//
// The embedded perl is statically linked: its DynaLoader has no
// dl_load_file, so XSLoader and DynaLoader refuse any module that is not
// compiled in. The xs package can load a module built against the native
// XS SDK into the host process; this file gives DynaLoader the
// functions it loads a module with, backed by that loader. From there the
// stock XSLoader.pm and DynaLoader.pm do what they do on any perl: find
// auto/Module/Module.so on @INC, load it, call its boot function.

import (
	"context"
	"fmt"

	"github.com/goccy/go-perl/internal"
)

// dynaLoaderGlue defines DynaLoader's loading functions over three host
// functions. The host functions return (token) on success and
// (undef, message) on failure; dl_error reports the last message, as
// DynaLoader's does.
const dynaLoaderGlue = `package DynaLoader;
my $error = '';
sub dl_error { $error }
sub dl_load_file {
    my ($lib, $message) = DynaLoader::__goperl_load_file($_[0]);
    $error = $message // '';
    return $lib;
}
sub dl_find_symbol {
    my ($sym, $message) = DynaLoader::__goperl_find_symbol($_[0], $_[1]);
    $error = $message // '';
    return $sym;
}
sub dl_install_xsub {
    DynaLoader::__goperl_install_xsub($_[0], $_[1]);
    no strict 'refs';
    return \&{$_[0]};
}
sub dl_undef_symbols { () }
# A loaded library is never unloaded: the loader resolves its own entry
# points from the modules it has loaded, and clones and fork children of an
# interpreter share its library handles. The call therefore fails, the way
# dl_unload_file reports any failure: a false result and a message for
# dl_error.
sub dl_unload_file {
    $error = 'dl_unload_file: a loaded library is never unloaded';
    return 0;
}
1;
`

// The host functions behind dynaLoaderGlue, in the order of dynaLoaderIDs.
var dynaLoaderNames = [...]string{
	"DynaLoader::__goperl_load_file",
	"DynaLoader::__goperl_find_symbol",
	"DynaLoader::__goperl_install_xsub",
}

// installDynaLoader makes native modules loadable through DynaLoader on
// this instance.
func (p *Perl) installDynaLoader() error {
	if _, linked := internal.XSDynamic(); !linked {
		return fmt.Errorf("perl: Config.NativeModules needs native XS support, which is not linked into this binary; import github.com/goccy/go-perl/xs")
	}
	if err := p.ensureDispatcher(); err != nil {
		return err
	}
	// The ids are baked into the guest subs below, and so into every clone
	// and fork child of this instance.
	p.funcsMu.Lock()
	for i := range p.dynaLoaderIDs {
		p.nextFuncID++
		p.dynaLoaderIDs[i] = p.nextFuncID
	}
	p.funcsMu.Unlock()
	p.bindDynaLoader()

	src := dynaLoaderGlue
	for i, name := range dynaLoaderNames {
		src = fmt.Sprintf("sub %s { main::__plwasm_go_call(%d, @_) }\n", name, p.dynaLoaderIDs[i]) + src
	}
	r, err := p.Eval(context.Background(), src)
	if err != nil {
		return err
	}
	if r.Error != nil {
		return fmt.Errorf("perl: install DynaLoader's loading functions: %w", r.Error)
	}
	return nil
}

// bindDynaLoader points the DynaLoader function ids at THIS instance's
// loader state. A clone or fork child inherits the ids with its memory but
// has loader state of its own, so each instance binds its own functions.
func (p *Perl) bindDynaLoader() {
	if p.dynaLoaderIDs[0] == 0 {
		return
	}
	loader, _ := internal.XSDynamic()
	failure := func(err error) ([]Value, error) {
		return []Value{Undef(), ValueOf(err.Error())}, nil
	}
	scalar := func(args []Value, i int) ScalarValue {
		if i < len(args) {
			if s, ok := args[i].(ScalarValue); ok {
				return s
			}
		}
		return Undef()
	}
	fns := [...]GoFunc{
		func(args []Value) ([]Value, error) {
			lib, err := loader.Open(p.raw, scalar(args, 0).String())
			if err != nil {
				return failure(err)
			}
			return []Value{ValueOf(int64(lib))}, nil
		},
		func(args []Value) ([]Value, error) {
			sym, err := loader.Symbol(p.raw, uint64(scalar(args, 0).Int()), scalar(args, 1).String())
			if err != nil {
				return failure(err)
			}
			return []Value{ValueOf(int64(sym))}, nil
		},
		func(args []Value) ([]Value, error) {
			return nil, loader.Install(p.raw, scalar(args, 0).String(), uint64(scalar(args, 1).Int()))
		},
	}
	p.funcsMu.Lock()
	if p.funcs == nil {
		p.funcs = map[int32]GoFunc{}
	}
	for i, id := range p.dynaLoaderIDs {
		p.funcs[id] = fns[i]
	}
	p.funcsMu.Unlock()
}
