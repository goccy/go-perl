package perl_test

import (
	"runtime"
	"testing"

	perl "github.com/goccy/go-perl"
)

// POSIX's CHAR_MIN and CHAR_MAX are the host's, which native code has: perl
// on aarch64 Linux, where char is unsigned, says 0 and 255 (FFI::Raw's tests
// pass a char at CHAR_MIN to native code and back); on x86 and on macOS,
// where it is signed, -128 and 127. The interpreter, built for wasm, has a
// signed char.
func TestPOSIXCharLimitsAreTheHosts(t *testing.T) {
	want := "-128 127\n"
	if runtime.GOOS == "linux" && runtime.GOARCH == "arm64" {
		want = "0 255\n"
	}
	r := runProgram(t, perl.Config{}, perl.Program{Code: code(
		`use POSIX ();`,
		`print POSIX::CHAR_MIN(), " ", POSIX::CHAR_MAX(), "\n";`,
	)}, "")
	r.want(t, 0, want)
}
