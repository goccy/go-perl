package perl

// Config configures a new Perl instance's sandbox.

import "io"

// Config configures a new Perl instance created by New. The zero value is a
// usable, FULLY SANDBOXED interpreter: a private in-memory filesystem
// pre-loaded with the embedded standard library, an empty environment (the
// host process env is not leaked), sandboxed stdio, and every capability
// hook denying — no outbound connections, no name resolution, no subprocess
// spawns, no fork, and an identity of its own instead of the host process's.
// Each capability is granted explicitly by setting its field.
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
	// the module's initial memory are ignored.
	MaxMemoryBytes int
	// MemoryReserveBytes, when > 0, is the initial linear-memory slice
	// capacity reserved for this instance. Reserving capacity up front makes
	// boot-time grows zero-copy reslices, dropping a freshly-booted
	// instance's resident memory. The reservation is virtual address space,
	// not resident memory. Only consulted when the instance boots privately;
	// an instance built on the copy-on-write process snapshot maps the
	// snapshot instead. When 0, a default headroom is used.
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
	//   - kill reaches the host's processes.
	//
	// When false (the default) the program is an isolated process: pid 42,
	// parent 1, user and group 0, empty user and group databases, and kill
	// reaches only the processes the program started. A pseudo-process forked from the program has the same user
	// and group but is never the host process: its pid is its own, and a
	// signal ends only that pseudo-process.
	HostProcess bool
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
