package gperl

import (
	"net/http"
	"os"
	"path/filepath"
)

// ManagedBin exposes the installed toolchain's tools to the tests.
var ManagedBin = managedBin

// StaticDeps exposes the libraries a static link of a module needs.
var StaticDeps = staticDeps

// RemoveSelfPerl removes the name the test binary has as the perl command:
// a test binary is gone after its run, and its name would point nowhere.
func RemoveSelfPerl() {
	link, _, _, ok := selfPerlPath()
	if !ok {
		return
	}
	if _, err := os.Lstat(link); err == nil {
		os.RemoveAll(filepath.Dir(link))
	}
}

// PSGIServer loads the PSGI application at path with its project's
// dependencies, as gperl psgi run does, and returns its server over two
// workers, and whether the workers load native modules from @INC.
func PSGIServer(path string) (server http.Handler, native bool, closeServer func(), err error) {
	app, err := projectApp(path)
	if err != nil {
		return nil, false, nil, err
	}
	c, _, err := parsePSGI("test", nil, []string{"--workers", "2"}, func(string) string { return "" })
	if err != nil {
		return nil, false, nil, err
	}
	s, err := app.newServer(c, nil)
	if err != nil {
		return nil, false, nil, err
	}
	return s, app.native, s.Close, nil
}
