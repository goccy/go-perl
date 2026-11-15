package perl

// Config configures a new Perl instance's sandbox.

import (
	"io"

	"github.com/goccy/go-perl/internal"
)

// Config configures a new Perl instance created by New. The zero value is a
// usable, FULLY SANDBOXED interpreter: a private in-memory filesystem
// pre-loaded with the embedded standard library, an empty environment (the
// host process env is not leaked), sandboxed stdio, and every capability
// hook denying — no outbound connections, no name resolution, no subprocess
// spawns, no fork, no system logger, and an identity of its own instead of
// the host process's.
// Each capability is granted explicitly by setting its field.
// HostMemoryLimit, as Config.MaxMemoryBytes, caps the instance's linear
// memory at the memory the host gives the process: the machine's memory
// (darwin hw.memsize; FreeBSD and DragonFly hw.physmem; OpenBSD and NetBSD
// hw.physmem64; Linux MemTotal of /proc/meminfo), or the memory limit of
// the control group (container) the process is in when that is lower
// (Linux: cgroup v2 memory.max, v1 memory.limit_in_bytes). Where neither
// can be read (other systems), the limit is the default, 8 GiB.
const HostMemoryLimit = internal.HostMemoryLimit

type Config struct {
	// StdlibDir is the path INSIDE the instance's filesystem (Config.FS)
	// holding the Perl standard library (the assembled lib/ tree). It is
	// the interpreter's own library directory: in @INC it stands where
	// perl has its built-in library directories — after the directories
	// the instance is given (AddInc, Program.Inc) and after PERL5LIB from
	// Env, so a module found there shadows the standard library's copy.
	// PERL5OPT in Env is not applied when the instance is created: it
	// belongs to a perl command line, which the host interprets (gperl
	// does). Empty defaults to "/" — right for the
	// library default and for NewStdlibMemFS. An instance on the host
	// filesystem (fs.NewHostFS) points this at a host directory, e.g. the
	// one ExtractStdlib returns.
	StdlibDir string
	// SiteDir is the path INSIDE the instance's filesystem under which the
	// interpreter has its site directories: where the modules installed
	// for it go (`perl Makefile.PL && make install`, `./Build install`) and
	// where it finds them again.
	//
	//	SiteDir/site_perl/<version>          modules (sitelib)
	//	SiteDir/site_perl/<version>/<arch>   architecture-specific files,
	//	                                     native modules among them
	//	                                     (sitearch; <arch> names the
	//	                                     machine this process runs on)
	//	SiteDir/bin                          programs (sitebin, sitescript)
	//	SiteDir/man/man1, SiteDir/man/man3   manual pages
	//
	// %Config reports them under every key that names a site directory,
	// and sitearch and sitelib are in @INC where perl has them: after the
	// directories the instance is given and PERL5LIB, before the standard
	// library. The directories are not created: an installation creates
	// what it needs.
	//
	// Empty, the default, is an interpreter without site directories:
	// @INC has none, and %Config keeps the directories of the build's
	// configuration (/usr/lib/perl5/site_perl/...), where nothing is.
	// gperl names a directory of the user's (HostConfig).
	SiteDir string
	// Env is the environment the guest sees (%ENV). nil means an empty
	// environment — the host process os.Environ() is NOT leaked.
	Env []string
	// Dial is the outbound-connection whitelist. It is called before each
	// connect with the network ("tcp"), the host NAME the guest asked for
	// (empty when it connected to a literal address), the resolved
	// dotted-quad IP, and the port; returning false denies the connection.
	// Receiving both the name and the IP lets policy match them jointly, and
	// the connection is made to the exact approved IP (no DNS rebinding).
	// FAIL-CLOSED: a nil Dial denies ALL outbound connections.
	Dial func(network, host, ip string, port int) bool
	// Resolve is the name-resolution whitelist. It is called with the host
	// being resolved before each lookup; returning false denies it.
	// FAIL-CLOSED: a nil Resolve denies ALL lookups (so a hostname
	// connection needs both Resolve and Dial; a literal-IP connection needs
	// only Dial).
	Resolve func(host string) bool
	// Stdin, when non-nil, backs the guest's fd 0 (STDIN). Defaults to an
	// empty stream (the host process stdin is NOT used).
	Stdin io.Reader
	// Stdout, when non-nil, receives the guest's fd 1 writes (os-level
	// stdout). Note: Perl-level print output is also captured into
	// Result.Stdout by the bridge; Stdout here is the raw fd sink.
	Stdout io.Writer
	// Stderr, when non-nil, receives the guest's fd 2 writes.
	Stderr io.Writer
	// MaxMemoryBytes, when > 0, caps this instance's wasm linear memory.
	// A guest allocation that would grow memory past this limit fails
	// (memory.grow returns -1) instead of growing the host process unbounded.
	// Rounded down to a multiple of the 64 KiB wasm page size; values below
	// the module's initial memory are ignored. HostMemoryLimit caps it at
	// the host's memory. Without one the limit is 8 GiB. The memory is
	// reserved as address space of the limit's size, not resident memory:
	// pages the program never writes cost nothing.
	MaxMemoryBytes int
	// MemoryReserveBytes, when > 0, is the initial linear-memory slice
	// capacity reserved for this instance. Reserving capacity up front makes
	// boot-time grows zero-copy reslices, dropping a freshly-booted
	// instance's resident memory. The reservation is virtual address space,
	// not resident memory. Only consulted when the instance boots privately
	// and its memory is an allocation of the Go heap (a host without mmap,
	// or a 32-bit one); on other hosts the memory is reserved address space
	// as long as the MaxMemoryBytes limit, however the instance boots, and
	// never moves. When 0, a default headroom is used.
	MemoryReserveBytes int
	// Exec is the subprocess whitelist. It is called before every process
	// spawn with the executable path and full argv; returning false denies
	// the spawn. Spawning runs a HOST binary, so a sandbox that allows
	// subprocess should set this to a strict allow-list. FAIL-CLOSED: a nil
	// Exec denies ALL spawns.
	Exec func(path string, argv []string) bool
	// Fork lets the guest fork. A forked child is a pseudo-process: a copy of
	// the whole instance (its memory, and a descriptor table sharing the
	// parent's open files) that runs concurrently inside this host process,
	// the way perl emulates fork on Windows. The child is subject to the
	// same Config as its parent, and exec in a child goes through Exec like
	// any other spawn. FAIL-CLOSED: when false, fork fails in the guest.
	//
	// The pseudo-processes live as long as the instance they were forked
	// from: Close ends the ones that are still running, as SIGKILL would
	// (no END block and no destructor of theirs runs), kills the commands
	// they started and closes their descriptors. It does not wait for them.
	Fork bool
	// HostProcess declares that the program IS the host process, the way a
	// command-line tool built around one instance is:
	//
	//   - $$ and getppid are the host process's;
	//   - the user and group ids ($<, $>, $(, $)) and the user and group
	//     databases (getpwuid, getgrnam, ...) are the host's;
	//   - a signal the host process receives reaches the program when the
	//     program handles it (%SIG), is ignored when the program ignores it,
	//     and otherwise does to the host process what that signal does;
	//   - a signal the program dies of ends the run with a *SignalError, for
	//     the embedder to end the host process the same way;
	//   - kill reaches the host's processes, and a command the program
	//     starts is known to it by its own pid;
	//   - the process group and session of the program are the host
	//     process's (getpgrp, setpgrp, POSIX::setsid), and so are its
	//     scheduling priority, its CPU times and the name of the system it
	//     runs on (POSIX::uname, Sys::Hostname);
	//   - setting the user or group ids is the host's to allow;
	//   - the host process follows the program's working directory and
	//     environment (chdir, %ENV), so that native code running for the
	//     program finds them;
	//   - the program may listen for network connections (bind, listen);
	//   - local time is the host's when the program names no time zone (the
	//     TZ variable is not set);
	//   - the numbers the program sees are the host operating system's
	//     (darwin and linux): the error numbers ($!, %!, Errno, POSIX, the
	//     exit status of a die), the signal numbers (%Config, POSIX, kill by
	//     number, the signal in $?), the flags of open and fcntl, the bits
	//     of poll, and the numbers of sockets (address families, types,
	//     levels, options, message flags, the flags and errors of
	//     getaddrinfo), as constants, as arguments and as results. Native
	//     code the program loads has the same numbers.
	//
	// When false (the default) the program is an isolated process: pid 42,
	// parent 1, user and group 0, empty user and group databases, kill
	// reaches only the processes the program started, binding a socket
	// to a local address fails with EACCES, local time is UTC unless
	// the program sets TZ, and the numbers of errors, signals and flags
	// are those of the interpreter's own C library, the same on every
	// host. A pseudo-process forked from the program has the same user
	// and group but is never the host process: its pid is its own, a
	// signal ends only that pseudo-process, it may form a process group or
	// a session of its own, and the host process does not follow its
	// working directory or environment — native code it calls sees the
	// program's.
	HostProcess bool
	// InheritDescriptors gives the program the descriptors above the
	// standard three that the host process was started with, under their
	// numbers: started as `prog 5<file`, the program finds descriptor 5
	// open and may read it or duplicate it (open(STDIN, "<&5")), as perl
	// does. An inherited socket is the socket it is (a listening socket
	// handed over by a supervisor accepts connections). The descriptors
	// become the program's and are closed when it closes them, so this is
	// for a program that is the host process (HostProcess) and nothing
	// else in the process uses them. They exist once: the first instance
	// created with InheritDescriptors gets them. Descriptor 3 is the
	// interpreter's own (the root directory it opens paths from) and is
	// not handed over.
	// FAIL-CLOSED: when false, descriptors the process inherited are out
	// of the program's reach.
	InheritDescriptors bool
	// Syslog lets the program log to the host's system logger. The
	// interpreter has no log daemon of its own: when Syslog is set, what the
	// program sends through syslog(3) — the "native" mechanism of
	// Sys::Syslog, the one that module tries first — is delivered to the
	// system logger of the host (the local log daemon, on a Unix host), under
	// the facility, level and ident the program chose. A host without a
	// reachable system logger reports the mechanism unavailable.
	// FAIL-CLOSED: when false, the program has no system logger; Sys::Syslog
	// sees its native mechanism as unavailable and goes on to its other
	// mechanisms (a log server over TCP or UDP, a log socket), each subject
	// to the capability it needs, or fails in openlog. A message is never
	// reported as logged without having been handed to a logger.
	Syslog bool
	// NativeModules lets Perl code load native XS modules the way perl
	// loads any dynamic module: XSLoader and DynaLoader find
	// auto/Module/Module.so on @INC and load it into the HOST process (the
	// module must be built against the go-perl XS SDK, see the xs package,
	// which has to be linked into the binary). Native code runs outside
	// the sandbox entirely, so this is for instances that are not
	// sandboxed in the first place. The path is opened on the host, which
	// is the same file only under a host filesystem backend (fs.NewHostFS).
	// FAIL-CLOSED: when false, only modules the embedder registered with
	// AddXSDir load.
	NativeModules bool
	// FS, when non-nil, is the filesystem backend this instance sees as its
	// entire guest filesystem ("/"). Every file operation is routed to it, so
	// giving two instances separate FS values isolates them completely.
	//
	// The go-perl/fs package provides the backends: fs.NewMemFS (private
	// in-memory), fs.NewHostFS (pass-through to the operating system's
	// filesystem - how the perl command behaves, and what gperl uses),
	// fs.DirFS (the host filesystem scoped to one directory), or any other
	// fs.FS implementation.
	//
	// When nil, the instance gets a PRIVATE in-memory filesystem
	// pre-loaded with the standard library (NewStdlibMemFS), so a library
	// embedding never touches the host disk unless explicitly asked to.
	FS FS
}
