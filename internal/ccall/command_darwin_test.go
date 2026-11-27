package ccall

import (
	"errors"
	"runtime"
	"syscall"
)

// compileCommand returns the command line that runs the C compiler cc with
// args for the test binary's architecture, amd64 or arm64. A test binary
// for amd64 runs translated (Rosetta) on an arm64 Mac, and so would the
// compiler it starts: arch(1) starts cc as an arm64 program then, which
// -arch makes build for amd64. A Mac without Rosetta has no
// sysctl.proc_translated.
func compileCommand(cc string, args []string) (string, []string, error) {
	arch := "arm64"
	if runtime.GOARCH == "amd64" {
		arch = "x86_64"
	}
	args = append([]string{"-arch", arch}, args...)
	translated, err := syscall.SysctlUint32("sysctl.proc_translated")
	if errors.Is(err, syscall.ENOENT) {
		return cc, args, nil
	}
	if err != nil {
		return "", nil, err
	}
	if translated == 0 {
		return cc, args, nil
	}
	return "/usr/bin/arch", append([]string{"-arm64", cc}, args...), nil
}
