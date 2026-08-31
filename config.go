package perl

// Config configures a new Perl instance's sandbox.

import "io"

// Config configures a new Perl instance created by New. The zero value is a
// usable interpreter: no host env is leaked and the embedded standard library
// is extracted and used automatically.
type Config struct {
	// StdlibDir is the path INSIDE the instance's filesystem (Config.FS)
	// holding the Perl standard library (the assembled lib/ tree); it
	// becomes @INC via perl_new. Empty defaults to "/" — right for the
	// library default and for NewStdlibMemFS. An instance on the host
	// filesystem (fs.NewHostFS) points this at a host directory, e.g. the
	// one ExtractStdlib returns.
	StdlibDir string
	// Env is the environment the guest sees (%ENV). nil means an empty
	// environment — the host process os.Environ() is NOT leaked.
	Env []string
	// NetAccess, when non-nil, gates the socket accept/recv/send surface.
	// op is "accept"/"recv"/"send"; returning false denies it.
	NetAccess func(op string) bool
	// Dial, when non-nil, is the OUTBOUND-connection whitelist. It is called
	// before each connect with the network ("tcp"), the host NAME the guest
	// asked for (empty when it connected to a literal address), the resolved
	// dotted-quad IP, and the port; returning false denies the connection.
	// Receiving both the name and the IP lets policy match them jointly, and
	// the connection is made to the exact approved IP (no DNS rebinding).
	// When nil, all outbound connections are allowed.
	Dial func(network, host, ip string, port int) bool
	// Resolve, when non-nil, is the name-resolution whitelist. It is called
	// with the host being resolved before each lookup; returning false denies
	// it. When nil, all lookups are allowed.
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
	// Exec, when non-nil, is the subprocess whitelist. It is called before
	// every process spawn with the executable path and full argv; returning
	// false denies the spawn. When nil, all spawns are permitted (subject to
	// host-subprocess being built in).
	Exec func(path string, argv []string) bool
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
