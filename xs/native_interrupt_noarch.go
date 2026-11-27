//go:build linux && !(amd64 || arm64 || riscv64 || loong64 || ppc64 || ppc64le || s390x)

package xs

// On these architectures the kernel's struct sigaction has another layout,
// which this package does not describe: native code is not interrupted by
// the program's signals.

func (s *state) nativeThreadSelf() (thread uint64, ok bool) { return 0, false }

func interruptNativeThread(thread uint64) {}
