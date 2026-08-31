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
