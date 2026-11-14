//go:build darwin || linux

package xs_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// A string constructor given no string (a NULL pointer) makes an undefined
// SV: newSVpvn(NULL, 0) is how XS code returns undef.
func TestNativeNullStringIsUndef(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"newSVpvn, newSVpv, newSVpvn_flags without a string; an empty string", `join ",", map { $_ ? "defined" : "undef" } PerlAPI::null_strings()`, "undef,undef,undef,defined"},
	})
}

// `local` on a variable with magic of a native module gives the new value
// the magic, as perl does: set magic sees the new value when the scope
// starts, every assignment in it, and the old value coming back; and the
// magic of the variable itself is still there afterwards.
func TestNativeMagicFollowsLocal(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"set magic through a local", `our $v = "a"; PerlAPI::watch($v); $v = "b"; { local $v = "c"; $v = "d" } $v = "e"; PerlAPI::watched()`, "|b|undef|c|d|b|e"},
		{"a local without a value", `our $w = "a"; PerlAPI::watch($w); { local $w; $w = "in" } $w = "out"; PerlAPI::watched()`, "|undef|in|a|out"},
	})
}

// sv_magicext keeps the length it is given: with HEf_SVKEY the name is an
// SV the magic holds a reference to, and gives up with the variable.
func TestNativeMagicNamedByAnSV(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"the length and the reference", `our $key = "slot"; our @holder = ("h"); join ",", PerlAPI::keyed_attach($holder[0], $key)`, "1,-2"},
		{"the key read back", `PerlAPI::keyed_read($holder[0])`, "slot"},
		{"the reference given up", `my $before = Internals::SvREFCNT($key); @holder = (); $before - Internals::SvREFCNT($key)`, "1"},
	})
}

// SvTIED_mg says whether a container is tied and SvTIED_obj to what, so a
// module that treats tied containers apart (to count the keys of a tied
// hash by iterating) does.
func TestNativeTieMagic(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"a plain hash", `my %h = (a => 1); PerlAPI::tied_to(\%h)`, "not tied"},
		{"a tied hash", `require Tie::Hash; tie my %h, "Tie::StdHash"; PerlAPI::tied_to(\%h)`, "Tie::StdHash"},
		{"a tied array", `require Tie::Array; tie my @a, "Tie::StdArray"; PerlAPI::tied_to(\@a)`, "Tie::StdArray"},
	})
}

// An overlong UTF-8 sequence is malformed unless the caller allows it.
func TestNativeOverlongUTF8(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"a well-formed character", `PerlAPI::decode_first("\xc3\xa9", 0)`, "e9:2"},
		{"overlong, strict", `PerlAPI::decode_first("\xc0\xaf", 0)`, "malformed"},
		{"overlong, allowed", `PerlAPI::decode_first("\xc0\xaf", 1)`, "fffd:2"},
		{"overlong, with its value", `PerlAPI::decode_first("\xe0\x80\xaf", 2)`, "2f:3"},
	})
}

// PerlIO_binmode and PerlIO_apply_layers put layers on a handle of the
// program, and a handle operation that fails leaves its reason in errno
// for the native code that called it.
func TestNativeHandleLayersAndErrors(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "text.txt")
	if err := os.WriteFile(file, []byte("caf\xc3\xa9\r\nline\r\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	if r, err := p.Eval(context.Background(), `our $file = '`+file+`'; 1`); err != nil || r.Error != nil {
		t.Fatalf("setup: err=%v error=%v", err, r.Error)
	}
	runEvalCases(t, p, []evalCase{
		{"PerlIO_binmode with a layer", `open my $fh, "<", $file or die $!; my $r = PerlAPI::pio_layers($fh, ":utf8", 0); my $l = <$fh>; "$r:" . length($l)`, "1:6"},
		{"PerlIO_apply_layers", `open my $fh, "<", $file or die $!; my $r = PerlAPI::pio_layers($fh, ":crlf", 1); my $l = <$fh>; "$r:" . length($l)`, "0:6"},
		{"an unknown layer", `open my $fh, "<", $file or die $!; no warnings; PerlAPI::pio_layers($fh, ":no_such_layer", 1)`, "-1"},
		// The seek fails with what the program's own seek on a pipe fails with.
		{"errno after an open and a seek that fail", `pipe(my $r, my $w) or die $!; my @f = PerlAPI::pio_failures("$file.missing", $r); seek($r, 0, 0); my $own = $!{ESPIPE} ? "ESPIPE" : $!{EINVAL} ? "EINVAL" : "$!"; join ",", @f[0 .. 2], ($f[3] eq $own ? "as the program's" : "$f[3] not $own")`, "1,ENOENT,1,as the program's"},
		{"errno left alone by an open and a seek that work", `open my $fh, "<", $file or die $!; join ",", PerlAPI::pio_failures($file, $fh)`, "0,none,0,none"},
	})
}

