package internal

// Registration point wiring the xs subpackage's native-module directory
// loader to the public AddXSDir entry, without the public package importing
// the loader (which would drag purego/dlopen into every embedder).

import (
	"fmt"
	"io/fs"
)

// xsDirLoader is installed by the xs package's init.
var xsDirLoader func(p *Perl, dir string) error

// RegisterXSDirLoader wires the native-module directory loader XSDirLoad
// dispatches to. Called by the xs package's init; not for application use.
func RegisterXSDirLoader(fn func(p *Perl, dir string) error) { xsDirLoader = fn }

// XSDirLoad registers the native XS modules under dir with the instance via
// the loader the xs package registered. Reports whether a loader is linked
// into the binary at all.
func XSDirLoad(p *Perl, dir string) (linked bool, err error) {
	if xsDirLoader == nil {
		return false, nil
	}
	p.nativeHoldsAddresses()
	return true, xsDirLoader(p, dir)
}

// XSABI is the native XS SDK's ABI version: it must match GOPERL_XS_ABI in
// the SDK's perl.h, and artifacts built against one version load only into
// it.
const XSABI = 44

// xsLinkedLoader is installed by the xs package's init.
var xsLinkedLoader func(p *Perl) error

// RegisterXSLinkedLoader wires the loader of the native XS modules linked
// into the binary. Called by the xs package's init; not for application use.
func RegisterXSLinkedLoader(fn func(p *Perl) error) { xsLinkedLoader = fn }

// XSLoadLinked registers the native XS modules linked into the binary with
// the instance. Without the xs package there are none.
func XSLoadLinked(p *Perl) error {
	if xsLinkedLoader == nil {
		return nil
	}
	return xsLinkedLoader(p)
}

// xsNative is the tree of loadable native modules a binary carries.
var xsNative fs.FS

// SetXSNative installs a tree of loadable native modules
// (auto/<Module/Path>/<Name>.so) for every new instance to load from.
// Called by the init of a generated go-perl/cpan package built without
// cgo; not for application use.
func SetXSNative(fsys fs.FS) { xsNative = fsys }

// XSNative returns the installed tree, nil when the binary carries none.
func XSNative() fs.FS { return xsNative }

// XSDynamicLoader is the host half of perl's dynamic loading (DynaLoader's
// dl_load_file / dl_find_symbol / dl_install_xsub), provided by the xs
// package. Libraries and symbols are named by tokens that are only
// meaningful to the instance that obtained them.
type XSDynamicLoader struct {
	// Open loads the library at path (a native XS module, or any shared
	// library, as dlopen loads it) and returns its token. flags are
	// DynaLoader's dl_load_file flags (a module's dl_load_flags): 0x01 makes
	// the library's symbols global, available to the libraries loaded
	// after it.
	Open func(p *Perl, path string, flags int) (lib uint64, err error)
	// Symbol resolves name in a loaded module.
	Symbol func(p *Perl, lib uint64, name string) (sym uint64, err error)
	// Install binds the function behind sym as the Perl sub name.
	Install func(p *Perl, name string, sym uint64) error
	// Release gives up what the loader keeps for an instance that closes:
	// its interpreter is destroyed and it is never entered again.
	Release func(p *Perl)
}

// xsDynamicLoader is installed by the xs package's init.
var xsDynamicLoader *XSDynamicLoader

// RegisterXSDynamicLoader wires the loader behind XSDynamic. Called by the
// xs package's init; not for application use. A library that loads may
// change the host process's signal actions as it initializes (ignore
// SIGPIPE, say): the signals the process then ignores are recorded
// (adoptIgnoredSignals).
func RegisterXSDynamicLoader(l XSDynamicLoader) {
	open := l.Open
	l.Open = func(p *Perl, path string, flags int) (uint64, error) {
		lib, err := open(p, path, flags)
		if err != nil || !p.opts.HostProcess {
			return lib, err
		}
		if err := adoptIgnoredSignals(); err != nil {
			// The library is loaded: the failure is the program's Close's.
			p.family.fail(failAdoption, fmt.Errorf("after loading %s: %w", path, err))
		}
		return lib, nil
	}
	xsDynamicLoader = &l
}

// XSDynamic returns the dynamic loader the xs package registered; linked
// is false when the package is not part of the binary.
func XSDynamic() (l XSDynamicLoader, linked bool) {
	if xsDynamicLoader == nil {
		return XSDynamicLoader{}, false
	}
	return *xsDynamicLoader, true
}
