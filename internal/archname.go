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

// archSystem is the system as Configure names it ($^O).
func archSystem(goos string) string {
	if goos == "illumos" {
		return "solaris"
	}
	return goos
}