// A module that has run the get magic of its argument itself (SvGETMAGIC)
// looks at the value without running it again: a tied scalar is fetched
// once, however often the module asks what it holds.
func TestNativeLooksRunNoGetMagic(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	if r, err := p.Eval(context.Background(), `
package Counted { sub TIESCALAR { bless { n => 0, v => $_[1] }, $_[0] } sub FETCH { $_[0]{n}++; $_[0]{v} } sub STORE { $_[0]{v} = $_[1] } }
sub fetches { my ($how, $value) = @_; my $obj = tie my $s, "Counted", $value; PerlAPI::look_without_magic($s, $how); $obj->{n} }
1`); err != nil || r.Error != nil {
		t.Fatalf("setup: err=%v error=%v", err, r.Error)
	}
	cases := []evalCase{{"SvGETMAGIC fetches once", `fetches("get", "text")`, "1"}}
	for _, how := range []string{"ok", "pv", "pv_nolen", "pvbyte", "iv", "uv", "nv", "true", "utf8",
		"tainted", "rok", "gmagical", "flags", "pok", "iok", "type", "cur", "copy"} {
		cases = append(cases, evalCase{how + " fetches nothing", `fetches("` + how + `", "12")`, "0"})
	}
	// These run the get magic themselves in perl: sv_isobject once, and
	// sv_derived_from as often as UNIVERSAL::isa does after its own fetch.
	cases = append(cases,
		evalCase{"sv_isobject fetches once", `fetches("isobject", "12")`, "1"},
		evalCase{"sv_derived_from fetches as perl's does",
			`my $obj = tie my $s, "Counted", "12"; UNIVERSAL::isa($s, "No::Such::Class"); fetches("derived", "12") - ($obj->{n} - 1)`, "0"})
	cases = append(cases, evalCase{"the whole of a module's reading: one fetch",
		`my $obj = tie my $s, "Counted", "text"; PerlAPI::look_without_magic($s, "get"); PerlAPI::look_without_magic($s, $_) for qw(ok pv utf8 tainted rok); $obj->{n}`, "1"})
	runEvalCases(t, p, cases)
}

// The interpreter calls every hook of the vtable of magic native code
// attached: get when the variable is read (the hook may change what is
// read), set when it is written, clear when an aggregate is emptied.
func TestNativeMagicVtableHooks(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"a scalar read and written", `my $v = "x"; PerlAPI::observe(\$v); my $a = $v; $v = "y"; my $b = "$v$v"; $v .= "z"; join ",", $a, $b, $v, PerlAPI::observed()`, "x,yy,yz,4,2,0"},
		{"the get hook gives the value", `my $v = "count"; PerlAPI::observe(\$v); PerlAPI::observed(); my $seen = $v; "$seen"`, "1"},
		{"a hash emptied", `my %h = (a => 1); PerlAPI::observe(\%h); PerlAPI::observed(); %h = (); my @c = PerlAPI::observed(); $c[2]`, "1"},
		{"a variable without the magic", `my $v = "x"; my $w = "x"; PerlAPI::observe(\$v); PerlAPI::observed(); $w = $w . "y"; join ",", PerlAPI::observed()`, "0,0,0"},
	})
}

