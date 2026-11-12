package perl_test

import (
	"testing"

	perl "github.com/goccy/go-perl"
)

// A module that is linked into the interpreter but was never loaded leaves
// no trace: its package has no sub, so nothing takes it for loaded. Loading
// it — through XSLoader or through DynaLoader's bootstrap — installs its
// boot function then, as a dynamically loading perl does.
func TestLinkedInExtensionsAreNotLoadedUntilUsed(t *testing.T) {
	r := runProgram(t, perl.Config{}, perl.Program{Code: code(
		`sub subs_of { my $pkg = shift; no strict "refs"; scalar grep { defined &{"${pkg}::$_"} } keys %{"${pkg}::"} }`,
		`for my $m (qw(Digest::MD5 PadWalker Want Storable Clone)) {`,
		`    print "$m before: subs=", subs_of($m), " bootstrap=", (defined &{"${m}::bootstrap"} ? 1 : 0), "\n";`,
		`}`,
		`require Digest::MD5;   # XSLoader::load`,
		`require PadWalker;     # bootstrap PadWalker, through DynaLoader`,
		`print "md5: ", Digest::MD5::md5_hex("abc"), "\n";`,
		`print "Digest::MD5 after: bootstrap=", (defined &Digest::MD5::bootstrap ? 1 : 0), "\n";`,
		`print "PadWalker after: bootstrap=", (defined &PadWalker::bootstrap ? 1 : 0),`,
		`    " peek_my=", (defined &PadWalker::peek_my ? 1 : 0), "\n";`,
		`print "Want still: subs=", subs_of("Want"), "\n";`,
		`print "not linked in: ", (eval { require XSLoader; XSLoader::load("No::Such::Extension"); 1 } ? "loaded" : "refused"), "\n";`,
	)}, "")
	r.want(t, 0, "Digest::MD5 before: subs=0 bootstrap=0\n"+
		"PadWalker before: subs=0 bootstrap=0\n"+
		"Want before: subs=0 bootstrap=0\n"+
		"Storable before: subs=0 bootstrap=0\n"+
		"Clone before: subs=0 bootstrap=0\n"+
		"md5: 900150983cd24fb0d6963f7d28e17f72\n"+
		"Digest::MD5 after: bootstrap=1\n"+
		"PadWalker after: bootstrap=1 peek_my=1\n"+
		"Want still: subs=0\n"+
		"not linked in: refused\n")
}

// Want and PadWalker each bring a helper named upcontext, with different
// signatures; both modules work in one program.
func TestWantAndPadWalkerTogether(t *testing.T) {
	r := runProgram(t, perl.Config{}, perl.Program{Code: code(
		`use Want;`,
		`use PadWalker qw(peek_my peek_sub);`,
		`sub up { Want::wantarray_up(0) ? "list" : "scalar" }`,
		`my @l = up(); my $s = up();`,
		`print "wantarray_up: $l[0] $s\n";`,
		`sub ctx { want("LIST") ? "LIST" : want("SCALAR") ? "SCALAR" : "VOID" }`,
		`my @a = ctx(); my $b = ctx(); ctx();`,
		`print "want: $a[0] $b\n";`,
		`sub howmany { want("COUNT") }`,
		`my ($x, $y, $z) = howmany();`,
		`print "count: ", (defined $x ? $x : "undef"), "\n";`,
		`sub lv :lvalue { my $r = want(qw(LVALUE ASSIGN)) ? "assigned" : "read"; print "lvalue sub: $r\n"; my $v; $v }`,
		`lv() = 1; my $got = lv();`,
		`sub outer { my $here = 42; inner() }`,
		`sub inner { ${ peek_my(1)->{'$here'} } }`,
		`print "peek_my: ", outer(), "\n";`,
		`my $closed = 7; my $code = sub { $closed + 1 };`,
		`print "peek_sub: ", join(",", sort keys %{ peek_sub($code) }), "\n";`,
	)}, "")
	r.want(t, 0, "wantarray_up: list scalar\n"+
		"want: LIST SCALAR\n"+
		"count: 3\n"+
		"lvalue sub: assigned\nlvalue sub: read\n"+
		"peek_my: 42\n"+
		"peek_sub: $closed\n")
}

// Devel::Caller reads the context frames PadWalker finds, so it is linked
// in with it.
func TestDevelCaller(t *testing.T) {
	r := runProgram(t, perl.Config{}, perl.Program{Code: code(
		`use Devel::Caller qw(caller_cv caller_args called_with called_as_method);`,
		`sub whoami { caller_cv(0) == \&whoami ? "itself" : "other" }`,
		`print "caller_cv: ", whoami(), "\n";`,
		`sub names { join(",", called_with(0, 1)) }`,
		`my ($first, @rest) = (1, 2);`,
		`print "called_with: ", names($first, @rest, "literal"), "\n";`,
		`sub args { join(",", caller_args(0)) }`,
		`print "caller_args: ", args("a", "b"), "\n";`,
		`sub how { called_as_method(0) ? "method" : "function" }`,
		`my $as_method = main->how;`,
		`my $as_function = how();`,
		`print "called_as_method: $as_method $as_function\n";`,
	)}, "")
	r.want(t, 0, "caller_cv: itself\n"+
		"called_with: $first,@rest,\n"+
		"caller_args: a,b\n"+
		"called_as_method: method function\n")
}
