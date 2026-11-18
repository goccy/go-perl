package internal

import "runtime"

// ArchName is the name of the architecture the interpreter reports
// ($Config{archname}) and names its architecture directories after: the
// architecture subdirectories perl adds to @INC with a directory
// (PERL5LIB, -I) and the site's sitearch. It is the host's processor and
// system, spelled as perl's Configure spells them on that host, followed
// by "-goperl": native modules are built for the host, and the mark keeps
// go-perl's architecture directories apart from those of a perl of the
// same host, whose native modules go-perl cannot load. One interpreter runs
// on every host; it is told the name when it starts (perl_new).
func ArchName() string { return archName(runtime.GOOS, runtime.GOARCH) }

func archName(goos, goarch string) string {
	if goos == "windows" {
		// perl names a Windows build by the system first
		cpu := map[string]string{"amd64": "x64", "386": "x86", "arm64": "ARM64"}[goarch]
		if cpu == "" {
			cpu = goarch
		}
		return "MSWin32-" + cpu + "-goperl"
	}
	return archProcessor(goos, goarch) + "-" + archSystem(goos) + "-goperl"
}

// archProcessor is the processor as uname -m reports it on the system,
// which is how Configure spells it in the name of the architecture.
func archProcessor(goos, goarch string) string {
	switch goos {
	case "linux":
		switch goarch {
		case "amd64":
			return "x86_64"
		case "arm64":
			return "aarch64"
		case "386":
			return "i686"
		case "arm":
			return "armv7l"
		case "mips64", "mips64le":
			return "mips64"
		case "mips", "mipsle":
			return "mips"
		case "loong64":
			return "loongarch64"
		}
	case "darwin":
		if goarch == "amd64" {
			return "x86_64"
		}
	case "netbsd":
		switch goarch {
		case "amd64":
			return "x86_64"
		case "arm64":
			return "aarch64"
		}
	case "freebsd", "openbsd", "dragonfly":
		if goarch == "386" {
			return "i386"
		}
	}
	// the other processors have Go's name on these systems (ppc64le,
	// s390x, riscv64; arm64 on macOS and on FreeBSD and OpenBSD; amd64 on
	// FreeBSD, OpenBSD and DragonFly)
	return goarch
}

// OSName is the name of the system the interpreter reports ($^O and
// $Config{osname}): the host's, spelled as perl's Configure spells it on
// that host, as perl there reports it. One interpreter runs on every host;
// it is told the name when it starts (perl_new).
func OSName() string { return archSystem(runtime.GOOS) }

// archSystem is the system as Configure names it ($^O).
func archSystem(goos string) string {
	if goos == "illumos" {
		return "solaris"
	}
	return goos
}

// HostABI is the word of facts of the host's C ABI that native code has and
// the interpreter, built for wasm, does not (perl-wasm's g_hostwords and
// GOPERL_HOSTABI_*): the interpreter is told them before any code runs.
func HostABI() uint32 { return hostABI(runtime.GOOS, runtime.GOARCH) }

// hostABICharUnsigned is the bit of HostABI saying that char is unsigned.
const hostABICharUnsigned = 1

func hostABI(goos, goarch string) uint32 {
	var w uint32
	if charUnsigned(goos, goarch) {
		w |= hostABICharUnsigned
	}
	return w
}

// charUnsigned says whether plain char is unsigned in the C ABI of the
// platform: so the psABIs of arm, arm64, ppc64, s390x and riscv64 have it,
// except where the system chose otherwise (Apple's and Microsoft's arm64,
// OpenBSD's); char is signed on x86 and the other processors.
func charUnsigned(goos, goarch string) bool {
	switch goarch {
	case "arm", "arm64", "ppc64", "ppc64le", "s390x", "riscv64":
		switch goos {
		case "darwin", "ios", "windows", "openbsd":
			return false
		}
		return true
	}
	return false
}
