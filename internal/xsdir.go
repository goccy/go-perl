package internal

// Registration point wiring the xs subpackage's native-module directory
// loader to the public AddXSDir entry, without the public package importing
// the loader (which would drag purego/dlopen into every embedder).

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
	return true, xsDirLoader(p, dir)
}

// XSDynamicLoader is the host half of perl's dynamic loading (DynaLoader's
// dl_load_file / dl_find_symbol / dl_install_xsub), provided by the xs
// package. Libraries and symbols are named by tokens that are only
// meaningful to the instance that obtained them.
type XSDynamicLoader struct {
	// Open loads the native XS module at path and returns its token.
	// flags are DynaLoader's dl_load_file flags (0x01: the library's
	// symbols are global).
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
// xs package's init; not for application use.
func RegisterXSDynamicLoader(l XSDynamicLoader) { xsDynamicLoader = &l }

// XSDynamic returns the dynamic loader the xs package registered; linked
// is false when the package is not part of the binary.
func XSDynamic() (l XSDynamicLoader, linked bool) {
	if xsDynamicLoader == nil {
		return XSDynamicLoader{}, false
	}
	return *xsDynamicLoader, true
}
