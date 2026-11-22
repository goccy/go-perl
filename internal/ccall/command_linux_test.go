package ccall

import (
	"fmt"
	"runtime"
	"slices"
	"syscall"
)

// machines are the names uname gives the machines of each architecture.
var machines = map[string][]string{
	"386":      {"i386", "i486", "i586", "i686"},
	"amd64":    {"x86_64"},
	"arm":      {"armv6l", "armv7l", "armv8l"},
	"arm64":    {"aarch64"},
	"loong64":  {"loongarch64"},
	"mips64le": {"mips64"},
	"ppc64le":  {"ppc64le"},
	"riscv64":  {"riscv64"},
	"s390x":    {"s390x"},
}

// compileCommand returns the command line that runs the C compiler cc with
// args for the test binary's architecture. The compiler builds for the
// machine it runs on: cc itself when that is the test's architecture, and
// otherwise (a test binary that runs translated) errOtherArchitecture.
func compileCommand(cc string, args []string) (string, []string, error) {
	var name syscall.Utsname
	if err := syscall.Uname(&name); err != nil {
		return "", nil, err
	}
	// Machine is an array of int8 on some architectures, which no function
	// of bytes or strings takes.
	var machine []byte
	for _, c := range name.Machine {
		if c == 0 {
			break
		}
		machine = append(machine, byte(c))
	}
	if !slices.Contains(machines[runtime.GOARCH], string(machine)) {
		return "", nil, fmt.Errorf("%w: %s on a %s machine", errOtherArchitecture, runtime.GOARCH, machine)
	}
	return cc, args, nil
}