// uvar magic on a hash: the module's function is called for every access
// with the key, and the key it leaves is the one that is used.
func TestNativeUvarMagic(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"the key is replaced", `my %h; PerlAPI::uvar_hash(\%h); $h{real} = "value"; my $got = $h{alias}; $h{alias} = "changed"; join ",", $got, $h{real}, join("|", sort keys %h), (PerlAPI::uvar_calls() >= 4 ? "called" : "not called")`, "value,changed,real,called"},
		{"another hash is left alone", `my %h; my %other = (alias => 1); PerlAPI::uvar_hash(\%h); PerlAPI::uvar_calls(); my $x = $other{alias}; PerlAPI::uvar_calls()`, "0"},
	})
}

// A croak in an svt_set hook of native code dies where the value is
// assigned, as in perl's mg_set: the value is assigned, and code after the
// assignment does not run.
func TestNativeMagicSetHookCroaks(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"refused", `my $v = "ok"; PerlAPI::refuse_set(\$v, "bad"); my $after = 0; my $r = eval { $v = "bad"; $after = 1; 1 }; join ",", $r ? "returned" : "died", $@ =~ /^refused: bad at / ? "message" : $@, $after, $v`, "died,message,0,bad"},
		{"accepted", `my $v = "ok"; PerlAPI::refuse_set(\$v, "bad"); $v = "good"; $v`, "good"},
	})
}

// uvar magic on a scalar: perl calls the ufuncs with their index when the
// variable is read and written; and a module that replaces the MAGIC's
// vtable after attaching it, as SWIG does for the variables of a wrapped
// library, gets its own svt_get and svt_set called, with the MAGIC as it
// attached it.
func TestNativeUvarMagicOnAScalar(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"the ufuncs", `my $v; PerlAPI::uvar_scalar(\$v, 100); PerlAPI::uvar_value(5); my $read = $v; $v = 142; join ",", $read, (PerlAPI::uvar_value())[0]`, "105,42"},
		{"a vtable of the module's", `PerlAPI::swig_variable("PerlAPI::answer"); PerlAPI::uvar_value(42); my $read = $PerlAPI::answer; $PerlAPI::answer = 7; join ",", $read, PerlAPI::uvar_value()`, "42,7,2"},
		{"a short name", `PerlAPI::swig_variable("main::x"); PerlAPI::uvar_value(3); my $read = $main::x; $main::x = 4; join ",", $read, PerlAPI::uvar_value()`, "3,4,2"},
	})
}

// A magic holds a reference on its mg_obj only when sv_magicext took one
// (MGf_REFCOUNTED), and that is the reference given up with the magic. A
// module that sets mg_obj itself after sv_magicext gives its reference up
// in its own svt_free: freeing the variable gives up nothing else, whether
// that magic is the variable's first or comes after another.
func TestNativeMagicHoldsOnlyTheReferenceItTook(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	const prelude = `my @w; local $SIG{__WARN__} = sub { push @w, @_ }; our ($pinned, $viewed) = ("p", "v" x 8); my ($p0, $v0) = (Internals::SvREFCNT($pinned), Internals::SvREFCNT($viewed)); `
	const tail = `join ",", Internals::SvREFCNT($pinned) - $p0, Internals::SvREFCNT($viewed) - $v0, scalar(@w), @w`
	runEvalCases(t, p, []evalCase{
		{"the flag", prelude + `my $h = "h"; PerlAPI::view_attach($h, $viewed); PerlAPI::pin_attach($h, $pinned); join ",", map { PerlAPI::magic_refcounted($h, $_) ? "counted" : "not counted" } qw(view pin)`, "counted,not counted"},
		{"set by the module, alone", prelude + `{ my $h = "h"; PerlAPI::pin_attach($h, $pinned); push @w, "held " . (Internals::SvREFCNT($pinned) - $p0) } ` + tail, "0,0,1,held 1"},
		{"set by the module, after a magic that took one", prelude + `{ my $h = "h"; PerlAPI::view_attach($h, $viewed); PerlAPI::pin_attach($h, $pinned) } ` + tail, "0,0,0"},
		{"taken by sv_magicext, after the module's", prelude + `{ my $h = "h"; PerlAPI::pin_attach($h, $pinned); PerlAPI::view_attach($h, $viewed) } ` + tail, "0,0,0"},
	})
}
