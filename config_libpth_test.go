package perl_test

import (
	"testing"

	perl "github.com/goccy/go-perl"
)

// %Config's library search path, which DynaLoader starts its own from,
// holds the host's conventional library directories — no directory of the
// machine the interpreter was built on.
func TestConfigLibraryPathIsNotTheBuildMachines(t *testing.T) {
	r := runProgram(t, perl.Config{}, perl.Program{Code: code(
		`use Config; use DynaLoader;`,
		`print "libpth=$Config{libpth}\n";`,
		`print "dl_library_path=@DynaLoader::dl_library_path\n";`,
	)}, "")
	r.want(t, 0, "libpth=/usr/local/lib /usr/lib /lib\n"+
		"dl_library_path=/usr/local/lib /usr/lib /lib\n")
}
